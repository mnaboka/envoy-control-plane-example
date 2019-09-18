package envoy

import "github.com/envoyproxy/go-control-plane/pkg/cache"

// Manager manages envoy dynamic config
type Manager interface {
	// Add a new cluster with a given name and route prefix to envoy config.
	AddCluster(name, routePrefix string) error

	// RemoveCluster removes a cluster from envoy config.
	RemoveCluster(name string) error

	// AddEndpoint adds a new endpoint to existing cluster.
	AddEndpoint(cluster, ipAddress string, port uint32) error

	// RemoveEndpoint removes an endpoint from existing cluster.
	RemoveEndpoint(cluster, ipAddress string, port uint32) error

	// Cache returns a control plane cache.
	Cache() cache.Cache

	// Commit updates the configuration. Must be called after any change to clusters or endpoints.
	Commit() error
}
