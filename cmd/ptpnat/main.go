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

package main

import (
	"log"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/version"
	"github.com/spacechunks/platform/internal/ptpnat"
)

func main() {
	handler, err := ptpnat.NewHandler()
	if err != nil {
		log.Fatalf("failed to create handler: %v", err)
	}
	cni := ptpnat.NewCNI(handler)
	skel.PluginMainFuncs(skel.CNIFuncs{
		Add:    cni.ExecAdd,
		Del:    cni.ExecDel,
		Check:  nil,
		GC:     nil,
		Status: nil,
	}, version.All, "TODO")
}
