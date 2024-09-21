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
	"encoding/binary"
	"fmt"
	"github.com/cilium/ebpf"
	"net"
	"net/netip"
)

func AddDNATTarget(key uint16, ip netip.Addr, ifaceIdx uint8, mac net.HardwareAddr) error {
	var maps dnatMaps
	if err := loadDnatObjects(&maps, &ebpf.CollectionOptions{
		Maps: ebpf.MapOptions{
			PinPath: pinPath,
		},
	}); err != nil {
		return fmt.Errorf("load dnat maps: %w", err)
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

func AddSNATTarget(key uint8, ip netip.Addr, ifaceIdx uint8) error {
	var maps snatMaps
	if err := loadSnatObjects(&maps, &ebpf.CollectionOptions{
		Maps: ebpf.MapOptions{
			PinPath: pinPath,
		},
	}); err != nil {
		return fmt.Errorf("load snat maps: %w", err)
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
