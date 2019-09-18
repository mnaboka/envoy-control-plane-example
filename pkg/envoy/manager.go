package envoy

import (
	"math/rand"
	"strconv"
	"sync"
	"time"

	api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	"github.com/envoyproxy/go-control-plane/pkg/cache"
	"github.com/sirupsen/logrus"
)

const routeName = "generic_route"

func New(xdsClusterName string, log *logrus.Logger) Manager {
	rand.Seed(time.Now().UnixNano())
	return &simpleManager{
		xdsClusterName: xdsClusterName,
		log:            log,
		cache:          cache.NewSnapshotCache(false, simpleNodeHash(defaultNodeHash), log),

		clustersMap:        make(map[string]properties),
		clusterRoutePrefix: make(map[string]string),

		// set versions to random values
		clustersVersion:  rand.Uint64(),
		endpointsVersion: rand.Uint64(),
		routeVersion:     rand.Uint64(),
	}
}

type simpleManager struct {
	sync.RWMutex
	log *logrus.Logger

	// contains a XDS server name, usually must be "xds_server"
	xdsClusterName string

	// contains snapshots with configurations
	cache cache.SnapshotCache

	// versions
	clustersVersion  uint64
	endpointsVersion uint64
	routeVersion     uint64

	clustersMap map[string]properties

	// mapping of a prefix to a cluster name, it must be unique
	clusterRoutePrefix map[string]string

	routesConfig *api.RouteConfiguration
}

func (m *simpleManager) routes() []cache.Resource {
	if len(m.routesConfig.VirtualHosts) > 0 {
		return []cache.Resource{m.routesConfig}
	}
	return nil
}

type properties struct {
	cluster      *api.Cluster
	addressesMap map[string]*core.Address
}

func (p *properties) addresses() []*core.Address {
	addresses := make([]*core.Address, len(p.addressesMap))

	index := 0
	for _, address := range p.addressesMap {
		addresses[index] = address
		index++
	}
	return addresses
}

func (m *simpleManager) buildSnapshot() cache.Snapshot {
	m.RLock()
	defer m.RUnlock()

	snapshot := cache.Snapshot{}

	// set the snapshot fields only for non empty resources
	if clusterResources := m.clusterResources(); len(clusterResources) > 0 {
		snapshot.Clusters = cache.NewResources(strconv.FormatUint(m.clustersVersion, 10), clusterResources)
	}

	if clusterLoadAssignmentResources := m.clusterLoadAssignmentResources(); len(clusterLoadAssignmentResources) > 0 {
		snapshot.Endpoints = cache.NewResources(strconv.FormatUint(m.endpointsVersion, 10), clusterLoadAssignmentResources)
	}

	if routes := m.routes(); len(routes) > 0 {
		snapshot.Routes = cache.NewResources(strconv.FormatUint(m.routeVersion, 10), routes)
	}

	return snapshot
}

func (m *simpleManager) clusterResources() []cache.Resource {
	resources := make([]cache.Resource, len(m.clustersMap))

	index := 0
	for _, props := range m.clustersMap {
		resources[index] = props.cluster
		index++
	}

	return resources
}

func (m *simpleManager) clusterLoadAssignmentResources() []cache.Resource {
	loadAssignments := []cache.Resource{}

	for clusterName, props := range m.clustersMap {
		if addresses := props.addresses(); len(addresses) > 0 {
			loadAssignments = append(loadAssignments, buildEndpoint(clusterName, addresses))
		}
	}

	return loadAssignments
}

func (m *simpleManager) Commit() error {
	err := m.cache.SetSnapshot(defaultNodeHash, m.buildSnapshot())
	if err != nil {
		return err
	}

	return nil
}

func (m *simpleManager) Cache() cache.Cache {
	return m.cache
}

// Cluster methods
func (m *simpleManager) AddCluster(name, routePrefix string) error {
	if name == "" || routePrefix == "" {
		return ErrRequiredParameterMissing
	}

	m.Lock()
	defer m.Unlock()

	if _, ok := m.clustersMap[name]; ok {
		return ErrClusterAlreadyExists
	}

	if _, ok := m.clusterRoutePrefix[routePrefix]; ok {
		return ErrPrefixAlreadyExists
	}

	// update prefix to a cluster name
	m.clusterRoutePrefix[routePrefix] = name

	// build a new route for a cluster with prefix
	m.routesConfig = buildRoute(routeName, m.clusterRoutePrefix)

	// update cluster resources
	m.clustersMap[name] = properties{
		cluster:      buildCluster(name, m.xdsClusterName),
		addressesMap: map[string]*core.Address{},
	}

	// bump versions
	m.clustersVersion++
	m.routeVersion++

	return nil
}

func (m *simpleManager) RemoveCluster(name string) error {
	if name == "" {
		return ErrRequiredParameterMissing
	}

	m.Lock()
	defer m.Unlock()

	if _, ok := m.clustersMap[name]; !ok {
		return ErrClusterNotFound
	}

	if len(m.clustersMap[name].addressesMap) > 0 {
		return ErrClusterHasEndpoints
	}

	delete(m.clustersMap, name)
	m.removePrefix(name)
	m.routesConfig = buildRoute(routeName, m.clusterRoutePrefix)

	// bump versions
	m.clustersVersion++
	m.routeVersion++

	return nil
}

func (m *simpleManager) removePrefix(name string) {
	for prefix, cluster := range m.clusterRoutePrefix {
		if cluster == name {
			delete(m.clusterRoutePrefix, prefix)
			return
		}
	}
}

// Endpoint methods
func (m *simpleManager) AddEndpoint(cluster, ipAddress string, port uint32) error {
	if cluster == "" || ipAddress == "" || port == 0 {
		return ErrRequiredParameterMissing
	}

	m.Lock()
	defer m.Unlock()

	if _, ok := m.clustersMap[cluster]; !ok {
		return ErrClusterNotFound
	}

	endpoint := ipAddressPort(ipAddress, port)
	if _, ok := m.clustersMap[cluster].addressesMap[endpoint]; ok {
		return ErrEndpointAlreadyExists
	}

	m.clustersMap[cluster].addressesMap[endpoint] = buildAddress(ipAddress, port)
	m.endpointsVersion++

	return nil
}

func (m *simpleManager) RemoveEndpoint(cluster, ipAddress string, port uint32) error {
	if cluster == "" || ipAddress == "" || port == 0 {
		return ErrRequiredParameterMissing
	}

	m.Lock()
	defer m.Unlock()

	if _, ok := m.clustersMap[cluster]; !ok {
		return ErrClusterNotFound
	}

	endpoint := ipAddressPort(ipAddress, port)
	if _, ok := m.clustersMap[cluster].addressesMap[endpoint]; !ok {
		return ErrEndpointNotFound
	}

	delete(m.clustersMap[cluster].addressesMap, endpoint)
	m.endpointsVersion++

	return nil
}
