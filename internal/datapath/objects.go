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

package datapath

import (
	"encoding/binary"
	"fmt"
	"net"
	"net/netip"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
)

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -cc clang-18 -strip llvm-strip-18 snat ../bpf/snat.c -- -I ../bpf/include
//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -cc clang-18 -strip llvm-strip-18 dnat ../bpf/dnat.c -- -I ../bpf/include
//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -cc clang-18 -strip llvm-strip-18 arp ../bpf/arp.c -- -I ../bpf/include

const pinPath = "/sys/fs/bpf"

type Objects struct {
	snatObjs snatObjects
	dnatObjs dnatObjects
	arpObjs  arpObjects
	pinPath  string
}

func LoadObjects() (*Objects, error) {
	var snatObjs snatObjects
	if err := loadSnatObjects(&snatObjs, &ebpf.CollectionOptions{
		Maps: ebpf.MapOptions{
			PinPath: pinPath,
		},
	}); err != nil {
		return nil, fmt.Errorf("load snat objs: %w", err)
	}

	var dnatObjs dnatObjects
	if err := loadDnatObjects(&dnatObjs, &ebpf.CollectionOptions{
		Maps: ebpf.MapOptions{
			PinPath: pinPath,
		},
	}); err != nil {
		return nil, fmt.Errorf("load dnat objs: %w", err)
	}

	var arpObjs arpObjects
	if err := loadArpObjects(&arpObjs, nil); err != nil {
		return nil, fmt.Errorf("load arp objs: %w", err)
	}

	return &Objects{
		snatObjs: snatObjs,
		dnatObjs: dnatObjs,
		arpObjs:  arpObjs,
		pinPath:  pinPath,
	}, nil
}

func (o *Objects) AttachAndPinSNAT(ifaceName string, ifaceIndex int) error {
	l, err := link.AttachTCX(link.TCXOptions{
		Interface: ifaceIndex,
		Program:   o.snatObjs.Snat,
		Attach:    ebpf.AttachTCXIngress,
	})
	if err != nil {
		return fmt.Errorf("attach: %w", err)
	}

	// pin because cni is short-lived
	if err := l.Pin(fmt.Sprintf("%s/ptp_snat_%s", o.pinPath, ifaceName)); err != nil {
		return fmt.Errorf("pin link: %w", err)
	}

	return nil
}

func (o *Objects) AttachAndPinDNAT(ifaceName string, ifaceIndex int) error {
	l, err := link.AttachTCX(link.TCXOptions{
		Interface: ifaceIndex,
		Program:   o.dnatObjs.Dnat,
		Attach:    ebpf.AttachTCXIngress,
	})
	if err != nil {
		return fmt.Errorf("attach: %w", err)
	}

	// pin because cni is short-lived
	if err := l.Pin(fmt.Sprintf("%s/ptp_dnat_%s", o.pinPath, ifaceName)); err != nil {
		return fmt.Errorf("pin link: %w", err)
	}

	return nil
}

func (o *Objects) AttachAndPinARP(ifaceName string, ifaceIndex int) error {
	l, err := link.AttachTCX(link.TCXOptions{
		Interface: ifaceIndex,
		Program:   o.arpObjs.Arp,
		Attach:    ebpf.AttachTCXIngress,
	})
	if err != nil {
		return fmt.Errorf("attach: %w", err)
	}

	// pin because cni is short-lived
	if err := l.Pin(fmt.Sprintf("%s/ptp_arp_%s", o.pinPath, ifaceName)); err != nil {
		return fmt.Errorf("pin link: %w", err)
	}

	return nil
}

func (o *Objects) AddDNATTarget(key uint16, ip netip.Addr, ifaceIdx uint8, mac net.HardwareAddr) error {
	sl := ip.As4()
	if err := o.dnatObjs.PtpDnatTargets.Put(key, dnatDnatTarget{
		IpAddr:   binary.BigEndian.Uint32(sl[:]), // network byte order is big endian
		IfaceIdx: ifaceIdx,
		MacAddr:  [6]byte(mac),
	}); err != nil {
		return fmt.Errorf("put config: %w", err)
	}

	return nil
}

func (o *Objects) AddSNATTarget(key uint8, ip netip.Addr, ifaceIdx uint8) error {
	sl := ip.As4()
	if err := o.snatObjs.PtpSnatConfig.Put(key, snatPtpSnatEntry{
		IpAddr:   binary.BigEndian.Uint32(sl[:]), // network byte order is big endian
		IfaceIdx: ifaceIdx,
	}); err != nil {
		return fmt.Errorf("put config: %w", err)
	}

	return nil
}
