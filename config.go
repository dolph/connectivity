package main

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	StatsdHost     string   `yaml:"statsd_host"`
	StatsdPort     int      `yaml:"statsd_port"`
	StatsdProtocol string   `yaml:"statsd_protocol"`
	URLs           []string `yaml:"urls"`
}

var ConfigPaths = [6]string{
	"connectivity.yml",
	"connectivity.yaml",
	"~/.connectivity.yml",
	"~/.connectivity.yaml",
	"/etc/connectivity.yml",
	"/etc/connectivity.yaml"}

func FindConfig() string {
	for _, path := range ConfigPaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return ""
}

func LoadConfig(path string) *Config {
	if path == "" {
		return &Config{}
	}

	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("Failed to open config file (%s): %v", path, err)
	}
	defer f.Close()

	log.Printf("Loading config from %s", path)

	var cfg Config
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cfg)
	if err != nil {
		log.Fatalf("Failed to parse YAML config file (%s): %v", path, err)
	}

	// Apply some default values
	if cfg.StatsdHost == "" {
		cfg.StatsdHost = "127.0.0.1"
	}
	if cfg.StatsdPort == 0 {
		cfg.StatsdPort = 8125
	}
	if cfg.StatsdProtocol == "" {
		cfg.StatsdProtocol = "udp"
	}

	return &cfg
}
