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

package ptpnat

import (
	"encoding/json"
	"fmt"
	current "github.com/containernetworking/cni/pkg/types/100"
	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/spacechunks/platform/internal/ptpnat/gobpf"
	"log"
	"net"
	"net/netip"
	"os"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
)

type Conf struct {
	types.NetConf
	HostIface string `json:"hostIface"`
}

type CNI struct {
	handler Handler
}

func NewCNI(h Handler) *CNI {
	return &CNI{
		handler: h,
	}
}

// ExecAdd sets up the veth pair for a container.
// internally the following happens:
// * first allocated ip address for host side veth using cni ipam plugin.
// * then create veth pair and move one peer into the containers netns.
// * configure ip address on container iface and bring it up.
// * configure ip address on host iface and bring it up.
// * attach snat bpf program to host-side veth peer (tc ingress)
func (c *CNI) ExecAdd(args *skel.CmdArgs) (err error) {
	var conf Conf
	if err := json.Unmarshal(args.StdinData, &conf); err != nil {
		return fmt.Errorf("parse network config: %v", err)
	}

	if conf.IPAM == (types.IPAM{}) {
		return fmt.Errorf("no IPAM configuration")
	}

	if conf.HostIface == "" {
		return fmt.Errorf("hostIface config value not set")
	}

	if err := c.handler.AttachDNATBPF(conf.HostIface); err != nil {
		return fmt.Errorf("failed to attach dnat bpf to %s: %w", conf.HostIface, err)
	}

	defer func() {
		if err != nil {
			if err := c.handler.DeallocIPs(conf.IPAM.Type, args.StdinData); err != nil {
				log.Printf("could not deallocate ips after CNI ADD failure: %v\n", err)
			}
		}
	}()

	ips, err := c.handler.AllocIPs(conf.IPAM.Type, args.StdinData)
	if err != nil {
		return fmt.Errorf("alloc ips: %w", err)
	}

	hostVethName, podVethName, err := c.handler.CreateAndConfigureVethPair(args.Netns, ips)
	if err != nil {
		return fmt.Errorf("configure veth pair: %w", err)
	}

	if err := c.handler.AttachHostVethBPF(hostVethName); err != nil {
		return fmt.Errorf("attach snat: %w", err)
	}

	if err := c.handler.ConfigureSNAT(conf.HostIface); err != nil {
		return fmt.Errorf("configure snat: %w", err)
	}

	if err := c.handler.AddDefaultRoute(args.Netns); err != nil {
		return fmt.Errorf("add default route: %w", err)
	}

	// TEST START
	var hwAddr net.HardwareAddr
	ns.WithNetNSPath(args.Netns, func(_ ns.NetNS) error {
		iface, err := net.InterfaceByName(podVethName)
		if err != nil {
			return err
		}
		hwAddr = iface.HardwareAddr
		return nil
	})
	if err != nil {
		return err
	}
	h, err := net.InterfaceByName(hostVethName)
	if err != nil {
		return fmt.Errorf("get host veth name: %w", err)
	}
	if err := gobpf.AddDNATTarget(
		uint16(80), netip.MustParseAddr("10.0.0.1"), uint8(h.Index), hwAddr, pinPath,
	); err != nil {
		return fmt.Errorf("add dnat target: %w", err)
	}
	// TEST END

	result := &current.Result{
		CNIVersion: supportedCNIVersion,
		Interfaces: []*current.Interface{
			{
				Name:    podVethName,
				Sandbox: args.Netns,
			},
		},
	}

	if err := result.PrintTo(os.Stdout); err != nil {
		return fmt.Errorf("print result: %w", err)
	}

	return nil
}

func (c *CNI) ExecDel(args *skel.CmdArgs) error {
	log.Println("del")
	// TODO: remove veth pairs
	return nil
}
