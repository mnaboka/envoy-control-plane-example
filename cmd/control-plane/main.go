package main

import (
	"flag"
	"net"
	"net/http"

	"github.com/mnaboka/envoy-control-plane-example/pkg/api"
	"github.com/mnaboka/envoy-control-plane-example/pkg/envoy"

	envoyapi "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	xds "github.com/envoyproxy/go-control-plane/pkg/server"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	xdsAddress string
	apiAddress string
)

func init() {
	flag.StringVar(&xdsAddress, "xds-address", "0.0.0.0:5678", "Set XDS server address")
	flag.StringVar(&apiAddress, "api-address", "0.0.0.0:8000", "Set server rest api address")
}

func main() {
	flag.Parse()
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)

	manager := envoy.New("xds_cluster", log)

	router := api.New(manager)
	go func() {
		log.Infof("Starting API server on %s\n", apiAddress)
		err := http.ListenAndServe(apiAddress, router)
		if err != nil {
			log.Fatal(err)
		}
	}()

	// Create gRPC controller.
	server := xds.NewServer(manager.Cache(), callbacksInit(log))
	grpcServer := grpc.NewServer()
	lis, err := net.Listen("tcp", xdsAddress)
	if err != nil {
		log.Fatal(err)
	}

	envoyapi.RegisterClusterDiscoveryServiceServer(grpcServer, server)
	envoyapi.RegisterEndpointDiscoveryServiceServer(grpcServer, server)
	envoyapi.RegisterRouteDiscoveryServiceServer(grpcServer, server)
	reflection.Register(grpcServer)

	log.Infof("Starting a server on %s\n", xdsAddress)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("error starting server: %s", err)
	}
}
