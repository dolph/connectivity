package main

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	StatsdHost     string `yaml:"statsd_host"`
	StatsdPort     int    `yaml:"statsd_port"`
	StatsdProtocol string `yaml:"statsd_protocol"`
	URLs           []Url
}

type Url struct {
	Label string `yaml:"-"`
	Url   string `yaml:"-"`
}

func (u Url) String() string {
	return u.Label
}

var ConfigPaths = [6]string{
	"connectivity.yml",
	"connectivity.yaml",
	"~/.connectivity.yml",
	"~/.connectivity.yaml",
	"/etc/connectivity.yml",
	"/etc/connectivity.yaml"}

func FindConfig() (string, error) {
	for _, path := range ConfigPaths {
		path = expandHomePath(path)
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}
	return "", errors.New("Failed to locate a config file: ./connectivity.yml ~/.connectivity.yml or /etc/connectivity.yml")
}

func expandHomePath(path string) string {
	if path == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return home
	}

	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}

	return path
}

func LoadConfig(path string) *Config {
	if path == "" {
		return &Config{}
	}

	f, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("Failed to open config file (%s): %v", path, err)
	}

	log.Printf("Loading config from %s", path)

	var configMap map[string]string
	err = yaml.Unmarshal(f, &configMap)
	if err != nil {
		log.Fatalf("Failed to parse YAML config file (%s): %v", path, err)
	}

	var cfg Config

	// Extract the URL labels & values from the struct
	for k, v := range configMap {
		cfg.URLs = append(cfg.URLs, Url{Label: k, Url: v})
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
