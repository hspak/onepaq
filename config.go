package main

import (
	"encoding/json"
	"os"
	"time"
)

type config struct {
	OpvaultPath   string        `json:"opvault_path"`
	ProfileName   string        `json:"profile_name"`
	HTTPAddr      string        `json:"http_addr"`
	UnlockTimeout time.Duration `json:"unlock_timeout"`
	CertCA        string        `json:"ca_file"`
	CertFile      string        `json:"cert_file"`
	KeyFile       string        `json:"key_file"`
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
	if cfg.HTTPAddr == "" {
		cfg.HTTPAddr = "localhost:8080"
	}
	if int(cfg.UnlockTimeout) == 0 {
		cfg.UnlockTimeout = 600 * time.Second
	}
	if cfg.ProfileName == "" {
		cfg.ProfileName = "default"
	}
	return cfg, nil
}
