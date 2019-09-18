package envoy

import (
	"fmt"
	"time"

	api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/api/v2/endpoint"
	route "github.com/envoyproxy/go-control-plane/envoy/api/v2/route"

	"github.com/golang/protobuf/ptypes"
)

func buildCluster(clusterName, xdsClusterName string) *api.Cluster {
	return &api.Cluster{
		Name:                 clusterName,
		ConnectTimeout:       ptypes.DurationProto(time.Second * 5),
		ClusterDiscoveryType: &api.Cluster_Type{Type: api.Cluster_EDS},
		EdsClusterConfig: &api.Cluster_EdsClusterConfig{
			EdsConfig: &core.ConfigSource{
				ConfigSourceSpecifier: &core.ConfigSource_ApiConfigSource{
					ApiConfigSource: &core.ApiConfigSource{
						ApiType: core.ApiConfigSource_GRPC,
						GrpcServices: []*core.GrpcService{{
							TargetSpecifier: &core.GrpcService_EnvoyGrpc_{
								EnvoyGrpc: &core.GrpcService_EnvoyGrpc{
									ClusterName: xdsClusterName,
								},
							},
						}},
					},
				},
			},
		},
	}
}

func buildRoute(routeName string, clusterPrefixMap map[string]string) *api.RouteConfiguration {
	routes := make([]*route.Route, len(clusterPrefixMap))

	index := 0
	for prefix, clusterName := range clusterPrefixMap {
		routes[index] = &route.Route{
			Match: &route.RouteMatch{
				PathSpecifier: &route.RouteMatch_Prefix{
					Prefix: prefix,
				},
			},
			Action: &route.Route_Route{
				Route: &route.RouteAction{
					PrefixRewrite: "/",
					ClusterSpecifier: &route.RouteAction_Cluster{
						Cluster: clusterName,
					},
				},
			},
		}
		index++
	}

	return &api.RouteConfiguration{
		Name: routeName,
		VirtualHosts: []*route.VirtualHost{
			{
				Name:    "virtual_host",
				Domains: []string{"*"},
				Routes:  routes,
			},
		},
	}
}

func buildEndpoint(clusterName string, ipAddresses []*core.Address) *api.ClusterLoadAssignment {
	lbEndpoints := make([]*endpoint.LbEndpoint, len(ipAddresses))
	for index, ipAddress := range ipAddresses {
		lbEndpoints[index] = &endpoint.LbEndpoint{
			HostIdentifier: &endpoint.LbEndpoint_Endpoint{
				Endpoint: &endpoint.Endpoint{
					Address: ipAddress,
				},
			},
		}
	}

	return &api.ClusterLoadAssignment{
		ClusterName: clusterName,
		Endpoints: []*endpoint.LocalityLbEndpoints{
			{
				LbEndpoints: lbEndpoints,
			},
		},
	}
}

func buildAddress(ipAddress string, port uint32) *core.Address {
	return &core.Address{
		Address: &core.Address_SocketAddress{
			SocketAddress: &core.SocketAddress{
				Protocol: core.SocketAddress_TCP,
				Address:  ipAddress,
				PortSpecifier: &core.SocketAddress_PortValue{
					PortValue: port,
				},
			},
		},
	}
}

func ipAddressPort(ipAddress string, port uint32) string {
	return fmt.Sprintf("%s:%d", ipAddress, port)
}
