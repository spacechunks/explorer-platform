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
	"fmt"
	"net"
)

func IfaceByName(name string) (Iface, error) {
	iface, err := net.InterfaceByName(name)
	if err != nil {
		return Iface{}, fmt.Errorf("get iface: %w", err)
	}
	return Iface{
		Index: iface.Index,
		Name:  iface.Name,
	}, nil
}
