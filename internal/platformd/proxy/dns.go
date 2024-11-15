package proxy

import (
	"fmt"
	"net/netip"

	tcpproxyv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/tcp_proxy/v3"
	"github.com/spacechunks/platform/internal/platformd/proxy/xds"

	xdscorev3 "github.com/cncf/xds/go/xds/core/v3"
	xsdmatcherv3 "github.com/cncf/xds/go/xds/type/matcher/v3"
	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpointv3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	udpproxyv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/udp/udp_proxy/v3"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const dnsGroupKey = "dns"

func dnsResourceGroup(clusterName string, upstreamAddr netip.AddrPort, listenerAddr netip.AddrPort) (xds.ResourceGroup, error) {
	udpCLA, udpListener, err := dnsUDPResources(clusterName, upstreamAddr, listenerAddr)
	if err != nil {
		return xds.ResourceGroup{}, fmt.Errorf("udp resources: %w", err)
	}
	tcpCLA, tcpListener, err := dnsTCPResources(clusterName, upstreamAddr, listenerAddr)
	if err != nil {
		return xds.ResourceGroup{}, fmt.Errorf("tcp resources: %w", err)
	}

	cluster := &clusterv3.Cluster{
		Name: clusterName,
		ClusterDiscoveryType: &clusterv3.Cluster_Type{
			Type: clusterv3.Cluster_EDS,
		},
		EdsClusterConfig: &clusterv3.Cluster_EdsClusterConfig{
			EdsConfig: &corev3.ConfigSource{
				ConfigSourceSpecifier: &corev3.ConfigSource_Ads{},
			},
		},
		LbPolicy: clusterv3.Cluster_ROUND_ROBIN,
	}

	return xds.ResourceGroup{
		Clusters:  []*clusterv3.Cluster{cluster},
		Listeners: []*listenerv3.Listener{udpListener, tcpListener},
		CLAS:      []*endpointv3.ClusterLoadAssignment{udpCLA, tcpCLA},
	}, nil
}

func dnsTCPResources(clusterName string, upstreamAddr netip.AddrPort, listenerAddr netip.AddrPort) (
	*endpointv3.ClusterLoadAssignment,
	*listenerv3.Listener,
	error,
) {
	l, err := xds.CreateListener(xds.ListenerConfig{
		ListenerName: "dns_tcp",
		StatPrefix:   "dns_tcp",
		Addr:         listenerAddr,
		Proto:        corev3.SocketAddress_TCP,
		FilterName:   "dns_tcp_proxy",
		FilterCfg: &tcpproxyv3.TcpProxy{
			StatPrefix: "dns_tcp",
			ClusterSpecifier: &tcpproxyv3.TcpProxy_Cluster{
				Cluster: clusterName,
			},
		},
	})
	if err != nil {
		return nil, nil, fmt.Errorf("create listener: %w", err)
	}
	return xds.CreateCLA(clusterName, upstreamAddr, corev3.SocketAddress_UDP), l, nil
}

func dnsUDPResources(clusterName string, upstreamAddr netip.AddrPort, listenerAddr netip.AddrPort) (
	*endpointv3.ClusterLoadAssignment,
	*listenerv3.Listener,
	error,
) {
	dnsRoute := &udpproxyv3.Route{
		Cluster: clusterName,
	}

	var dnsRouteAny anypb.Any
	if err := anypb.MarshalFrom(&dnsRouteAny, dnsRoute, proto.MarshalOptions{}); err != nil {
		return nil, nil, fmt.Errorf("dns route to any: %w", err)
	}

	l, err := xds.CreateListener(xds.ListenerConfig{
		ListenerName: "dns_udp",
		StatPrefix:   "dns_udp",
		Addr:         listenerAddr,
		Proto:        corev3.SocketAddress_UDP,
		FilterName:   "dns_udp_proxy",
		FilterCfg: &udpproxyv3.UdpProxyConfig{
			StatPrefix: "dns_udp",
			RouteSpecifier: &udpproxyv3.UdpProxyConfig_Matcher{
				Matcher: &xsdmatcherv3.Matcher{
					OnNoMatch: &xsdmatcherv3.Matcher_OnMatch{
						OnMatch: &xsdmatcherv3.Matcher_OnMatch_Action{
							Action: &xdscorev3.TypedExtensionConfig{
								Name:        "route",
								TypedConfig: &dnsRouteAny,
							},
						},
					},
				},
			},
			UpstreamSocketConfig: &corev3.UdpSocketConfig{
				MaxRxDatagramSize: wrapperspb.UInt64(9000),
			},
		},
	})
	if err != nil {
		return nil, nil, fmt.Errorf("create listener: %w", err)
	}

	l.UdpListenerConfig = &listenerv3.UdpListenerConfig{
		DownstreamSocketConfig: &corev3.UdpSocketConfig{
			MaxRxDatagramSize: wrapperspb.UInt64(9000),
		},
	}

	return xds.CreateCLA(clusterName, upstreamAddr, corev3.SocketAddress_UDP), l, nil
}
