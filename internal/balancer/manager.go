package balancer

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/elysiandb/elysian-gate/internal/forward"
	"github.com/elysiandb/elysian-gate/internal/global"
	"github.com/elysiandb/elysian-gate/internal/logger"
	"github.com/elysiandb/elysian-gate/internal/nodes"
	"github.com/elysiandb/elysian-gate/internal/state"
)

type Operation struct {
	Method  string
	Path    string
	Payload string
	Seq     int64
}

var (
	lastSeq    int64
	pendingOps []Operation
	mu         sync.Mutex
)

func SendWriteRequestToMaster(method string, path string, payload string) (int, string, error) {
	master := getMaster()
	if master == nil {
		logger.Error("no master node available for write")
		return 0, "", fmt.Errorf("no master node available for write")
	}

	state.MarkAllSlavesDirty()

	url := fmt.Sprintf("http://%s:%d%s", master.HTTP.Host, master.HTTP.Port, path)
	status, body, err := forward.ForwardRequest(method, url, payload)
	if err != nil || status >= 300 {
		logger.Error(fmt.Sprintf("write to master failed: %v", err))
		return status, body, err
	}

	finalPayload := payload
	if method == "POST" && body != "" {
		finalPayload = body
	}

	mu.Lock()
	op := Operation{
		Method:  method,
		Path:    path,
		Payload: finalPayload,
		Seq:     atomic.AddInt64(&lastSeq, 1),
	}
	pendingOps = append(pendingOps, op)
	mu.Unlock()

	return status, body, nil
}

func GetReadRequestNode() *global.Node {
	mu.Lock()
	hasPending := len(pendingOps) > 0
	mu.Unlock()

	if hasPending {
		return getMaster()
	}

	slaves := getFreshReadySlaves()
	if len(slaves) == 0 {
		logger.Info("no fresh slaves, read from master")
		return getMaster()
	}

	idx := rand.Intn(len(slaves))
	chosen := &slaves[idx]
	logger.Info(fmt.Sprintf("read from slave %s", chosen.Name))
	return chosen
}

func getFreshReadySlaves() []global.Node {
	res := []global.Node{}
	for _, n := range nodes.ElysianCluster.Nodes {
		if n.Role != "slave" || !n.Ready {
			continue
		}
		if !state.IsSlaveFresh(n.Name) {
			state.SetSlaveAsFresh(&n)
		}
		if !state.IsSlaveSyncing(n.Name) {
			res = append(res, n)
		}
	}
	return res
}

func SyncSlaves() {
	mu.Lock()
	ops := append([]Operation(nil), pendingOps...)
	mu.Unlock()
	if len(ops) == 0 {
		return
	}

	var wg sync.WaitGroup
	allSynced := true
	var allMu sync.Mutex

	for i := range nodes.ElysianCluster.Nodes {
		n := &nodes.ElysianCluster.Nodes[i]
		if n.Role != "master" && n.Ready {
			if state.IsSlaveSyncing(n.Name) {
				continue
			}
			wg.Add(1)
			state.MarkSlaveSyncing(n.Name, true)
			go func(nn *global.Node) {
				defer wg.Done()
				defer state.MarkSlaveSyncing(nn.Name, false)
				ok := applyOpsToSlave(nn, ops)
				allMu.Lock()
				if ok {
					state.SetSlaveAsFresh(nn)
				} else {
					allSynced = false
				}
				allMu.Unlock()
			}(n)
		}
	}
	wg.Wait()

	if allSynced {
		mu.Lock()
		pendingOps = nil
		mu.Unlock()
	}
}

func applyOpsToSlave(nn *global.Node, ops []Operation) bool {
	for _, op := range ops {
		url := fmt.Sprintf("http://%s:%d%s", nn.HTTP.Host, nn.HTTP.Port, op.Path)
		status, _, err := forward.ForwardRequest(op.Method, url, op.Payload)
		if err != nil || status >= 300 {
			logger.Error(fmt.Sprintf("sync failed on slave %s: %v", nn.Name, err))
			return false
		}
	}
	return true
}

func getMaster() *global.Node {
	for i := range nodes.ElysianCluster.Nodes {
		if nodes.ElysianCluster.Nodes[i].Role == "master" {
			return &nodes.ElysianCluster.Nodes[i]
		}
	}
	return nil
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
