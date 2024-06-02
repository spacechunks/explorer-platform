package tun

import (
	"encoding/json"
	"fmt"
	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
)

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
	return nil
}

func cniDel(args *skel.CmdArgs) error {
	return nil
}
