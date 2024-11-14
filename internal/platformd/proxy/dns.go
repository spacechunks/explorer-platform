package proxy

import (
	"fmt"
	"net/netip"

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

func dnsClusterResources(upstreamAddr netip.AddrPort, listenerAddr netip.AddrPort) (
	*clusterv3.Cluster,
	*endpointv3.ClusterLoadAssignment,
	*listenerv3.Listener,
	error,
) {
	const dnsClusterName = "dns"

	dnsRoute := &udpproxyv3.Route{
		Cluster: dnsClusterName,
	}

	var dnsRouteAny anypb.Any
	if err := anypb.MarshalFrom(&dnsRouteAny, dnsRoute, proto.MarshalOptions{}); err != nil {
		return nil, nil, nil, fmt.Errorf("dns route to any: %w", err)
	}

	proxyCfg := &udpproxyv3.UdpProxyConfig{
		StatPrefix: "dns",
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
	}

	var proxyCfgAny anypb.Any
	if err := anypb.MarshalFrom(&proxyCfgAny, proxyCfg, proto.MarshalOptions{}); err != nil {
		return nil, nil, nil, fmt.Errorf("udp proxy config to any: %w", err)
	}

	l := &listenerv3.Listener{
		Name: "dns_listener",
		Address: &corev3.Address{
			Address: &corev3.Address_SocketAddress{
				SocketAddress: &corev3.SocketAddress{
					Protocol: corev3.SocketAddress_UDP,
					Address:  listenerAddr.Addr().String(),
					PortSpecifier: &corev3.SocketAddress_PortValue{
						PortValue: uint32(listenerAddr.Port()),
					},
				},
			},
		},
		StatPrefix: "dns_listener",
		ListenerFilters: []*listenerv3.ListenerFilter{
			{
				Name: "dns_udp_proxy",
				ConfigType: &listenerv3.ListenerFilter_TypedConfig{
					TypedConfig: &proxyCfgAny,
				},
			},
		},
		UdpListenerConfig: &listenerv3.UdpListenerConfig{
			DownstreamSocketConfig: &corev3.UdpSocketConfig{
				MaxRxDatagramSize: wrapperspb.UInt64(9000),
			},
		},
	}

	cluster := &clusterv3.Cluster{
		Name: dnsClusterName,
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

	cla := &endpointv3.ClusterLoadAssignment{
		ClusterName: "dns",
		Endpoints: []*endpointv3.LocalityLbEndpoints{
			{
				LbEndpoints: []*endpointv3.LbEndpoint{
					{
						HostIdentifier: &endpointv3.LbEndpoint_Endpoint{
							Endpoint: &endpointv3.Endpoint{
								Address: &corev3.Address{
									Address: &corev3.Address_SocketAddress{
										SocketAddress: &corev3.SocketAddress{
											Protocol: corev3.SocketAddress_UDP,
											Address:  upstreamAddr.Addr().String(),
											PortSpecifier: &corev3.SocketAddress_PortValue{
												PortValue: uint32(upstreamAddr.Port()),
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	return cluster, cla, l, nil
}
