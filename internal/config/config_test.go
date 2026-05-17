package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadCreatesDefaultConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.BindAddr != "127.0.0.1:7777" {
		t.Fatalf("BindAddr = %q", cfg.BindAddr)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected config file: %v", err)
	}
}

func TestEnsureTokenCreatesAndReusesToken(t *testing.T) {
	path := filepath.Join(t.TempDir(), "token")

	first, err := EnsureToken(path)
	if err != nil {
		t.Fatalf("EnsureToken() error = %v", err)
	}
	if len(first) != 64 {
		t.Fatalf("token length = %d", len(first))
	}
	second, err := EnsureToken(path)
	if err != nil {
		t.Fatalf("EnsureToken() second error = %v", err)
	}
	if first != second {
		t.Fatalf("token was not reused")
	}
}
