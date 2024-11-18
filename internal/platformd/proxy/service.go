package proxy

import (
	"context"
	"fmt"
	"net/netip"

	"github.com/spacechunks/platform/internal/platformd/proxy/xds"
)

type Service interface {
	CreateListener(addr netip.Addr) error
	ConfigureDNS(ctx context.Context, listener, upstreamAddr netip.AddrPort) error
}

type proxyService struct {
	resourceMap *xds.Map
}

func NewService(resourceMap *xds.Map) Service {
	return &proxyService{
		resourceMap: resourceMap,
	}
}

func (s *proxyService) ConfigureDNS(ctx context.Context, listener, upstreamAddr netip.AddrPort) error {
	const cluster = "dns"
	rg, err := dnsResourceGroup(cluster, listener, upstreamAddr)
	if err != nil {
		return fmt.Errorf("create dns resources: %w", err)
	}

	if _, err := s.resourceMap.Apply(ctx, dnsGroupKey, rg); err != nil {
		return fmt.Errorf("apply envoy config: %w", err)
	}

	return nil
}

func (s *proxyService) CreateListener(addr netip.Addr) error {
	return nil
}
