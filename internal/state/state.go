package state

import (
	"sync"

	"github.com/elysiandb/elysian-gate/internal/global"
)

var (
	slaveState = struct {
		sync.Mutex
		fresh   map[string]bool
		syncing map[string]bool
	}{
		fresh:   map[string]bool{},
		syncing: map[string]bool{},
	}
)

func SetSlaveAsFresh(n *global.Node) {
	slaveState.Lock()
	slaveState.fresh[n.Name] = true
	slaveState.Unlock()
}

func MarkSlaveSyncing(name string, val bool) {
	slaveState.Lock()
	slaveState.syncing[name] = val
	slaveState.Unlock()
}

func IsSlaveSyncing(name string) bool {
	slaveState.Lock()
	defer slaveState.Unlock()
	return slaveState.syncing[name]
}

func IsSlaveFresh(name string) bool {
	slaveState.Lock()
	defer slaveState.Unlock()
	return slaveState.fresh[name]
}

func MarkAllSlavesDirty() {
	slaveState.Lock()
	defer slaveState.Unlock()
	for k := range slaveState.fresh {
		slaveState.fresh[k] = false
	}
}
