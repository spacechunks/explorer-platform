package xds

import (
	"context"
	"fmt"
	"sync"

	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	endpointv3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
)

// ResourceGroup should be used for grouping related resources.
// For example all resources related to a DNS proxy should be
// placed into one specific instance of a ResourceGroup.
type ResourceGroup struct {
	Clusters  []*clusterv3.Cluster
	Listeners []*listenerv3.Listener
	CLAS      []*endpointv3.ClusterLoadAssignment
}

// Map is responsible for holding and applying envoy configuration resources.
// This is necessary, because when applying new resources all previous ones that
// are not contained in the snapshot will be removed by envoy. This map is safe
// for concurrent use.
type Map struct {
	cache     cache.SnapshotCache
	mu        sync.Mutex
	resources map[string]ResourceGroup
}

func (m *Map) Get(key string) ResourceGroup {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.resources[key]
}

// Apply saves the passed resource group in the map under the provided
// key, creates a new snapshot and applies it to all known envoy nodes at the time.
func (m *Map) Apply(ctx context.Context, key string, rg ResourceGroup) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.resources[key] = rg

	typeToRes := make(map[resource.Type][]types.Resource)

	// update snapshot
	for _, v := range m.resources {
		for _, c := range v.Clusters {
			typeToRes[resource.ListenerType] = append(typeToRes[resource.ClusterType], c)
		}
		for _, l := range v.Listeners {
			typeToRes[resource.ListenerType] = append(typeToRes[resource.ListenerType], l)
		}
		for _, cla := range v.CLAS {
			typeToRes[resource.ListenerType] = append(typeToRes[resource.EndpointType], cla)
		}
	}

	// TODO: somehow determine version
	snap, err := cache.NewSnapshot("", typeToRes)
	if err != nil {
		return fmt.Errorf("create snapshot: %w", err)
	}

	for _, nodeID := range m.cache.GetStatusKeys() {
		if err := m.cache.SetSnapshot(ctx, nodeID, snap); err != nil {
			return fmt.Errorf("set snapshot: %w", err)
		}
	}

	return nil
}
