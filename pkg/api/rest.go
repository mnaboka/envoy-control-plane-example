package api

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/mnaboka/envoy-control-plane-example/pkg/envoy"
	"net/http"
)

func New(manager envoy.Manager) *mux.Router {
	server := &apiServer{
		manager: manager,
	}

	r := mux.NewRouter()

	r.HandleFunc("/api/v1/cluster", server.addCluster).Methods("POST")
	r.HandleFunc("/api/v1/cluster", server.removeCluster).Methods("DELETE")
	r.HandleFunc("/api/v1/endpoint", server.addEndpoint).Methods("POST")
	r.HandleFunc("/api/v1/endpoint", server.removeEndpoint).Methods("DELETE")
	r.HandleFunc("/api/v1/commit", server.commitChanges).Methods("POST")
	return r
}

type apiServer struct {
	manager envoy.Manager
}

func (api *apiServer) addCluster(w http.ResponseWriter, r *http.Request) {
	type addClusterRequest struct {
		Name   string
		Prefix string
	}

	decoder := json.NewDecoder(r.Body)

	req := &addClusterRequest{}
	err := decoder.Decode(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = api.manager.AddCluster(req.Name, req.Prefix)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (api *apiServer) removeCluster(w http.ResponseWriter, r *http.Request) {
	type removeClusterRequest struct {
		Name string
	}

	decoder := json.NewDecoder(r.Body)

	req := &removeClusterRequest{}
	err := decoder.Decode(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = api.manager.RemoveCluster(req.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (api *apiServer) addEndpoint(w http.ResponseWriter, r *http.Request) {
	type addEndpointRequest struct {
		Cluster   string
		IpAddress string
		Port      uint32
	}

	decoder := json.NewDecoder(r.Body)

	req := &addEndpointRequest{}
	err := decoder.Decode(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = api.manager.AddEndpoint(req.Cluster, req.IpAddress, req.Port)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (api *apiServer) removeEndpoint(w http.ResponseWriter, r *http.Request) {
	type removeEndpointRequest struct {
		Cluster   string
		IpAddress string
		Port      uint32
	}

	decoder := json.NewDecoder(r.Body)

	req := &removeEndpointRequest{}
	err := decoder.Decode(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = api.manager.RemoveEndpoint(req.Cluster, req.IpAddress, req.Port)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (api *apiServer) commitChanges(w http.ResponseWriter, r *http.Request) {
	err := api.manager.Commit()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
