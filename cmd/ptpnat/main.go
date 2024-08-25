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
	"flag"
	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/version"
	ff "github.com/peterbourgon/ff/v3"
	"github.com/spacechunks/platform/internal/ptpnat"
	"log"
	"os"
)

func main() {
	fs := flag.NewFlagSet("cni", flag.ContinueOnError)
	var (
		hostIface = fs.String("host-iface", "eth0", "network interface where packets from users will arrive on")
		snatMap   = fs.String("snat-map-pin", "", "path to where snat bpf map has been pinned to")
	)
	if err := ff.Parse(fs, os.Args[1:],
		ff.WithEnvVarPrefix("PTPNAT"),
		ff.WithConfigFileParser(ff.PlainParser),
	); err != nil {
		log.Fatalf("failed to parse flags: %v", err)
	}
	var (
		handler = ptpnat.NewHandler()
		cni     = ptpnat.NewCNI(handler)
	)
	if *hostIface != "" {
		log.Fatalf("host interface not specified")
	}
	if *snatMap != "" {
		log.Fatalf("main interface not specified")
	}
	if err := handler.ConfigureSNAT(*snatMap); err != nil {
		log.Fatalf("failed to configure snat: %v", err)
	}
	if err := handler.AttachDNATBPF(*hostIface); err != nil {
		log.Fatalf("failed to attach dnat bpf to %s: %v", *hostIface, err)
	}
	skel.PluginMainFuncs(skel.CNIFuncs{
		Add:    cni.ExecAdd,
		Del:    cni.ExecDel,
		Check:  nil,
		GC:     nil,
		Status: nil,
	}, version.All, "TODO")
}
