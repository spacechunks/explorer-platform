package proxy

import (
	"context"
	"fmt"
	"net/netip"
	"time"

	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/spacechunks/platform/internal/platformd/proxy/xds"
)

type Service interface {
	CreateListener(ctx context.Context, workloadID string, addr netip.Addr) error
	ConfigureDNS(ctx context.Context, listener, upstreamAddr netip.AddrPort) error
	ApplyOriginalDstCluster(ctx context.Context) error
}

type proxyService struct {
	resourceMap *xds.Map
}

func NewService(resourceMap *xds.Map) Service {
	return &proxyService{
		resourceMap: resourceMap,
	}
}

func (s *proxyService) CreateListener(ctx context.Context, workloadID string, addr netip.Addr) error {
	rg, err := workloadResources(
		workloadID,
		netip.AddrPortFrom(addr, 0),
		netip.AddrPortFrom(addr, 0),
		originalDstClusterName,
	)
	if err != nil {
		return fmt.Errorf("create workload resources: %w", err)
	}

	if _, err := s.resourceMap.Apply(ctx, workloadID, rg); err != nil {
		return fmt.Errorf("apply envoy config: %w", err)
	}

	return nil
}

// ConfigureDNS configures a UDP and TCP listener listening on the passed address
// redirecting all traffic to upstreamAddr.
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

// ApplyOriginalDstCluster configures the original destination cluster
// where all traffic originating from the container destined to the
// outside world will be routed through.
func (s *proxyService) ApplyOriginalDstCluster(ctx context.Context) error {
	rg := xds.ResourceGroup{
		Clusters: []*clusterv3.Cluster{
			{
				Name: originalDstClusterName,
				ClusterDiscoveryType: &clusterv3.Cluster_Type{
					Type: clusterv3.Cluster_ORIGINAL_DST,
				},
				ConnectTimeout:  durationpb.New(time.Second * 5),
				DnsLookupFamily: clusterv3.Cluster_V4_ONLY,
			},
		},
	}
	if _, err := s.resourceMap.Apply(ctx, "original_dst_cluster", rg); err != nil {
		return fmt.Errorf("apply envoy config: %w", err)
	}
	return nil
}
