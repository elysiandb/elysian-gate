package balancer

import (
	"encoding/json"
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

func SendReadRequest(path string, query string) (int, []byte, error) {
	nodes := GetReadRequestNodes()
	if len(nodes) == 0 {
		return 503, []byte(`{"error":"no available node"}`), fmt.Errorf("no available node")
	}

	for _, node := range nodes {
		logger.Info(fmt.Sprintf("trying read from node %s", node.Name))
		url := fmt.Sprintf("http://%s:%d%s", node.HTTP.Host, node.HTTP.Port, path)
		if query != "" {
			url += "?" + query
		}

		status, body, err := forward.ForwardRequest("GET", url, "")
		if err != nil || status >= 300 {
			logger.Error(fmt.Sprintf("read from node %s failed: %v", node.Name, err))
			continue
		}

		var formatted any
		if json.Unmarshal([]byte(body), &formatted) == nil {
			data, _ := json.MarshalIndent(formatted, "", "  ")
			return status, data, nil
		}

		return status, []byte(body), nil
	}

	return 502, []byte(`{"error":"all nodes failed"}`), fmt.Errorf("all nodes failed")
}

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

func GetReadRequestNodes() []*global.Node {
	mu.Lock()
	hasPending := len(pendingOps) > 0
	mu.Unlock()

	nodesList := []*global.Node{}

	if hasPending {
		master := getMaster()
		if master != nil {
			nodesList = append(nodesList, master)
		}
		return nodesList
	}

	slaves := getFreshReadySlaves()
	if len(slaves) > 0 {
		rand.Shuffle(len(slaves), func(i, j int) { slaves[i], slaves[j] = slaves[j], slaves[i] })
		for i := range slaves {
			nodesList = append(nodesList, &slaves[i])
		}
	}

	master := getMaster()
	if master != nil {
		nodesList = append(nodesList, master)
	}

	return nodesList
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
