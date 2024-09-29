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

package gobpf

import (
	"encoding/binary"
	"fmt"
	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"net"
	"net/netip"
)

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -cc clang-18 -strip llvm-strip-18 snat ../bpf/snat.c -- -I ../bpf/include
//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -cc clang-18 -strip llvm-strip-18 dnat ../bpf/dnat.c -- -I ../bpf/include
//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -cc clang-18 -strip llvm-strip-18 arp ../bpf/arp.c -- -I ../bpf/include

// TODO: tests

func AttachAndPinSNAT(ifaceName string, ifaceIndex int, pinPath string) error {
	var snatProgs snatPrograms
	if err := loadSnatObjects(&snatProgs, &ebpf.CollectionOptions{
		Maps: ebpf.MapOptions{
			PinPath: pinPath,
		},
	}); err != nil {
		return fmt.Errorf("load objs: %w", err)
	}
	l, err := link.AttachTCX(link.TCXOptions{
		Interface: ifaceIndex,
		Program:   snatProgs.Snat,
		Attach:    ebpf.AttachTCXIngress,
	})
	if err != nil {
		return fmt.Errorf("attach: %w", err)
	}

	// pin because cni is short-lived
	if err := l.Pin(fmt.Sprintf("/sys/fs/bpf/ptp_snat_%s", ifaceName)); err != nil {
		return fmt.Errorf("pin link: %w", err)
	}

	return nil
}

func AttachAndPinDNAT(ifaceName string, ifaceIndex int, pinPath string) error {
	var dnatObjs dnatObjects
	if err := loadDnatObjects(&dnatObjs, &ebpf.CollectionOptions{
		Maps: ebpf.MapOptions{
			PinPath: pinPath,
		},
	}); err != nil {
		return fmt.Errorf("load dnat objs: %w", err)
	}

	// TODO: if link is already present just update
	l, err := link.AttachTCX(link.TCXOptions{
		Interface: ifaceIndex,
		Program:   dnatObjs.Dnat,
		Attach:    ebpf.AttachTCXIngress,
	})
	if err != nil {
		return fmt.Errorf("attach: %w", err)
	}

	// pin because cni is short-lived
	if err := l.Pin(fmt.Sprintf("/sys/fs/bpf/ptp_dnat_%s", ifaceName)); err != nil {
		return fmt.Errorf("pin link: %w", err)
	}

	return nil
}

func AttachAndPinARP(ifaceName string, ifaceIndex int) error {
	var arpObjs arpObjects
	if err := loadArpObjects(&arpObjs, nil); err != nil {
		return fmt.Errorf("load dnat objs: %w", err)
	}

	// TODO: if link is already present just update
	l, err := link.AttachTCX(link.TCXOptions{
		Interface: ifaceIndex,
		Program:   arpObjs.Arp,
		Attach:    ebpf.AttachTCXIngress,
	})
	if err != nil {
		return fmt.Errorf("attach: %w", err)
	}

	// pin because cni is short-lived
	if err := l.Pin(fmt.Sprintf("/sys/fs/bpf/ptp_arp_%s", ifaceName)); err != nil {
		return fmt.Errorf("pin link: %w", err)
	}

	return nil
}

func AddDNATTarget(key uint16, ip netip.Addr, ifaceIdx uint8, mac net.HardwareAddr, pinPath string) error {
	var maps dnatMaps
	if err := loadDnatObjects(&maps, &ebpf.CollectionOptions{
		Maps: ebpf.MapOptions{
			PinPath: pinPath,
		},
	}); err != nil {
		return fmt.Errorf("load maps: %w", err)
	}

	sl := ip.As4()
	if err := maps.PtpDnatTargets.Put(key, dnatDnatTarget{
		IpAddr:   binary.BigEndian.Uint32(sl[:]), // network byte order is big endian
		IfaceIdx: ifaceIdx,
		MacAddr:  [6]byte(mac),
	}); err != nil {
		return fmt.Errorf("put config: %w", err)
	}

	return nil
}

func AddSNATTarget(key uint8, ip netip.Addr, ifaceIdx uint8, pinPath string) error {
	var maps snatMaps
	if err := loadSnatObjects(&maps, &ebpf.CollectionOptions{
		Maps: ebpf.MapOptions{
			PinPath: pinPath,
		},
	}); err != nil {
		return fmt.Errorf("load maps: %w", err)
	}

	sl := ip.As4()
	if err := maps.PtpSnatConfig.Put(key, snatPtpSnatEntry{
		IpAddr:   binary.BigEndian.Uint32(sl[:]), // network byte order is big endian
		IfaceIdx: ifaceIdx,
	}); err != nil {
		return fmt.Errorf("put config: %w", err)
	}

	return nil
}
