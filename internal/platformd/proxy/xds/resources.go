package xds

import (
	"fmt"
	"net/netip"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpointv3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

type ListenerConfig struct {
	ListenerName string
	StatPrefix   string
	Addr         netip.AddrPort
	Proto        corev3.SocketAddress_Protocol
	FilterName   string
	FilterCfg    proto.Message
}

// CreateListener creates a listener with a single listener filter config
func CreateListener(cfg ListenerConfig) (*listenerv3.Listener, error) {
	var filterCfgAny anypb.Any
	if err := anypb.MarshalFrom(&filterCfgAny, cfg.FilterCfg, proto.MarshalOptions{}); err != nil {
		return nil, fmt.Errorf("marshal to any: %w", err)
	}
	return &listenerv3.Listener{
		Name: cfg.ListenerName,
		Address: &corev3.Address{
			Address: &corev3.Address_SocketAddress{
				SocketAddress: &corev3.SocketAddress{
					Protocol: cfg.Proto,
					Address:  cfg.Addr.Addr().String(),
					PortSpecifier: &corev3.SocketAddress_PortValue{
						PortValue: uint32(cfg.Addr.Port()),
					},
				},
			},
		},
		StatPrefix: cfg.StatPrefix,
		ListenerFilters: []*listenerv3.ListenerFilter{
			{
				Name: cfg.StatPrefix,
				ConfigType: &listenerv3.ListenerFilter_TypedConfig{
					TypedConfig: &filterCfgAny,
				},
			},
		},
	}, nil
}

// CreateCLA creates a cluster load assignment with a single upstream endpoint
func CreateCLA(
	clusterName string,
	upstream netip.AddrPort,
	proto corev3.SocketAddress_Protocol,
) *endpointv3.ClusterLoadAssignment {
	return &endpointv3.ClusterLoadAssignment{
		ClusterName: clusterName,
		Endpoints: []*endpointv3.LocalityLbEndpoints{
			{
				LbEndpoints: []*endpointv3.LbEndpoint{
					{
						HostIdentifier: &endpointv3.LbEndpoint_Endpoint{
							Endpoint: &endpointv3.Endpoint{
								Address: &corev3.Address{
									Address: &corev3.Address_SocketAddress{
										SocketAddress: &corev3.SocketAddress{
											Protocol: proto,
											Address:  upstream.Addr().String(),
											PortSpecifier: &corev3.SocketAddress_PortValue{
												PortValue: uint32(upstream.Port()),
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
}
