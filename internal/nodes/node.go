package nodes

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/elysiandb/elysian-gate/internal/configuration"
	"github.com/elysiandb/elysian-gate/internal/global"
	"github.com/elysiandb/elysian-gate/internal/logger"
	"github.com/elysiandb/elysian-gate/internal/replication"
	"github.com/elysiandb/elysian-gate/internal/state"
)

type Cluster struct {
	Nodes []global.Node
	mu    sync.Mutex
}

var ElysianCluster *Cluster

func Init() {
	ElysianCluster = &Cluster{}
	cfg := configuration.Config
	for name, nodeCfg := range cfg.Nodes {
		n := global.Node{
			Name: name,
			Role: nodeCfg.Role,
			HTTP: global.Transport{
				Host: nodeCfg.HTTP.Host,
				Port: nodeCfg.HTTP.Port,
				Up:   false,
			},
			TCP: global.Transport{
				Host: nodeCfg.TCP.Host,
				Port: nodeCfg.TCP.Port,
				Up:   false,
			},
		}
		if n.Role == "slave" {
			n.Ready = false
		} else {
			n.Ready = true
		}
		ElysianCluster.Nodes = append(ElysianCluster.Nodes, n)
	}

	if cfg.Gateway.StartsNodes {
		logger.Info(fmt.Sprintf("Starting %d ElysianDB nodes...\n", len(cfg.Nodes)))
		for _, n := range ElysianCluster.Nodes {
			bin := filepath.Join("elysiandb", "bin", "elysiandb")
			cmd := exec.Command(bin, "--http", fmt.Sprintf("%s:%d", n.HTTP.Host, n.HTTP.Port), "--tcp", fmt.Sprintf("%s:%d", n.TCP.Host, n.TCP.Port))
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Start(); err != nil {
				logger.Error(fmt.Sprintf("Failed to start node %s: %v\n", n.Name, err))
				continue
			}
			logger.Info(fmt.Sprintf(" → Node %s started on HTTP %s:%d | TCP %s:%d\n", n.Name, n.HTTP.Host, n.HTTP.Port, n.TCP.Host, n.TCP.Port))
			time.Sleep(200 * time.Millisecond)
		}
		logger.Info("\nAll nodes are up and running!")
	}
}

func (c *Cluster) monitor() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	prevSnapshot := ""

	for range ticker.C {
		changed := c.refreshStatuses()
		snapshot := c.clusterSnapshot()

		if changed || snapshot != prevSnapshot {
			clearScreen()
			fmt.Println("\033[1;36m╔══════════════════════════════════════════════╗")
			fmt.Println("║               ElysianGate Launcher           ║")
			fmt.Println("╚══════════════════════════════════════════════╝\033[0m")
			fmt.Println(time.Now().Format("15:04:05"))
			fmt.Print(snapshot)
			prevSnapshot = snapshot
		}
	}
}

func (c *Cluster) refreshStatuses() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	changed := false
	for i := range c.Nodes {
		httpUp := pingHTTP(c.Nodes[i].HTTP.Host, c.Nodes[i].HTTP.Port)
		tcpUp := pingTCP(c.Nodes[i].TCP.Host, c.Nodes[i].TCP.Port)

		prevHTTP := c.Nodes[i].HTTP.Up
		prevTCP := c.Nodes[i].TCP.Up
		prevReady := c.Nodes[i].Ready

		if prevHTTP != httpUp {
			c.Nodes[i].HTTP.Up = httpUp
			changed = true
		}
		if prevTCP != tcpUp {
			c.Nodes[i].TCP.Up = tcpUp
			changed = true
		}

		if !httpUp || !tcpUp {
			if c.Nodes[i].Role == "slave" && c.Nodes[i].Ready {
				c.Nodes[i].Ready = false
				changed = true
			}
		} else {
			if c.Nodes[i].Role == "slave" && !prevReady {
				go resyncSlaveFromMaster(&c.Nodes[i])
			}
		}

		if prevReady != c.Nodes[i].Ready {
			changed = true
		}
	}
	return changed
}

func (c *Cluster) clusterSnapshot() string {
	c.mu.Lock()
	defer c.mu.Unlock()

	var out strings.Builder
	for _, n := range c.Nodes {
		httpState := "🔴 HTTP down"
		if n.HTTP.Up {
			httpState = "🟢 HTTP up"
		}
		tcpState := "🔴 TCP down"
		if n.TCP.Up {
			tcpState = "🟢 TCP up"
		}
		readyState := "🔴 NotReady"
		if n.Ready {
			readyState = "🟢 Ready"
		}
		out.WriteString(fmt.Sprintf(
			"Node %s (%s) [HTTP %s:%d | TCP %s:%d] : %s | %s | %s\n",
			n.Name, n.Role,
			n.HTTP.Host, n.HTTP.Port,
			n.TCP.Host, n.TCP.Port,
			httpState, tcpState, readyState,
		))
	}
	return out.String()
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

func resyncSlaveFromMaster(n *global.Node) {
	master := GetMasterNode()
	if master == nil {
		logger.Error(fmt.Sprintf("No master found, cannot replicate to %s", n.Name))
		return
	}
	if n.Ready {
		return
	}

	logger.Info(fmt.Sprintf("Replicating master → %s ...", n.Name))
	if err := replication.ReplicateMasterToNode(master, n); err != nil {
		logger.Error(fmt.Sprintf("Replication failed for %s: %v", n.Name, err))
		return
	}
	n.Ready = true

	state.SetSlaveAsFresh(n)

	logger.Info(fmt.Sprintf("Node %s replication complete, now marked as Ready & Fresh", n.Name))
}

func GetMasterNode() *global.Node {
	for i := range ElysianCluster.Nodes {
		if ElysianCluster.Nodes[i].Role == "master" {
			return &ElysianCluster.Nodes[i]
		}
	}
	return nil
}

func pingTCP(host string, port int) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), 2*time.Second)
	if err != nil {
		return false
	}
	defer conn.Close()
	conn.Write([]byte("PING\n"))
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	resp, _ := bufio.NewReader(conn).ReadString('\n')
	return strings.TrimSpace(resp) == "PONG"
}

func pingHTTP(host string, port int) bool {
	client := http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(fmt.Sprintf("http://%s:%d/health", host, port))
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200
}

func (c *Cluster) StartMonitoring() {
	go c.monitor()
}
