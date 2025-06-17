package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type ProxyConfig struct {
	Routes []RouteConfig `json:"routes"`
}

type RouteConfig struct {
	Prefix  string    `json:"prefix"`
	Backend []Backend `json:"backends"`
}

type Backend struct {
	URL    string `json:"url"`
	Weight int64  `json:"weight"`
	Health bool   `json:"health"`
}

var (
	Config ProxyConfig
)

const (
	ConfigPath = "./config.json"
)

func LoadConfig() error {
	data, err := os.ReadFile(ConfigPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	err = json.Unmarshal(data, &Config)
	if err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	return nil
}
