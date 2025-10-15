package main

import (
	"flag"
	"os"
	"time"

	"github.com/elysiandb/elysian-gate/internal/boot"
	"github.com/elysiandb/elysian-gate/internal/configuration"
	"github.com/elysiandb/elysian-gate/internal/logger"
	"github.com/elysiandb/elysian-gate/internal/nodes"
)

func main() {
	clear := flag.Bool("clear", false, "Clear all data before starting")
	configFile := flag.String("config", "elysiangate.yaml", "Path to gateway config file")
	flag.Parse()

	logger.Info("Starting ElysianGate...")

	if *clear {
		logger.Info("Clearing previous data...")
		os.RemoveAll("/tmp/elysian*")
		time.Sleep(300 * time.Millisecond)
	}

	configuration.LoadConfig(configFile)

	nodes.Init()
	boot.BootSyncer()

	logger.Info("───────────────────────────────────────────────")
	logger.Info(" Gateway is ready to orchestrate the cluster  ")
	logger.Info("───────────────────────────────────────────────")
	nodes.ElysianCluster.StartMonitoring()

	boot.InitHTTP()

	for {
		time.Sleep(time.Hour)
	}
}
