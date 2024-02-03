package testdata

import _ "embed"

//go:embed unpack-img.tar.gz
var UnpackImage []byte

//go:embed repack-img.tar.gz
var RepackImage []byte
