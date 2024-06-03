package tun

import (
	"encoding/json"
	"fmt"
	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/vishvananda/netlink"
	"strconv"
	"strings"
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
	var conf Conf
	if err := json.Unmarshal(args.StdinData, &conf); err != nil {
		return fmt.Errorf("add: parse network config: %v", err)
	}
	var (
		hostVethName = fmt.Sprintf("host%s", args.ContainerID)
		podVethName  = fmt.Sprintf("pod%s", args.ContainerID)
	)
	// TODO: create in pod netns
	podVeth := &netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{
			Name: podVethName,
			MTU:  mtu,
		},
		PeerName: hostVethName,
	}
	hostVeth := &netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{
			Name: hostVethName,
			MTU:  mtu,
		},
		PeerName: podVethName,
	}
	if err := netlink.LinkAdd(podVeth); err != nil {
		return fmt.Errorf("add pod veth: %v", err)
	}
	if err := netlink.LinkAdd(hostVeth); err != nil {
		return fmt.Errorf("add host veth: %v", err)
	}
	// TODO: attach ebpf progs
	return nil
}

func cniDel(args *skel.CmdArgs) error {
	return nil
}

func parseNetnsID(netnsPath string) (int, error) {
	// /run/netns/[nsname]
	parts := strings.Split(netnsPath, "/")
	id, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		return -1, fmt.Errorf("parse netns id: %v", err)
	}
	return id, nil
}
