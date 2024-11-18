package xds_test

import (
	"context"
	"testing"

	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	endpointv3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/spacechunks/platform/internal/mock"
	"github.com/spacechunks/platform/internal/platformd/proxy/xds"
	mocky "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestMap(t *testing.T) {
	tests := []struct {
		name  string
		check func(*xds.Map, *mock.MockCacheSnapshotCache)
	}{
		{
			name: "check resource group is saved",
			check: func(m *xds.Map, mockCache *mock.MockCacheSnapshotCache) {
				expectedRg := xds.ResourceGroup{
					Clusters: []*clusterv3.Cluster{
						{
							Name: "c1",
						},
					},
				}

				snap, err := cache.NewSnapshot("1", expectedRg.ResourcesByType())
				require.NoError(t, err)
				mockCache.EXPECT().SetSnapshot(mocky.Anything, mocky.Anything, snap).Return(nil)

				require.NoError(t, m.Apply(context.Background(), "key", expectedRg))
				require.Equal(t, expectedRg, m.Get("key"))
			},
		},
		{
			name: "all resource groups are applied",
			check: func(m *xds.Map, mockCache *mock.MockCacheSnapshotCache) {
				rg1 := xds.ResourceGroup{
					Clusters: []*clusterv3.Cluster{
						{
							Name: "c1",
						},
					},
					Listeners: []*listenerv3.Listener{
						{
							Name: "l1",
						},
						{
							Name: "l2",
						},
					},
					CLAS: []*endpointv3.ClusterLoadAssignment{
						{
							ClusterName: "c1",
						},
					},
				}
				rg2 := xds.ResourceGroup{
					Clusters: []*clusterv3.Cluster{
						{
							Name: "c2",
						},
					},
					CLAS: []*endpointv3.ClusterLoadAssignment{
						{
							ClusterName: "c2",
						},
					},
				}

				merged := make(map[resource.Type][]types.Resource)
				for _, rg := range []xds.ResourceGroup{rg1, rg2} {
					for k, v := range rg.ResourcesByType() {
						merged[k] = append(merged[k], v...)
					}
				}

				snap1, err := cache.NewSnapshot("1", rg1.ResourcesByType())
				require.NoError(t, err)

				snap2, err := cache.NewSnapshot("2", merged)
				require.NoError(t, err)

				mockCache.EXPECT().SetSnapshot(mocky.Anything, mocky.Anything, snap1).Return(nil)
				mockCache.EXPECT().SetSnapshot(mocky.Anything, mocky.Anything, snap2).Return(nil)

				require.NoError(t, m.Apply(context.Background(), "key1", rg1))
				require.NoError(t, m.Apply(context.Background(), "key2", rg2))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				mockCache = mock.NewMockCacheSnapshotCache(t)
				m         = xds.NewMap(mockCache)
			)
			mockCache.EXPECT().GetStatusKeys().Return([]string{"node1"})
			tt.check(m, mockCache)
		})
	}
}
