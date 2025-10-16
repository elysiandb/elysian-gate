package replication_test

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/elysiandb/elysian-gate/internal/global"
	"github.com/elysiandb/elysian-gate/internal/replication"
)

func TestReplicateMasterToNode_Success(t *testing.T) {
	var callCount int

	master := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		switch {
		case r.URL.Path == "/kv/api:entity:types:list":
			json.NewEncoder(w).Encode(map[string]string{"key": "api:entity:types:list", "value": "article"})
		case r.URL.Path == "/api/article":
			if r.Method == "GET" {
				json.NewEncoder(w).Encode([]map[string]interface{}{{"id": "1", "title": "test"}})
			}
		default:
			w.WriteHeader(200)
		}
	}))
	defer master.Close()

	hostM, portStrM, _ := net.SplitHostPort(master.Listener.Addr().String())
	var portM int
	fmt.Sscanf(portStrM, "%d", &portM)
	masterNode := &global.Node{HTTP: global.Transport{Host: hostM, Port: portM}}

	slave := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer slave.Close()

	hostS, portStrS, _ := net.SplitHostPort(slave.Listener.Addr().String())
	var portS int
	fmt.Sscanf(portStrS, "%d", &portS)
	slaveNode := &global.Node{HTTP: global.Transport{Host: hostS, Port: portS}}

	err := replication.ReplicateMasterToNode(masterNode, slaveNode)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if callCount == 0 {
		t.Fatalf("expected calls to master, got %d", callCount)
	}
}

func TestReplicateMasterToNode_Fail_ListTypes(t *testing.T) {
	masterNode := &global.Node{HTTP: global.Transport{Host: "127.0.0.1", Port: 1}}
	slaveNode := &global.Node{HTTP: global.Transport{Host: "127.0.0.1", Port: 8081}}

	err := replication.ReplicateMasterToNode(masterNode, slaveNode)
	if err == nil {
		t.Fatalf("expected error but got nil")
	}
}
