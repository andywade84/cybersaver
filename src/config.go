package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type appConfig struct {
	Port int `json:"port"`
}

func configPath() string {
	exe, err := os.Executable()
	if err != nil {
		return "config.json"
	}
	return filepath.Join(filepath.Dir(exe), "config.json")
}

func loadConfig() appConfig {
	path := configPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return appConfig{}
	}
	var cfg appConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return appConfig{}
	}
	return cfg
}

func saveConfig(cfg appConfig) error {
	if cfg.Port == 0 {
		cfg.Port = defaultPort
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath(), data, 0o644)
}

func configPort(cfg appConfig) int {
	if cfg.Port <= 0 || cfg.Port > 65535 {
		return defaultPort
	}
	return cfg.Port
}

func requirePort(cfg appConfig) appConfig {
	if cfg.Port != 0 {
		return cfg
	}
	port := promptPort()
	cfg.Port = port
	if err := saveConfig(cfg); err != nil {
		fmt.Printf("could not save config: %v\n", err)
	}
	return cfg
}
