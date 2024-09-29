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
	"log"
	"net"
	"os"
	"runtime"
	"testing"

	"github.com/spacechunks/platform/test"
	"github.com/stretchr/testify/require"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
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
	name := test.RandHexStr(t)
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

// AddRandVethPair adds a veth pair with a random name.
// This is mostly used for tests where a dummy network
// interface is needed.
func AddRandVethPair(t *testing.T) (string, netlink.Link) {
	var (
		ifaceName = test.RandHexStr(t)
		vethpair  = &netlink.Veth{
			LinkAttrs: netlink.LinkAttrs{
				Name: ifaceName,
			},
			PeerName: ifaceName + "-p",
		}
	)
	log.Println(ifaceName)
	require.NoError(t, netlink.LinkAdd(vethpair))
	return ifaceName, vethpair
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

func RequireMACConfigured(t *testing.T, ifaceName, expectedMAC string) {
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		t.Fatalf("get iface by name (%s): %v", ifaceName, err)
	}
	if iface.HardwareAddr.String() != expectedMAC {
		t.Fatalf("expected %s got %s", expectedMAC, iface.HardwareAddr.String())
	}
}
