package main

import (
	"context"

	"github.com/envoyproxy/go-control-plane/envoy/api/v2"
	xds "github.com/envoyproxy/go-control-plane/pkg/server"
	"github.com/sirupsen/logrus"
)

type callbacks struct {
	log *logrus.Logger
}

func (cb *callbacks) OnStreamOpen(_ context.Context, id int64, typeURL string) error {
	cb.log.Debugf("OnStreamOpen stream id %d, type url %s", id, typeURL)
	return nil
}

func (cb *callbacks) OnStreamClosed(id int64) {
	cb.log.Debugf("OnStreamClosed stream id %d", id)
}

func (cb *callbacks) OnStreamRequest(id int64, req *envoy_api_v2.DiscoveryRequest) error {
	cb.log.Debugf("OnStreamRequest stream id %d, req %s", id, req)
	return nil
}

func (cb *callbacks) OnStreamResponse(id int64, req *envoy_api_v2.DiscoveryRequest, resp *envoy_api_v2.DiscoveryResponse) {
	cb.log.Debugf("OnStreamResponse stream id %d, req %s, resp %s", id, req, resp)
}

func (cb *callbacks) OnFetchRequest(_ context.Context, req *envoy_api_v2.DiscoveryRequest) error {
	cb.log.Debugf("OnFetchRequest req %s", req)
	return nil
}

func (cb *callbacks) OnFetchResponse(req *envoy_api_v2.DiscoveryRequest, resp *envoy_api_v2.DiscoveryResponse) {
	cb.log.Debugf("OnFetchResponse req %s, resp %s", req, resp)
}

func callbacksInit(log *logrus.Logger) xds.Callbacks {
	return &callbacks{log}
}
