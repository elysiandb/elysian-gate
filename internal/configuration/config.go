package configuration

import (
	"fmt"
	"os"

	"github.com/elysiandb/elysian-gate/internal/logger"
	"gopkg.in/yaml.v3"
)

type Transport struct {
	Enabled bool   `yaml:"enabled"`
	Host    string `yaml:"host"`
	Port    int    `yaml:"port"`
}

type Node struct {
	Role string    `yaml:"role"`
	HTTP Transport `yaml:"http"`
	TCP  Transport `yaml:"tcp"`
}

type ElysianGateConfig struct {
	Nodes   map[string]Node `yaml:"nodes"`
	Gateway struct {
		StartsNodes bool `yaml:"startsNodes"`
		HTTP        struct {
			Host string `yaml:"host"`
			Port int    `yaml:"port"`
		} `yaml:"http"`
		SynchronizationInterval int `yaml:"synchronizationInterval"`
	} `yaml:"gateway"`
}

var Config ElysianGateConfig

func ReadElysianConfig(path string) (ElysianGateConfig, error) {
	var cfg ElysianGateConfig
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func LoadConfig(configFile *string) error {
	data, err := os.ReadFile(*configFile)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to read config file: %v\n", err))
		return err
	}
	if err := yaml.Unmarshal(data, &Config); err != nil {
		logger.Error(fmt.Sprintf("Invalid YAML config: %v\n", err))
		return err
	}
	if len(Config.Nodes) == 0 {
		logger.Error("No nodes defined in config file.")
		return fmt.Errorf("no nodes defined")
	}
	return nil
}
