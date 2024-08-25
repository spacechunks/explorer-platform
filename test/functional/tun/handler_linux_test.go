package tun_test

import (
	"fmt"
	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/spacechunks/platform/internal/tun"
	tuntesting "github.com/spacechunks/platform/test/functional/tun"
	"github.com/stretchr/testify/require"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
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

// TestSetup tests that ip addresses cloud be allocated
// and configured on the veth-pairs.
func TestSetup(t *testing.T) {
	var (
		created, origin, name = tuntesting.CreateNetns(t)
		h                     = tun.NewHandler()
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
	tuntesting.SetCNIEnvVars(ctrID, "ignored", "ignored")

	ips, err := h.AllocIPs("host-local", stdinData)
	require.NoError(t, err)

	hostVethName, podVethName, err := h.CreateAndConfigureVethPair(nsPath, ips)
	require.NoError(t, err)

	var (
		hostVeth = tuntesting.GetLinkByNS(t, hostVethName, origin)
		podVeth  = tuntesting.GetLinkByNS(t, podVethName, created)
	)

	require.NotNil(t, podVeth, "pod veth not found")
	require.NotNil(t, hostVeth, "host veth not found")
	require.Equal(t, tun.VethMTU, podVeth.Attrs().MTU)

	err = ns.WithNetNSPath(nsPath, func(netNS ns.NetNS) error {
		tuntesting.RequireAddrConfigured(t, podVethName, tun.ContainerIP4)
		return nil
	})
	require.NoError(t, err)

	require.NoError(t, netns.Set(origin))
	tuntesting.RequireAddrConfigured(t, hostVethName, ips[0].Address.String())
}

func TestBPFAttach(t *testing.T) {
	tests := []struct {
		name               string
		pinPrefix          string
		expectedAttachType uint32
		attach             func(*testing.T, tun.Handler, string)
	}{
		{
			name:               "attach dnat",
			pinPrefix:          "dnat_",
			expectedAttachType: 46, // BPF_TCX_INGRESS, see github.com/cilium/ebpf/internal/sys/types.go
			attach: func(t *testing.T, h tun.Handler, ifaceName string) {
				require.NoError(t, h.AttachDNATBPF(ifaceName))
			},
		},
		{
			name:               "attach snat",
			pinPrefix:          "snat_",
			expectedAttachType: 46,
			attach: func(t *testing.T, h tun.Handler, ifaceName string) {
				require.NoError(t, h.AttachSNATBPF(ifaceName))
			},
		},
	}

	// use for different iface names
	count := 0

	for _, tt := range tests {
		count++
		t.Run(tt.name, func(t *testing.T) {
			var (
				ifaceName = fmt.Sprintf("functestveth%d", count)
				vethpair  = &netlink.Veth{
					LinkAttrs: netlink.LinkAttrs{
						Name: ifaceName,
					},
					PeerName: ifaceName + "-p",
				}
			)

			require.NoError(t, netlink.LinkAdd(vethpair))
			defer netlink.LinkDel(vethpair)

			h := tun.NewHandler()
			tt.attach(t, h, ifaceName)

			l, err := link.LoadPinnedLink("/sys/fs/bpf/"+tt.pinPrefix+ifaceName, &ebpf.LoadPinOptions{})
			require.NoError(t, err)

			defer l.Unpin()

			info, err := l.Info()
			require.NoError(t, err)

			require.Equal(t, tt.expectedAttachType, uint32(info.TCX().AttachType))
		})
	}
}
