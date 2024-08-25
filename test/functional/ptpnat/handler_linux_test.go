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

package ptpnat_test

import (
	"errors"
	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/spacechunks/platform/internal/ptpnat"
	ptptesting "github.com/spacechunks/platform/test/functional/ptpnat"
	"github.com/stretchr/testify/require"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"net"
	"testing"
)

// we use github.com/vishvananda/netns library and
// github.com/containernetworking/plugins/pkg/ns
// because
// * github.com/vishvananda/netns
//   provides us with the ability to create/destroy named network namespaces.
//   the other one does not provide this feature.
// * github.com/containernetworking/plugins/pkg/ns
//   provides us with the ability to execute functions in the context of
//   a given network namespace.

var stdinData = []byte(`
{
  "cniVersion": "1.0.0",
  "name":"test",
  "ipam":{
    "type": "host-local", 
    "ranges":[
      [{"subnet": "10.0.10.0/24"}]
    ]
  }
}
`)

// TODO: remove below code and instead write functional test for cni.go

// TestSetup tests that ip addresses cloud be allocated
// and configured on the veth-pairs.
func TestSetup(t *testing.T) {
	var (
		created, origin, name = ptptesting.CreateNetns(t)
		h                     = ptpnat.NewHandler()
		ctrID                 = "ABC"
		nsPath                = "/var/run/netns/" + name
	)

	defer func() {
		created.Close()
		origin.Close()
		netns.DeleteNamed(name)
		h.DeallocIPs("host-local", stdinData)
	}()

	// host-local cni plugin requires container id
	ptptesting.SetCNIEnvVars(ctrID, "ignored", "ignored")

	ips, err := h.AllocIPs("host-local", stdinData)
	require.NoError(t, err)

	hostVethName, podVethName, err := h.CreateAndConfigureVethPair(nsPath, ips)
	require.NoError(t, err)

	var (
		hostVeth = ptptesting.GetLinkByNS(t, hostVethName, origin)
		podVeth  = ptptesting.GetLinkByNS(t, podVethName, created)
	)

	require.NotNil(t, podVeth, "pod veth not found")
	require.NotNil(t, hostVeth, "host veth not found")
	require.Equal(t, ptpnat.VethMTU, podVeth.Attrs().MTU)

	err = ns.WithNetNSPath(nsPath, func(netNS ns.NetNS) error {
		ptptesting.RequireAddrConfigured(t, podVethName, ptpnat.ContainerIP4)
		return nil
	})
	require.NoError(t, err)

	require.NoError(t, netns.Set(origin))
	ptptesting.RequireAddrConfigured(t, hostVethName, ips[0].Address.String())
}

func TestBPFAttach(t *testing.T) {
	tests := []struct {
		name               string
		pinPrefix          string
		expectedAttachType uint32
		attach             func(*testing.T, ptpnat.Handler, string)
	}{
		{
			name:               "attach dnat",
			pinPrefix:          "dnat_",
			expectedAttachType: 46, // BPF_TCX_INGRESS, see github.com/cilium/ebpf/internal/sys/types.go
			attach: func(t *testing.T, h ptpnat.Handler, ifaceName string) {
				require.NoError(t, h.AttachDNATBPF(ifaceName))
			},
		},
		{
			name:               "attach snat",
			pinPrefix:          "snat_",
			expectedAttachType: 46,
			attach: func(t *testing.T, h ptpnat.Handler, ifaceName string) {
				require.NoError(t, h.AttachSNATBPF(ifaceName))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ifaceName, veth := ptptesting.AddRandVethPair(t)
			defer netlink.LinkDel(veth)

			h := ptpnat.NewHandler()
			tt.attach(t, h, ifaceName)

			l, err := link.LoadPinnedLink("/sys/fs/bpf/ptp_"+tt.pinPrefix+ifaceName, &ebpf.LoadPinOptions{})
			require.NoError(t, err)

			defer l.Unpin()

			info, err := l.Info()
			require.NoError(t, err)

			require.Equal(t, tt.expectedAttachType, uint32(info.TCX().AttachType))
		})
	}
}

func TestConfigureSNAT(t *testing.T) {
	tests := []struct {
		name string
		prep func(*testing.T, netlink.Link)
		err  error
	}{
		{
			name: "works",
			prep: func(t *testing.T, veth netlink.Link) {
				require.NoError(t, netlink.AddrAdd(veth, &netlink.Addr{
					IPNet: &net.IPNet{
						IP:   net.ParseIP("10.0.0.1"),
						Mask: []byte{255, 255, 255, 0},
					},
				}))
			},
		},
		{
			name: "no addresses configured",
			prep: func(t *testing.T, veth netlink.Link) {},
			err:  errors.New("no addresses configured"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				h               = ptpnat.NewHandler()
				ifaceName, veth = ptptesting.AddRandVethPair(t)
			)

			tt.prep(t, veth)
			defer netlink.LinkDel(veth)

			iface, err := net.InterfaceByName(ifaceName)
			require.NoError(t, err)

			if tt.err != nil {
				require.EqualError(t, h.ConfigureSNAT(iface), tt.err.Error())
				return
			}

			require.NoError(t, h.ConfigureSNAT(iface))

			conf, err := ebpf.LoadPinnedMap("/sys/fs/bpf/ptp_snat_config", nil)
			require.NoError(t, err)

			defer conf.Unpin()

			type ptpSnatEntry struct {
				IpAddr   uint32
				IfaceIdx uint8
				_        [3]byte
			}

			expected := ptpSnatEntry{
				IpAddr:   16777226, // 10.0.0.1 in little endian decimal
				IfaceIdx: 3,
			}

			var actual ptpSnatEntry
			require.NoError(t, conf.Lookup(uint8(0), &actual))

			require.Equal(t, expected, actual)
		})
	}
}
