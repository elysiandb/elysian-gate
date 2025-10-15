package balancer

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/elysiandb/elysian-gate/internal/forward"
	"github.com/elysiandb/elysian-gate/internal/nodes"
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
	freshMu    sync.Mutex
	slaveFresh = map[string]bool{}
)

func SendWriteRequestToMaster(method string, path string, payload string) (int, string, error) {
	master := getMaster()
	if master == nil {
		return 0, "", fmt.Errorf("no master node available for write")
	}

	url := fmt.Sprintf("http://%s:%d%s", master.HTTP.Host, master.HTTP.Port, path)
	status, body, err := forward.ForwardRequest(method, url, payload)
	if err != nil || status >= 300 {
		return status, body, err
	}

	finalPayload := payload
	if method == "POST" && body != "" {
		finalPayload = body
	}

	stackOperation(Operation{
		Method:  method,
		Path:    path,
		Payload: finalPayload,
	})

	SetSlavesAsDirty()

	return status, body, nil
}

func GetReadRequestNode() *nodes.Node {
	mu.Lock()
	defer mu.Unlock()
	if len(pendingOps) > 0 {
		return getMaster()
	}
	slaves := getFreshSlaves()
	if len(slaves) == 0 {
		return getMaster()
	}
	idx := rand.Intn(len(slaves))
	return &slaves[idx]
}

func stackOperation(op Operation) {
	mu.Lock()
	op.Seq = atomic.AddInt64(&lastSeq, 1)
	pendingOps = append(pendingOps, op)
	mu.Unlock()
}

func SetSlavesAsDirty() {
	freshMu.Lock()
	for i := range nodes.ElysianCluster.Nodes {
		if nodes.ElysianCluster.Nodes[i].Role != "master" {
			slaveFresh[nodes.ElysianCluster.Nodes[i].Name] = false
		}
	}
	freshMu.Unlock()
}

func SyncSlaves() {
	mu.Lock()
	ops := append([]Operation(nil), pendingOps...)
	mu.Unlock()

	if len(ops) == 0 {
		return
	}

	var wg sync.WaitGroup
	allSyncedForAll := true
	var allMu sync.Mutex

	for i := range nodes.ElysianCluster.Nodes {
		n := &nodes.ElysianCluster.Nodes[i]
		if n.Role != "master" {
			wg.Add(1)
			go func(nn *nodes.Node) {
				defer wg.Done()
				ok := applyOpsToSlave(nn, ops)
				allMu.Lock()
				if ok {
					SetSlaveAsFresh(nn)
				} else {
					allSyncedForAll = false
				}
				allMu.Unlock()
			}(n)
		}
	}
	wg.Wait()

	if allSyncedForAll {
		ClearPendingOperations()
	}
}

func applyOpsToSlave(nn *nodes.Node, ops []Operation) bool {
	for _, op := range ops {
		url := fmt.Sprintf("http://%s:%d%s", nn.HTTP.Host, nn.HTTP.Port, op.Path)
		status, _, err := forward.ForwardRequest(op.Method, url, op.Payload)
		if err != nil || status >= 300 {
			return false
		}
	}
	return true
}

func ClearPendingOperations() {
	mu.Lock()
	pendingOps = nil
	mu.Unlock()
}

func SetSlaveAsFresh(n *nodes.Node) {
	freshMu.Lock()
	slaveFresh[n.Name] = true
	freshMu.Unlock()
}

func getMaster() *nodes.Node {
	for i := range nodes.ElysianCluster.Nodes {
		if nodes.ElysianCluster.Nodes[i].Role == "master" {
			return &nodes.ElysianCluster.Nodes[i]
		}
	}
	return nil
}

func getFreshSlaves() []nodes.Node {
	freshMu.Lock()
	defer freshMu.Unlock()
	slaves := []nodes.Node{}
	for _, n := range nodes.ElysianCluster.Nodes {
		if n.Role == "slave" && slaveFresh[n.Name] {
			slaves = append(slaves, n)
		}
	}
	return slaves
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
