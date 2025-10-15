package configuration

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type ElysianConfig struct {
	Server struct {
		HTTP struct {
			Enabled bool   `yaml:"enabled"`
			Host    string `yaml:"host"`
			Port    int    `yaml:"port"`
		} `yaml:"http"`
		TCP struct {
			Enabled bool   `yaml:"enabled"`
			Host    string `yaml:"host"`
			Port    int    `yaml:"port"`
		} `yaml:"tcp"`
	} `yaml:"server"`
}

type ElysianGateConfig struct {
	Nodes   []string `yaml:"nodes"`
	Gateway struct {
		StartsNodes bool `yaml:"startsNodes"`
		HTTP        struct {
			Host string `yaml:"host"`
			Port int    `yaml:"port"`
		} `yaml:"http"`
	} `yaml:"gateway"`
}

var Config ElysianGateConfig

func ReadElysianConfig(path string) (ElysianConfig, error) {
	var cfg ElysianConfig
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
		fmt.Printf("Failed to read config file: %v\n", err)
		return err
	}
	if err := yaml.Unmarshal(data, &Config); err != nil {
		fmt.Printf("Invalid YAML config: %v\n", err)
		return err
	}
	if len(Config.Nodes) == 0 {
		fmt.Println("No nodes defined in config file.")
		return fmt.Errorf("no nodes defined")
	}
	return nil
}
