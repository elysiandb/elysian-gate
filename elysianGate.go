package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/elysiandb/elysian-gate/internal/boot"
	"github.com/elysiandb/elysian-gate/internal/configuration"
	"github.com/elysiandb/elysian-gate/internal/nodes"
)

func main() {
	clear := flag.Bool("clear", false, "Clear all data before starting")
	configFile := flag.String("config", "elysiangate.yaml", "Path to gateway config file")
	flag.Parse()

	fmt.Println("\033[1;36m╔══════════════════════════════════════════════╗")
	fmt.Println("║               ElysianGate Launcher           ║")
	fmt.Println("╚══════════════════════════════════════════════╝\033[0m")

	if *clear {
		fmt.Println("Clearing previous data...")
		os.RemoveAll("/tmp/elysian*")
		time.Sleep(300 * time.Millisecond)
	}

	configuration.LoadConfig(configFile)

	nodes.Init()

	fmt.Println("───────────────────────────────────────────────")
	fmt.Println(" Gateway is ready to orchestrate the cluster  ")
	fmt.Println("───────────────────────────────────────────────")

	nodes.ElysianCluster.StartMonitoring()

	boot.InitHTTP()

	for {
		time.Sleep(time.Hour)
	}
}
