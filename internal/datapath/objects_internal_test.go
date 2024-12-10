//go:build functests

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
	"net"
	"net/netip"
	"testing"

	"github.com/cilium/ebpf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddSNATTarget(t *testing.T) {
	objs, err := LoadBPF()
	require.NoError(t, err)

	var (
		ip       = netip.MustParseAddr("10.0.0.1")
		ifaceIdx = uint8(3)
		expected = snatPtpSnatEntry{
			IpAddr:   uint32(167772161), // 10.0.0.1 in big endian decimal
			IfaceIdx: ifaceIdx,
		}
	)

	require.NoError(t, objs.AddSNATTarget(0, ip, ifaceIdx))

	var maps snatMaps
	err = loadSnatObjects(&maps, &ebpf.CollectionOptions{
		Maps: ebpf.MapOptions{
			PinPath: pinPath,
		},
	})
	require.NoError(t, err)

	var actual snatPtpSnatEntry
	err = maps.PtpSnatConfig.Lookup(uint8(0), &actual)
	require.NoError(t, err)

	assert.Equal(t, expected, actual)
}

func TestAddDNATTarget(t *testing.T) {
	objs, err := LoadBPF()
	require.NoError(t, err)

	var (
		ip       = netip.MustParseAddr("10.0.0.1")
		ifaceIdx = uint8(3)
		hwAddr   = net.HardwareAddr{
			0x7e, 0x90, 0xc4, 0xed, 0xdf, 0xd0,
		}
		expected = dnatDnatTarget{
			IpAddr:   uint32(167772161), // 10.0.0.1 in big endian decimal
			IfaceIdx: ifaceIdx,
			MacAddr: [6]uint8{
				0x7e, 0x90, 0xc4, 0xed, 0xdf, 0xd0,
			},
		}
	)

	require.NoError(t, objs.AddDNATTarget(0, ip, ifaceIdx, hwAddr))

	var maps dnatMaps
	err = loadDnatObjects(&maps, &ebpf.CollectionOptions{
		Maps: ebpf.MapOptions{
			PinPath: pinPath,
		},
	})
	require.NoError(t, err)

	var actual dnatDnatTarget
	err = maps.PtpDnatTargets.Lookup(uint16(0), &actual)
	require.NoError(t, err)

	assert.Equal(t, expected, actual)
}
