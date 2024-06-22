package mockgen

//go:generate go run github.com/vektra/mockery/v2 --srcpkg github.com/spacechunks/platform/internal/tun --name Handler --outpkg mock --output internal/mock --structname Handler --filename cni_handler.go --with-expecter=true --disable-version-string
