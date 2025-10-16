package nodes_test

import (
	"testing"
	"time"

	"github.com/elysiandb/elysian-gate/internal/configuration"
	"github.com/elysiandb/elysian-gate/internal/global"
	"github.com/elysiandb/elysian-gate/internal/nodes"
)

func TestInit(t *testing.T) {
	configuration.Config = configuration.ElysianGateConfig{
		Nodes: map[string]configuration.Node{
			"master": {
				Role: "master",
				HTTP: configuration.Transport{Host: "127.0.0.1", Port: 8080, Enabled: true},
				TCP:  configuration.Transport{Host: "127.0.0.1", Port: 9090, Enabled: true},
			},
			"slave": {
				Role: "slave",
				HTTP: configuration.Transport{Host: "127.0.0.1", Port: 8081, Enabled: true},
				TCP:  configuration.Transport{Host: "127.0.0.1", Port: 9091, Enabled: true},
			},
		},
	}
	configuration.Config.Gateway.StartsNodes = false

	nodes.Init()

	if nodes.ElysianCluster == nil || len(nodes.ElysianCluster.Nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %v", nodes.ElysianCluster)
	}

	foundMaster := false
	for _, n := range nodes.ElysianCluster.Nodes {
		switch n.Role {
		case "master":
			foundMaster = true
			if !n.Ready {
				t.Fatalf("master node should be ready")
			}
		case "slave":
			if n.Ready {
				t.Fatalf("slave node should not be ready initially")
			}
		}
	}
	if !foundMaster {
		t.Fatalf("no master node initialized")
	}
}

func TestGetMasterNode(t *testing.T) {
	nodes.ElysianCluster = &nodes.Cluster{
		Nodes: []global.Node{
			{Name: "slave", Role: "slave"},
			{Name: "master", Role: "master"},
		},
	}
	m := nodes.GetMasterNode()
	if m == nil || m.Role != "master" {
		t.Fatalf("expected master node, got %#v", m)
	}
}

func TestStartMonitoring(t *testing.T) {
	c := &nodes.Cluster{}
	c.StartMonitoring()
	time.Sleep(100 * time.Millisecond)
}
