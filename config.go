package main

import (
	"encoding/json"
	"os"
	"time"
)

type config struct {
	OpvaultPath   string        `json:"opvault_path"`
	ProfileName   string        `json:"profile_name"`
	HTTPPort      int           `json:"http_port"`
	UnlockTimeout time.Duration `json:"unlock_timeout"`
	LogPath       string        `json:"log_path"`
}

func parseConfig(path string) (config, error) {
	var cfg config
	file, err := os.Open(path)
	if err != nil {
		return cfg, err
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&cfg)
	if err != nil {
		return cfg, err
	}

	// Set sane(?) defaults
	if len(cfg.LogPath) == 0 {
		cfg.LogPath = "/var/log/onepaq/onepaq.log"
	}
	if cfg.HTTPPort == 0 {
		cfg.HTTPPort = 8080
	}
	if int(cfg.UnlockTimeout) == 0 {
		cfg.UnlockTimeout = 600 * time.Second
	}
	if len(cfg.ProfileName) == 0 {
		cfg.ProfileName = "default"
	}
	return cfg, nil
}
