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

const (
	mtu    = 1400
	ctrIP4 = "10.0.0.1/24"
)

type Handler interface {
	CreateAndConfigureVethPair(ctrID, netNS string, ips []*current.IPConfig) (string, error)
	AttachEgressBPF(ifaceName string) error
	// TODO: check if we really need to request an ip address
	//       from ipam plugin, because our ingress bpf program will
	//       take care of redirecting the packet to the correct iface.
	//       so in theory all host-side interfaces could have the same
	//       ip configured (tm).
	AllocIPs(plugin string, stdinData []byte) ([]*current.IPConfig, error)
	DeallocIPs(plugin string, stdinData []byte) error
}

type cniHandler struct {
}

func NewHandler() Handler {
	return &cniHandler{}
}

func (h *cniHandler) AttachEgressBPF(ifaceName string) error {
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		return fmt.Errorf("get iface: %w", err)
	}
	var ingObjs ingressObjects
	if err := loadIngressObjects(&ingObjs, nil); err != nil {
		return fmt.Errorf("load ingress objs: %w", err)
	}
	l, err := link.AttachTCX(link.TCXOptions{
		Interface: iface.Index,
		Program:   ingObjs.Ingress,
		Attach:    ebpf.AttachTCXIngress,
	})
	if err != nil {
		return fmt.Errorf("attach ingress: %w", err)
	}
	if err := l.Pin(fmt.Sprintf("ingress_%s", ifaceName)); err != nil {
		return fmt.Errorf("pin link: %w", err)
	}
	return nil
}

func (h *cniHandler) CreateAndConfigureVethPair(ctrID, netNS string, ips []*current.IPConfig) (string, error) {
	var (
		hostVethName = shorten(ctrID)
		podVethName  = shorten(ctrID)
	)
	ctrNS, err := createAndMoveVethPair(hostVethName, podVethName, netNS)
	if err != nil {
		return "", fmt.Errorf("setup veth pair: %w", err)
	}
	defer ctrNS.Close()
	if err := configureCTRIface(ctrNS, podVethName); err != nil {
		return "", fmt.Errorf("setup ctr side veth: %w", err)
	}
	if err := configureHostIface(ips, hostVethName); err != nil {
		return "", fmt.Errorf("setup host side veth: %w", err)
	}
	return podVethName, nil
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

func configureCTRIface(ctrNS ns.NetNS, ifaceName string) error {
	if err := ctrNS.Do(func(ns.NetNS) error {
		_, ipNet, _ := net.ParseCIDR(ctrIP4)
		return configureIface(ifaceName, ipNet)
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

func createAndMoveVethPair(hostVethName, podVethName string, netNS string) (ns.NetNS, error) {
	vethpair := &netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{
			Name: podVethName,
			MTU:  mtu,
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

// shorten shortens string to max 15 chars
func shorten(str string) string {
	if len(str) <= 15 {
		return str
	}
	return str[:15]
}
