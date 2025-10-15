package boot

import (
	"time"

	"github.com/elysiandb/elysian-gate/internal/balancer"
)

func BootSyncer() {
	go syncSlavesRoutine()
}

func syncSlavesRoutine() {
	for {
		balancer.SyncSlaves()
		time.Sleep(100 * time.Millisecond)
	}
}
