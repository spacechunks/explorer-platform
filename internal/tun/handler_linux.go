/*
Explorer Platform, a platform for hosting and discovering Minecraft servers.
Copyright (C) 2024 Yannic Rieger <oss@76k.io>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package tun

import (
	"crypto/rand"
	"errors"
	"fmt"
	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	current "github.com/containernetworking/cni/pkg/types/100"
	"github.com/containernetworking/plugins/pkg/ipam"
	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/vishvananda/netlink"
	"net"
)

// just for reference: _ctr_ is short for _container_

const (
	VethMTU      = 1400
	ContainerIP4 = "10.0.0.1/24"
)

type Handler interface {
	CreateAndConfigureVethPair(netNS string, ips []*current.IPConfig) (string, string, error)
	AttachSNATBPF(ifaceName string) error
	// TODO: check if we really need to request an ip address
	//       from ipam plugin, because our ingress bpf program will
	//       take care of redirecting the packet to the correct iface.
	//       so in theory all host-side interfaces could have the same
	//       ip configured (tm).
	AllocIPs(plugin string, stdinData []byte) ([]*current.IPConfig, error)
	DeallocIPs(plugin string, stdinData []byte) error
	AttachDNATBPF(ifaceName string) error
	ConfigureSNAT(mapPin string) error
}

type cniHandler struct {
}

func NewHandler() Handler {
	return &cniHandler{}
}

func (h *cniHandler) AttachSNATBPF(ifaceName string) error {
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		return fmt.Errorf("get iface: %w", err)
	}

	var snatObjs snatObjects
	if err := loadSnatObjects(&snatObjs, nil); err != nil {
		return fmt.Errorf("load snat objs: %w", err)
	}

	l, err := link.AttachTCX(link.TCXOptions{
		Interface: iface.Index,
		Program:   snatObjs.Snat,
		Attach:    ebpf.AttachTCXIngress,
	})
	if err != nil {
		return fmt.Errorf("attach snat: %w", err)
	}

	// pin because cni is short-lived
	if err := l.Pin(fmt.Sprintf("/sys/fs/bpf/snat_%s", ifaceName)); err != nil {
		return fmt.Errorf("pin link: %w", err)
	}

	return nil
}

func (h *cniHandler) CreateAndConfigureVethPair(netNS string, ips []*current.IPConfig) (string, string, error) {
	hostVethName, err := randHexStr()
	if err != nil {
		return "", "", fmt.Errorf("could not generate host-side veth name: %w", err)
	}

	podVethName, err := randHexStr()
	if err != nil {
		return "", "", fmt.Errorf("could not generate pod-side veth name: %w", err)
	}

	ctrNS, err := createAndMoveVethPair(hostVethName, podVethName, netNS)
	if err != nil {
		return "", "", fmt.Errorf("setup veth pair: %w", err)
	}

	defer ctrNS.Close()

	if err := configureCTRIface(ctrNS, podVethName); err != nil {
		return "", "", fmt.Errorf("setup ctr side veth: %w", err)
	}

	if err := configureHostIface(ips, hostVethName); err != nil {
		return "", "", fmt.Errorf("setup host side veth: %w", err)
	}

	return hostVethName, podVethName, nil
}

func (h *cniHandler) AllocIPs(plugin string, stdinData []byte) ([]*current.IPConfig, error) {
	ipamRes, err := ipam.ExecAdd(plugin, stdinData)
	if err != nil {
		return nil, fmt.Errorf("ipam: %v", err)
	}

	// convert ipam result into the current versions result type
	result, err := current.NewResultFromResult(ipamRes)
	if err != nil {
		return nil, fmt.Errorf("convert ipam result: %v", err)
	}

	if len(result.IPs) == 0 {
		return nil, errors.New("ipam plugin returned missing IPs")
	}

	return result.IPs, nil
}

func (h *cniHandler) DeallocIPs(plugin string, stdinData []byte) error {
	return ipam.ExecDel(plugin, stdinData)
}

func (h *cniHandler) AttachDNATBPF(ifaceName string) error {
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		return fmt.Errorf("get iface: %w", err)
	}

	var dnatObjs dnatObjects
	if err := loadDnatObjects(&dnatObjs, nil); err != nil {
		return fmt.Errorf("load dnat objs: %w", err)
	}

	// TODO: if link is already present just update
	l, err := link.AttachTCX(link.TCXOptions{
		Interface: iface.Index,
		Program:   dnatObjs.Dnat,
		Attach:    ebpf.AttachTCXIngress,
	})
	if err != nil {
		return fmt.Errorf("attach dnat: %w", err)
	}

	// pin because cni is short-lived
	if err := l.Pin(fmt.Sprintf("/sys/fs/bpf/dnat_%s", ifaceName)); err != nil {
		return fmt.Errorf("pin link: %w", err)
	}

	return nil
}

func (h *cniHandler) ConfigureSNAT(mapPin string) error {
	// TODO: put snat config into ebpf map
	//       if map is not present -> add and pin
	//       if map is present -> reuse
	return nil
}

func configureCTRIface(ctrNS ns.NetNS, ifaceName string) error {
	if err := ctrNS.Do(func(ns.NetNS) error {
		ip, ipNet, _ := net.ParseCIDR(ContainerIP4)
		// for some reason the host part is lost
		// in ipNet. 10.0.0.1/24 -> 10.0.0.0/24
		return configureIface(ifaceName, &net.IPNet{
			IP:   ip,
			Mask: ipNet.Mask,
		})
	}); err != nil {
		return fmt.Errorf("ctr ns: %w", err)
	}
	return nil
}

func configureHostIface(ips []*current.IPConfig, ifaceName string) error {
	for _, ip := range ips {
		if err := configureIface(ifaceName, &ip.Address); err != nil {
			return fmt.Errorf("configure iface (%s): %w", ip.String(), err)
		}
	}
	return nil
}

func configureIface(ifaceName string, ipNet *net.IPNet) error {
	l, err := netlink.LinkByName(ifaceName)
	if err != nil {
		return fmt.Errorf("lookup link: %w", err)
	}

	if err := netlink.AddrAdd(l, &netlink.Addr{IPNet: ipNet}); err != nil {
		return fmt.Errorf("add addr: %w", err)
	}

	if err := netlink.LinkSetUp(l); err != nil {
		return fmt.Errorf("link up: %w", err)
	}

	return nil
}

func createAndMoveVethPair(hostVethName, podVethName, netNS string) (ns.NetNS, error) {
	vethpair := &netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{
			Name: podVethName,
			MTU:  VethMTU,
		},
		PeerName: hostVethName,
	}

	if err := netlink.LinkAdd(vethpair); err != nil {
		return nil, fmt.Errorf("add veth: %w", err)
	}

	ctr, err := ns.GetNS(netNS)
	if err != nil {
		return nil, fmt.Errorf("get netns fd: %w", err)
	}

	if err := netlink.LinkSetNsFd(vethpair, int(ctr.Fd())); err != nil {
		return nil, fmt.Errorf("move pod veth to ns %d: %w", ctr, err)
	}

	return ctr, nil
}

func randHexStr() (string, error) {
	bytes := make([]byte, 16) // are enough to achieve a negligible collision chance
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	// return first 15 chars
	return fmt.Sprintf("%x", bytes)[:15], nil
}
