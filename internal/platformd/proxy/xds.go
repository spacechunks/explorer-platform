package proxy

import (
	"context"

	discoverygrpc "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	serverv3 "github.com/envoyproxy/go-control-plane/pkg/server/v3"

	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"google.golang.org/grpc"
)

func CreateAndRegisterXDSServer(ctx context.Context, server *grpc.Server, cache cache.SnapshotCache) {
	xdsServer := serverv3.NewServer(ctx, cache, nil)
	discoverygrpc.RegisterAggregatedDiscoveryServiceServer(server, xdsServer)
}
