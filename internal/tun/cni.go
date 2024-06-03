package tun

import (
	"fmt"
	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

const mtu = 1400

type Conf struct {
	types.NetConf
}

func CNIFuncs() skel.CNIFuncs {
	return skel.CNIFuncs{
		Add:    cniAdd,
		Del:    cniDel,
		Check:  nil,
		GC:     nil,
		Status: nil,
	}
}

func cniAdd(args *skel.CmdArgs) error {
	//var conf Conf
	/*if err := json.Unmarshal(args.StdinData, &conf); err != nil {
		return fmt.Errorf("add: parse network config: %v", err)
	}*/
	var (
		hostVethName = fmt.Sprintf("host%s", args.ContainerID)
		podVethName  = fmt.Sprintf("pod%s", args.ContainerID)
	)
	podVeth := &netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{
			Name: podVethName,
			MTU:  mtu,
		},
		PeerName: hostVethName,
	}
	if err := netlink.LinkAdd(podVeth); err != nil {
		return fmt.Errorf("add pod veth: %v", err)
	}
	handle, err := netns.GetFromPath(args.Netns)
	if err != nil {
		return fmt.Errorf("get netns fd: %v", err)
	}
	if err := netlink.LinkSetNsFd(podVeth, int(handle)); err != nil {
		return fmt.Errorf("move pod veth to ns %d: %v", int(handle), err)
	}
	// TODO: attach ebpf progs
	return nil
}

func cniDel(args *skel.CmdArgs) error {
	return nil
}
