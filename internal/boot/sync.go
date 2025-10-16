package boot

import (
	"time"

	"github.com/elysiandb/elysian-gate/internal/balancer"
	"github.com/elysiandb/elysian-gate/internal/configuration"
	"github.com/elysiandb/elysian-gate/internal/nodes"
	"github.com/elysiandb/elysian-gate/internal/replication"
)

func BootSyncer() {
	initSlavesReplication()
	go syncSlavesRoutine()
}

func syncSlavesRoutine() {
	for {
		balancer.SyncSlaves()
		time.Sleep(time.Duration(configuration.Config.Gateway.SynchronizationInterval) * time.Second)
	}
}

func initSlavesReplication() {
	master := nodes.GetMasterNode()
	if master == nil {
		return
	}
	for i := range nodes.ElysianCluster.Nodes {
		n := &nodes.ElysianCluster.Nodes[i]
		if n.Role == "slave" && !n.Ready {
			if err := replication.ReplicateMasterToNode(master, n); err == nil {
				n.Ready = true
			}
		}
	}
}
