package envoy

import (
	"testing"

	api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func initManager() *simpleManager {
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)

	m := New("xds_cluster", log)
	sm := m.(*simpleManager)
	sm.clustersVersion = 0
	sm.routeVersion = 0
	sm.endpointsVersion = 0
	return sm
}

func TestManager_Cluster(t *testing.T) {
	m := initManager()
	err := m.AddCluster("test1", "/prefix1")

	assert.Nil(t, err)
	err = m.AddCluster("test2", "/prefix2")
	assert.Nil(t, err)

	assert.Len(t, m.clusterRoutePrefix, 2)
	assert.Contains(t, m.clusterRoutePrefix, "/prefix1")
	assert.Contains(t, m.clusterRoutePrefix, "/prefix2")

	snapshot := m.buildSnapshot()
	assert.Equal(t, "2", snapshot.Clusters.Version)
	assert.Len(t, snapshot.Clusters.Items, 2)

	assert.Len(t, snapshot.Routes.Items, 1)
	assert.Equal(t, "2", snapshot.Routes.Version)

	// must be 2 routes
	routeCfg, ok := snapshot.Routes.Items[routeName].(*api.RouteConfiguration)
	assert.True(t, ok)
	// must be 2 virtual hosts /prefix1 and /prefix2
	assert.Len(t, routeCfg.VirtualHosts, 1)
	assert.Len(t, routeCfg.VirtualHosts[0].Routes, 2)
	assert.Equal(t, "/prefix1", routeCfg.VirtualHosts[0].Routes[0].Match.GetPrefix())
	assert.Equal(t, "/prefix2", routeCfg.VirtualHosts[0].Routes[1].Match.GetPrefix())

	err = m.RemoveCluster("test1")
	assert.Nil(t, err)

	assert.Len(t, m.clusterRoutePrefix, 1)
	assert.Contains(t, m.clusterRoutePrefix, "/prefix2")

	snapshot = m.buildSnapshot()
	assert.Equal(t, "3", snapshot.Clusters.Version)
	assert.Len(t, snapshot.Clusters.Items, 1)

	assert.Len(t, snapshot.Routes.Items, 1)
	assert.Equal(t, "3", snapshot.Routes.Version)

	// try to add a duplicate cluster
	err = m.AddCluster("test2", "/prefix2")
	assert.NotNil(t, err)
	assert.Equal(t, ErrClusterAlreadyExists, err)

	// try to add a cluster with different name but existing routePrefix
	err = m.AddCluster("test42", "/prefix2")
	assert.NotNil(t, err)
	assert.Equal(t, ErrPrefixAlreadyExists, err)
}

func TestManager_Endpoint(t *testing.T) {
	m := initManager()
	err := m.AddCluster("test1", "/prefix1")
	assert.Nil(t, err)

	err = m.AddEndpoint("test1", "127.0.0.1", 8080)
	assert.Nil(t, err)
	assert.Contains(t, m.clustersMap, "test1")
	assert.Len(t, m.clustersMap["test1"].addressesMap, 1)
	assert.Contains(t, m.clustersMap["test1"].addressesMap, "127.0.0.1:8080")

	err = m.AddEndpoint("test1", "127.0.0.2", 8080)
	assert.Nil(t, err)
	assert.Len(t, m.clustersMap["test1"].addressesMap, 2)
	assert.Contains(t, m.clustersMap["test1"].addressesMap, "127.0.0.2:8080")

	snapshot := m.buildSnapshot()
	assert.Equal(t, "2", snapshot.Endpoints.Version)
	assert.Len(t, snapshot.Endpoints.Items, 1)
	cla, ok := snapshot.Endpoints.Items["test1"].(*api.ClusterLoadAssignment)
	assert.True(t, ok)
	assert.Len(t, cla.Endpoints[0].LbEndpoints, 2)

	err = m.AddCluster("test2", "/prefix2")
	assert.Nil(t, err)

	err = m.AddEndpoint("test2", "10.10.0.1", 8080)
	assert.Nil(t, err)

	err = m.AddEndpoint("test2", "10.10.0.2", 8080)
	assert.Nil(t, err)

	snapshot = m.buildSnapshot()
	assert.Len(t, snapshot.Clusters.Items, 2)
	assert.Equal(t, "2", snapshot.Clusters.Version)
	assert.Len(t, snapshot.Endpoints.Items, 2)
	assert.Equal(t, "4", snapshot.Endpoints.Version)
}

func TestManager_RemoveClusterWithEndpoints(t *testing.T) {
	m := initManager()
	err := m.AddCluster("test1", "/prefix1")
	assert.Nil(t, err)

	err = m.AddEndpoint("test1", "127.0.0.1", 8080)
	assert.Nil(t, err)
	err = m.AddEndpoint("test1", "127.0.0.2", 8081)
	assert.Nil(t, err)

	err = m.RemoveCluster("test1")
	assert.NotNil(t, err)
	assert.Equal(t, ErrClusterHasEndpoints, err)

	assert.Contains(t, m.clustersMap, "test1")
	assert.Contains(t, m.clustersMap["test1"].addressesMap, "127.0.0.1:8080")
	assert.Contains(t, m.clustersMap["test1"].addressesMap, "127.0.0.2:8081")

	err = m.RemoveEndpoint("test1", "127.0.0.1", 8080)
	assert.Nil(t, err)
	err = m.RemoveEndpoint("test1", "127.0.0.2", 8081)
	assert.Nil(t, err)

	err = m.RemoveCluster("test1")
	assert.Nil(t, err)
	assert.Empty(t, m.clustersMap)

	snapshot := m.buildSnapshot()
	assert.Empty(t, snapshot.Endpoints)
	assert.Empty(t, snapshot.Clusters)
}
