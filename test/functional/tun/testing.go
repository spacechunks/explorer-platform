package tun

import (
	"crypto/rand"
	"fmt"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"net"
	"os"
	"runtime"
	"testing"
)

// SetCNIEnvVars sets environment variables needed in order
// to make cni IPAM plugin work.
func SetCNIEnvVars(containerID, ifname, cniNetNS string) {
	_ = os.Setenv("CNI_CONTAINERID", containerID)
	_ = os.Setenv("CNI_IFNAME", ifname)
	_ = os.Setenv("CNI_NETNS", cniNetNS)
}

func CreateNetns(t *testing.T) (netns.NsHandle, netns.NsHandle, string) {
	// lock the OS Thread, so we don't accidentally switch namespaces
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	// generate random netns name to avoid collisions
	// when running multiple tests at once.
	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		t.Fatalf("failed reading random bytes: %v", err)
	}
	name := fmt.Sprintf("%x", bytes)
	origin, err := netns.Get()
	if err != nil {
		t.Fatalf("create netns: %v", err)
	}
	handle, err := netns.NewNamed(name)
	if err != nil {
		t.Fatalf("create netns: %v", err)
	}
	if err := netns.Set(origin); err != nil {
		t.Fatalf("set netns: %v", err)
	}
	return handle, origin, name
}

func GetLinkByNS(t *testing.T, name string, h netns.NsHandle) netlink.Link {
	if err := netns.Set(h); err != nil {
		t.Fatalf("switch netns (%d): %v", int(h), err)
	}
	l, err := netlink.LinkByName(name)
	if err != nil {
		t.Fatalf("get link by name (%s): %v", name, err)
	}
	return l
}

func RequireAddrConfigured(t *testing.T, ifaceName, expectedAddr string) {
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		t.Fatalf("get iface by name (%s): %v", ifaceName, err)
	}
	addrs, err := iface.Addrs()
	if err != nil {
		t.Fatalf("get addrs of iface (%s): %v", ifaceName, err)
	}
	for _, addr := range addrs {
		if addr.String() == expectedAddr {
			return
		}
	}
	t.Fatalf("expected %s to be configured", expectedAddr)
}
