package main

import (
	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/version"
	"github.com/spacechunks/platform/internal/tun"
)

func main() {
	skel.PluginMainFuncs(tun.CNIFuncs(), version.All, "TODO")
}
