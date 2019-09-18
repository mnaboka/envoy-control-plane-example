package envoy

import core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"

const defaultNodeHash = "default"

type simpleNodeHash string

func (s simpleNodeHash) ID(node *core.Node) string {
	return string(s)
}
