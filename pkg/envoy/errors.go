package envoy

import "errors"

var (
	ErrPrefixAlreadyExists      = errors.New("prefix already exists")
	ErrClusterNotFound          = errors.New("cluster not found")
	ErrClusterAlreadyExists     = errors.New("cluster already exists")
	ErrClusterHasEndpoints      = errors.New("cluster has endpoints")
	ErrEndpointNotFound         = errors.New("endpoint not found")
	ErrEndpointAlreadyExists    = errors.New("endpoint already exists")
	ErrRequiredParameterMissing = errors.New("required parameter missing")
)
