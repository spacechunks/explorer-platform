package proxy

import (
	"context"
	"fmt"
	"log/slog"
	"net/netip"
	"time"

	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/spacechunks/platform/internal/platformd/proxy/xds"
)

type Service interface {
	CreateListeners(ctx context.Context, workloadID string, addr netip.Addr) error
	ApplyOriginalDstCluster(ctx context.Context) error
}

type proxyService struct {
	logger      *slog.Logger
	cfg         Config
	resourceMap *xds.Map
}

func NewService(logger *slog.Logger, cfg Config, resourceMap *xds.Map) Service {
	return &proxyService{
		logger:      logger,
		cfg:         cfg,
		resourceMap: resourceMap,
	}
}

// CreateListeners creates HTTP, TCP as well as UDP(DNS) and TCP(DNS) listeners for the provided
// workload. this will fail if the workload does not exist.
func (s *proxyService) CreateListeners(ctx context.Context, workloadID string, addr netip.Addr) error {
	wrg, err := workloadResources(
		workloadID,
		netip.AddrPortFrom(addr, proxyHTTPPort),
		netip.AddrPortFrom(addr, proxyTCPPort),
		originalDstClusterName,
	)
	if err != nil {
		return fmt.Errorf("create workload resources: %w", err)
	}

	drg, err := dnsResourceGroup(dnsCluster, netip.AddrPortFrom(addr, proxyDNSPort), s.cfg.DNSUpstream)
	if err != nil {
		return fmt.Errorf("create dns resources: %w", err)
	}

	merged := xds.ResourceGroup{}
	merged.Listeners = append(wrg.Listeners, drg.Listeners...)
	merged.Clusters = append(wrg.Clusters, drg.Clusters...)
	merged.CLAS = append(wrg.CLAS, drg.CLAS...)

	s.logger.InfoContext(ctx, "applying workload resources", "workload_id", workloadID)

	if _, err := s.resourceMap.Apply(ctx, workloadID, merged); err != nil {
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
