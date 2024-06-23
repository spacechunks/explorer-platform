package tun_test

import (
	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/spacechunks/platform/internal/tun"
	tuntesting "github.com/spacechunks/platform/test/functional/tun"
	"github.com/stretchr/testify/assert"
	"github.com/vishvananda/netns"
	"log"
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
	tuntesting.SetCNIEnvVars(ctrID, "ignored", "ignored")
	ips, err := h.AllocIPs("host-local", stdinData)
	if err != nil {
		t.Fatalf("alloc ips: %v", err)
	}
	hostVethName, podVethName, err := h.CreateAndConfigureVethPair(nsPath, ips)
	if err != nil {
		t.Fatalf("create and configure failed: %v", err)
	}
	var (
		hostVeth = tuntesting.GetLinkByNS(t, hostVethName, origin)
		podVeth  = tuntesting.GetLinkByNS(t, podVethName, created)
	)
	assert.NotNil(t, podVeth, "pod veth not found")
	assert.NotNil(t, hostVeth, "host veth not found")
	assert.Equal(t, tun.VethMTU, podVeth.Attrs().MTU)
	if err := ns.WithNetNSPath(nsPath, func(netNS ns.NetNS) error {
		tuntesting.AssertAddrConfigured(t, podVethName, tun.ContainerIP4)
		return nil
	}); err != nil {
		log.Fatalf("check pod ns: %v", err)
	}
	if err := netns.Set(origin); err != nil {
		log.Fatalf("switch back: %v", err)
	}
	tuntesting.AssertAddrConfigured(t, hostVethName, ips[0].Address.String())
}
