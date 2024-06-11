package tun

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	current "github.com/containernetworking/cni/pkg/types/100"
	"github.com/containernetworking/plugins/pkg/ipam"
	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/vishvananda/netlink"
)

const (
	mtu    = 1400
	ctrIP4 = "10.0.0.1/24"
)

// Conf currently functions only as a wrapper struct for types.NetConf.
type Conf struct {
	types.NetConf
}

func CNIFuncs() skel.CNIFuncs {
	return skel.CNIFuncs{
		Add:    execAdd,
		Del:    execDel,
		Check:  nil,
		GC:     nil,
		Status: nil,
	}
}

// execAdd sets up the veth pair for a container.
// internally the following happens:
// * first allocated ip address for host side veth  using cni ipam plugin.
// * then create veth pair and move one peer into the containers netns.
// * configure ip address on container iface and bring it up.
// * configure ip address on host iface and bring it up.
func execAdd(args *skel.CmdArgs) (err error) {
	var conf Conf
	if err := json.Unmarshal(args.StdinData, &conf); err != nil {
		return fmt.Errorf("parse network config: %v", err)
	}
	ipamRes, err := ipam.ExecAdd(conf.IPAM.Type, args.StdinData)
	if err != nil {
		return fmt.Errorf("ipam: %v", err)
	}
	defer func() {
		if err != nil {
			if err := ipam.ExecDel(conf.IPAM.Type, args.StdinData); err != nil {
				log.Print("could not delete allocated ips after CNI ADD failure.")
			}
		}
	}()
	// convert ipam result into the current versions result type
	result, err := current.NewResultFromResult(ipamRes)
	if err != nil {
		return fmt.Errorf("convert ipam result: %v", err)
	}
	if len(result.IPs) == 0 {
		return errors.New("ipam plugin returned missing IPs")
	}
	var (
		hostVethName = fmt.Sprintf("veth%s", shorten(args.ContainerID))
		podVethName  = fmt.Sprintf("veth%s", shorten(args.ContainerID))
	)
	ctrNS, err := createAndMoveVethPair(hostVethName, podVethName, args)
	if err != nil {
		return fmt.Errorf("setup veth pair: %v", err)
	}
	defer ctrNS.Close()
	if err := configureCTRIface(ctrNS, podVethName); err != nil {
		return fmt.Errorf("setup ctr side veth: %v", err)
	}
	if err := configureHostIface(result.IPs, hostVethName); err != nil {
		return fmt.Errorf("setup host side veth: %v", err)
	}
	// TODO: attach ebpf progs
	return nil
}

func execDel(args *skel.CmdArgs) error {
	// TODO: remove veth pairs
	return nil
}

func configureCTRIface(ctrNS ns.NetNS, ifaceName string) error {
	if err := ctrNS.Do(func(ns.NetNS) error {
		_, ipNet, _ := net.ParseCIDR(ctrIP4)
		return configureIface(ifaceName, ipNet)
	}); err != nil {
		return fmt.Errorf("ctr ns: %v", err)
	}
	return nil
}

func configureHostIface(ips []*current.IPConfig, ifaceName string) error {
	for _, ip := range ips {
		if err := configureIface(ifaceName, &ip.Address); err != nil {
			return fmt.Errorf("configure iface (%s): %v", ip.String(), err)
		}
	}
	return nil
}

func configureIface(ifaceName string, ipNet *net.IPNet) error {
	l, err := netlink.LinkByName(ifaceName)
	if err != nil {
		return fmt.Errorf("lookup link: %v", err)
	}
	if err := netlink.AddrAdd(l, &netlink.Addr{IPNet: ipNet}); err != nil {
		return fmt.Errorf("add addr: %v", err)
	}
	if err := netlink.LinkSetUp(l); err != nil {
		return fmt.Errorf("link up: %v", err)
	}
	return nil
}

func createAndMoveVethPair(hostVethName, podVethName string, args *skel.CmdArgs) (ns.NetNS, error) {
	vethpair := &netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{
			Name: podVethName,
			MTU:  mtu,
		},
		PeerName: hostVethName,
	}
	if err := netlink.LinkAdd(vethpair); err != nil {
		return nil, fmt.Errorf("add veth: %v", err)
	}
	ctr, err := ns.GetNS(args.Netns)
	if err != nil {
		return nil, fmt.Errorf("get netns fd: %v", err)
	}
	if err := netlink.LinkSetNsFd(vethpair, int(ctr.Fd())); err != nil {
		return nil, fmt.Errorf("move pod veth to ns %d: %v", ctr, err)
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
