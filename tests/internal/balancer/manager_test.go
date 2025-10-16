package balancer_test

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/elysiandb/elysian-gate/internal/balancer"
	"github.com/elysiandb/elysian-gate/internal/global"
	"github.com/elysiandb/elysian-gate/internal/nodes"
)

func mockServer(status int, body string, fail bool) *httptest.Server {
	if fail {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "fail", 500)
		}))
	}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(status)
		w.Write([]byte(body))
	}))
}

func TestSendReadRequest_NoNodes(t *testing.T) {
	nodes.ElysianCluster = &nodes.Cluster{}
	status, _, err := balancer.SendReadRequest("/x", "")
	if status != 503 || err == nil {
		t.Fail()
	}
}

func TestSendReadRequest_AllFail(t *testing.T) {
	s := mockServer(500, "err", true)
	defer s.Close()
	addr := s.Listener.Addr().(*net.TCPAddr)
	nodes.ElysianCluster = &nodes.Cluster{
		Nodes: []global.Node{
			{Name: "m1", Role: "master", Ready: true, HTTP: global.Transport{Host: addr.IP.String(), Port: addr.Port}},
		},
	}
	status, _, err := balancer.SendReadRequest("/api", "")
	if status != 502 || err == nil {
		t.Fail()
	}
}

func TestSendReadRequest_Success(t *testing.T) {
	s := mockServer(200, `{"ok":true}`, false)
	defer s.Close()
	addr := s.Listener.Addr().(*net.TCPAddr)
	nodes.ElysianCluster = &nodes.Cluster{
		Nodes: []global.Node{
			{Name: "m1", Role: "master", Ready: true, HTTP: global.Transport{Host: addr.IP.String(), Port: addr.Port}},
		},
	}
	status, body, err := balancer.SendReadRequest("/api", "")
	if status != 200 || err != nil || len(body) == 0 {
		t.Fail()
	}
}

func TestSendWriteRequestToMaster_NoMaster(t *testing.T) {
	nodes.ElysianCluster = &nodes.Cluster{}
	status, _, err := balancer.SendWriteRequestToMaster("POST", "/api", "{}")
	if status != 0 || err == nil {
		t.Fail()
	}
}

func TestSendWriteRequestToMaster_Success(t *testing.T) {
	s := mockServer(200, `{"id":"1"}`, false)
	defer s.Close()
	addr := s.Listener.Addr().(*net.TCPAddr)
	nodes.ElysianCluster = &nodes.Cluster{
		Nodes: []global.Node{
			{Name: "master", Role: "master", Ready: true, HTTP: global.Transport{Host: addr.IP.String(), Port: addr.Port}},
		},
	}
	status, body, err := balancer.SendWriteRequestToMaster("POST", "/api", "{}")
	if status != 200 || err != nil || body == "" {
		t.Fail()
	}
}

func TestGetReadRequestNodes_NoPending(t *testing.T) {
	nodes.ElysianCluster = &nodes.Cluster{
		Nodes: []global.Node{
			{Name: "m", Role: "master", Ready: true},
		},
	}
	res := balancer.GetReadRequestNodes()
	if len(res) == 0 {
		t.Fail()
	}
}

func TestGetReadRequestNodes_WithSlavesAndMaster(t *testing.T) {
	s := mockServer(200, "ok", false)
	defer s.Close()
	addr := s.Listener.Addr().(*net.TCPAddr)
	master := global.Node{Name: "master", Role: "master", Ready: true, HTTP: global.Transport{Host: addr.IP.String(), Port: addr.Port}}
	slave := global.Node{Name: "slave", Role: "slave", Ready: true, HTTP: global.Transport{Host: addr.IP.String(), Port: addr.Port}}
	nodes.ElysianCluster = &nodes.Cluster{Nodes: []global.Node{master, slave}}
	res := balancer.GetReadRequestNodes()
	if len(res) < 1 {
		t.Fail()
	}
}

func TestSyncSlaves_NoPendingOps(t *testing.T) {
	nodes.ElysianCluster = &nodes.Cluster{}
	balancer.SyncSlaves()
}

func TestSyncSlaves_WithOps(t *testing.T) {
	s := mockServer(200, "ok", false)
	defer s.Close()
	addr := s.Listener.Addr().(*net.TCPAddr)
	master := global.Node{Name: "master", Role: "master", Ready: true}
	slave := global.Node{Name: "slave", Role: "slave", Ready: true, HTTP: global.Transport{Host: addr.IP.String(), Port: addr.Port}}
	nodes.ElysianCluster = &nodes.Cluster{Nodes: []global.Node{master, slave}}
	balancer.SendWriteRequestToMaster("POST", "/api", "{}")
	balancer.SyncSlaves()
}
