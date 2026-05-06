package proxyruntime

import (
	"fmt"

	"api-v2ray/internal/model"
	"api-v2ray/internal/xrayruntime"
)

type Endpoint struct {
	Scheme string
	Host   string
	Port   int
}

type Registry struct {
	nodes map[string]Endpoint
}

func New(nodes []model.ProxyNode) *Registry {
	r := &Registry{nodes: map[string]Endpoint{}}
	for _, n := range nodes {
		r.nodes[n.ID] = Endpoint{
			Scheme: n.Scheme,
			Host:   n.Host,
			Port:   n.Port,
		}
	}
	return r
}

func NewFromGenerated(instances []xrayruntime.GeneratedInstance) *Registry {
	r := &Registry{nodes: map[string]Endpoint{}}
	for _, inst := range instances {
		r.nodes[inst.NodeID] = Endpoint{
			Scheme: "socks5",
			Host:   inst.ListenHost,
			Port:   inst.ListenPort,
		}
	}
	return r
}

func (r *Registry) Get(nodeID string) (*Endpoint, error) {
	e, ok := r.nodes[nodeID]
	if !ok {
		return nil, fmt.Errorf("proxy runtime endpoint not found: %s", nodeID)
	}
	return &e, nil
}
