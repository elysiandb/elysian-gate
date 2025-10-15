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
)

type Transport struct {
	Host string
	Port int
	Up   bool
}

type Node struct {
	ID   int
	HTTP Transport
	TCP  Transport
}

type Cluster struct {
	Nodes []Node
	mu    sync.Mutex
}

var ElysianCluster *Cluster

func Init() {
	ElysianCluster = &Cluster{}
	cfg := configuration.Config
	for i, path := range cfg.Nodes {
		elyCfg, err := configuration.ReadElysianConfig(path)
		if err != nil {
			continue
		}
		ElysianCluster.Nodes = append(ElysianCluster.Nodes, Node{
			ID: i + 1,
			HTTP: Transport{
				Host: elyCfg.Server.HTTP.Host,
				Port: elyCfg.Server.HTTP.Port,
				Up:   false,
			},
			TCP: Transport{
				Host: elyCfg.Server.TCP.Host,
				Port: elyCfg.Server.TCP.Port,
				Up:   false,
			},
		})
	}

	if cfg.Gateway.StartsNodes {
		fmt.Printf("Starting %d ElysianDB nodes...\n", len(cfg.Nodes))
		for i, path := range cfg.Nodes {
			bin := filepath.Join("elysiandb", "bin", "elysiandb")
			cmd := exec.Command(bin, "--config", path)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Start(); err != nil {
				fmt.Printf("Failed to start node %d: %v\n", i+1, err)
				continue
			}
			fmt.Printf(" → Node %d started with config: %s\n", i+1, path)
			time.Sleep(200 * time.Millisecond)
		}
		fmt.Println("\nAll nodes are up and running!")
	}
}

func (c *Cluster) monitor() {
	for {
		changed := false
		output := ""
		c.mu.Lock()
		for i := range c.Nodes {
			httpUp := pingHTTP(c.Nodes[i].HTTP.Host, c.Nodes[i].HTTP.Port)
			tcpUp := pingTCP(c.Nodes[i].TCP.Host, c.Nodes[i].TCP.Port)
			if c.Nodes[i].HTTP.Up != httpUp || c.Nodes[i].TCP.Up != tcpUp {
				c.Nodes[i].HTTP.Up = httpUp
				c.Nodes[i].TCP.Up = tcpUp
				changed = true
			}
			httpState := "🔴 HTTP down"
			tcpState := "🔴 TCP down"
			if httpUp {
				httpState = "🟢 HTTP up"
			}
			if tcpUp {
				tcpState = "🟢 TCP up"
			}
			output += fmt.Sprintf("Node %d [HTTP %s:%d | TCP %s:%d] : %s | %s\n",
				c.Nodes[i].ID,
				c.Nodes[i].HTTP.Host, c.Nodes[i].HTTP.Port,
				c.Nodes[i].TCP.Host, c.Nodes[i].TCP.Port,
				httpState, tcpState)
		}
		c.mu.Unlock()
		if changed {
			fmt.Println(time.Now().Format("15:04:05"))
			fmt.Print(output)
			fmt.Println("───────────────────────────────────────────────")
		}
		time.Sleep(2 * time.Second)
	}
}

func (c *Cluster) upNodes() []Node {
	c.mu.Lock()
	defer c.mu.Unlock()
	var upList []Node
	for _, n := range c.Nodes {
		if n.HTTP.Up || n.TCP.Up {
			upList = append(upList, n)
		}
	}
	return upList
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
