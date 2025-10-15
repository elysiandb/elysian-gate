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
	Name string
	Role string
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
	i := 0
	for name, nodeCfg := range cfg.Nodes {
		i++
		ElysianCluster.Nodes = append(ElysianCluster.Nodes, Node{
			Name: name,
			Role: nodeCfg.Role,
			HTTP: Transport{
				Host: nodeCfg.HTTP.Host,
				Port: nodeCfg.HTTP.Port,
				Up:   false,
			},
			TCP: Transport{
				Host: nodeCfg.TCP.Host,
				Port: nodeCfg.TCP.Port,
				Up:   false,
			},
		})
	}

	if cfg.Gateway.StartsNodes {
		fmt.Printf("Starting %d ElysianDB nodes...\n", len(cfg.Nodes))
		for _, n := range ElysianCluster.Nodes {
			bin := filepath.Join("elysiandb", "bin", "elysiandb")
			cmd := exec.Command(bin, "--http", fmt.Sprintf("%s:%d", n.HTTP.Host, n.HTTP.Port), "--tcp", fmt.Sprintf("%s:%d", n.TCP.Host, n.TCP.Port))
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Start(); err != nil {
				fmt.Printf("Failed to start node %s: %v\n", n.Name, err)
				continue
			}
			fmt.Printf(" → Node %s started on HTTP %s:%d | TCP %s:%d\n", n.Name, n.HTTP.Host, n.HTTP.Port, n.TCP.Host, n.TCP.Port)
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
			output += fmt.Sprintf("Node %s (%s) [HTTP %s:%d | TCP %s:%d] : %s | %s\n",
				c.Nodes[i].Name,
				c.Nodes[i].Role,
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
