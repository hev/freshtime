package config

import (
	"path/filepath"
	"testing"
)

func TestPath(t *testing.T) {
	p := Path()
	if filepath.Base(p) != "config.json" {
		t.Errorf("expected config.json, got %s", filepath.Base(p))
	}
	if filepath.Base(filepath.Dir(p)) != "freshtime" {
		t.Errorf("expected freshtime dir, got %s", filepath.Base(filepath.Dir(p)))
	}
}

func TestSaveAndLoad(t *testing.T) {
	// Use a temp dir to avoid touching real config
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	cfg := &Config{
		AccessToken:     "test-token",
		RefreshToken:    "test-refresh",
		AccountID:       "123456",
		BusinessID:      42,
		ClientRates:     map[string]string{"client-1": "150.00"},
		DefaultCurrency: "USD",
	}

	if err := Save(cfg); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.AccessToken != cfg.AccessToken {
		t.Errorf("AccessToken: got %q, want %q", loaded.AccessToken, cfg.AccessToken)
	}
	if loaded.RefreshToken != cfg.RefreshToken {
		t.Errorf("RefreshToken: got %q, want %q", loaded.RefreshToken, cfg.RefreshToken)
	}
	if loaded.AccountID != cfg.AccountID {
		t.Errorf("AccountID: got %q, want %q", loaded.AccountID, cfg.AccountID)
	}
	if loaded.BusinessID != cfg.BusinessID {
		t.Errorf("BusinessID: got %d, want %d", loaded.BusinessID, cfg.BusinessID)
	}
	if loaded.DefaultCurrency != cfg.DefaultCurrency {
		t.Errorf("DefaultCurrency: got %q, want %q", loaded.DefaultCurrency, cfg.DefaultCurrency)
	}
	if loaded.ClientRates["client-1"] != "150.00" {
		t.Errorf("ClientRates[client-1]: got %q, want %q", loaded.ClientRates["client-1"], "150.00")
	}
}

func TestLoadMissingFile(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	_, err := Load()
	if err == nil {
		t.Error("expected error for missing config, got nil")
	}
}
