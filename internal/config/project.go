package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const ProjectConfigFile = ".freshtime.json"

// ProjectConfig holds per-project defaults for time logging.
type ProjectConfig struct {
	ClientID  int `json:"client_id,omitempty"`
	ProjectID int `json:"project_id,omitempty"`
	ServiceID int `json:"service_id,omitempty"`
}

// LoadProjectConfig reads a .freshtime.json from the given directory.
func LoadProjectConfig(dir string) (*ProjectConfig, error) {
	data, err := os.ReadFile(filepath.Join(dir, ProjectConfigFile))
	if err != nil {
		return nil, fmt.Errorf("no %s found in %s", ProjectConfigFile, dir)
	}
	var pc ProjectConfig
	if err := json.Unmarshal(data, &pc); err != nil {
		return nil, fmt.Errorf("invalid %s: %w", ProjectConfigFile, err)
	}
	return &pc, nil
}

// LoadProjectConfigFromCwd reads .freshtime.json from the current working directory.
func LoadProjectConfigFromCwd() (*ProjectConfig, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	return LoadProjectConfig(cwd)
}

// SaveProjectConfig writes a .freshtime.json to the given directory.
func SaveProjectConfig(dir string, pc *ProjectConfig) error {
	data, err := json.MarshalIndent(pc, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(filepath.Join(dir, ProjectConfigFile), data, 0o644)
}
