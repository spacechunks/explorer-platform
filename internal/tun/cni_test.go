//go:build linux

package tun_test

import (
	"github.com/containernetworking/cni/pkg/skel"
	"github.com/spacechunks/platform/internal/tun"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"runtime"
	"testing"
)

func createNetns(t *testing.T) (netns.NsHandle, netns.NsHandle, string) {
	// lock the OS Thread, so we don't accidentally switch namespaces
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	origin, err := netns.Get()
	if err != nil {
		t.Fatalf("create netns: %v", err)
	}
	handle, err := netns.NewNamed("test")
	if err != nil {
		t.Fatalf("create netns: %v", err)
	}
	if err := netns.Set(origin); err != nil {
		t.Fatalf("set netns: %v", err)
	}
	return handle, origin, "test"
}

func TestAdd(t *testing.T) {
	created, origin, name := createNetns(t)
	defer created.Close()
	defer origin.Close()
	defer netns.DeleteNamed(name)
	args := &skel.CmdArgs{
		ContainerID: "ABC",
		Netns:       "/var/run/netns/" + name,
	}
	if err := tun.CNIFuncs().Add(args); err != nil {
		t.Fatalf("add: %v", err)
	}
	podVeth := getLinkByNS(t, "podABC", created)
	hostVeth := getLinkByNS(t, "hostABC", origin)
	if podVeth == nil {
		t.Fatal("pod veth not found")
	}
	if hostVeth == nil {
		t.Fatal("host veth not found")
	}
}

func getLinkByNS(t *testing.T, name string, h netns.NsHandle) netlink.Link {
	if err := netns.Set(h); err != nil {
		t.Fatalf("switch netns (%d): %v", int(h), err)
	}
	l, err := netlink.LinkByName(name)
	if err != nil {
		t.Fatalf("get link by name (%s): %v", name, err)
	}
	return l
}
