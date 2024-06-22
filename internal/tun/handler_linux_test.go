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

package tun

import (
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"runtime"
	"testing"
)

// we use github.com/vishvananda/netns library for testing instead of
// github.com/containernetworking/plugins/pkg/ns, because it provides
// us with the ability to create/destroy named network namespaces.
// the other one does not provide this feature.

func TestCreateAndConfigureVethPair(t *testing.T) {
	created, origin, name := createNetns(t)
	defer created.Close()
	defer origin.Close()
	defer netns.DeleteNamed(name)
	h := cniHandler{}
	if _, err := h.CreateAndConfigureVethPair("ABC", "/var/run/netns/"+name, nil); err != nil {
		t.Fatalf("create and configure failed: %v", err)
	}
	podVeth := getLinkByNS(t, "ABC", created)
	hostVeth := getLinkByNS(t, "ABC", origin)
	if podVeth == nil {
		t.Fatal("pod veth not found")
	}
	if hostVeth == nil {
		t.Fatal("host veth not found")
	}
}

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
