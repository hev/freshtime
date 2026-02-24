package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds the freshtime configuration.
type Config struct {
	AccessToken     string            `json:"access_token"`
	RefreshToken    string            `json:"refresh_token,omitempty"`
	AccountID       string            `json:"account_id"`
	BusinessID      int               `json:"business_id"`
	ClientRates     map[string]string `json:"client_rates,omitempty"`
	DefaultCurrency string            `json:"default_currency,omitempty"`
}

func configDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "freshtime")
}

// Path returns the path to the config file.
func Path() string {
	return filepath.Join(configDir(), "config.json")
}

// Load reads and parses the config file.
func Load() (*Config, error) {
	data, err := os.ReadFile(Path())
	if err != nil {
		return nil, fmt.Errorf("config not found. Run `freshtime setup` to configure your token")
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("invalid config file: %w", err)
	}
	return &cfg, nil
}

// Save writes the config to disk.
func Save(cfg *Config) error {
	if err := os.MkdirAll(configDir(), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(Path(), data, 0o644)
}
