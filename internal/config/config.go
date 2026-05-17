package config

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	BindAddr     string `json:"bindAddr"`
	DataPath     string `json:"dataPath"`
	TokenPath    string `json:"tokenPath"`
	CodexCommand string `json:"codexCommand"`
	TmuxPrefix   string `json:"tmuxPrefix"`
}

func DefaultConfigPath() string {
	dir, err := os.UserConfigDir()
	if err != nil || dir == "" {
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, ".config")
	}
	return filepath.Join(dir, "codexd", "config.json")
}

func Default() Config {
	home, _ := os.UserHomeDir()
	configDir, err := os.UserConfigDir()
	if err != nil || configDir == "" {
		configDir = filepath.Join(home, ".config")
	}
	return Config{
		BindAddr:     "127.0.0.1:7777",
		DataPath:     filepath.Join(home, ".local", "share", "codexd", "state.json"),
		TokenPath:    filepath.Join(configDir, "codexd", "token"),
		CodexCommand: "codex",
		TmuxPrefix:   "codexd-",
	}
}

func Load(path string) (Config, error) {
	if path == "" {
		path = DefaultConfigPath()
	}

	cfg := Default()
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		if err := writeConfig(path, cfg); err != nil {
			return Config{}, err
		}
		return cfg, nil
	}
	if err != nil {
		return Config{}, err
	}
	if err := json.Unmarshal(b, &cfg); err != nil {
		return Config{}, err
	}
	return fillDefaults(cfg), nil
}

func EnsureToken(path string) (string, error) {
	if path == "" {
		path = Default().TokenPath
	}
	if b, err := os.ReadFile(path); err == nil {
		token := strings.TrimSpace(string(b))
		if token != "" {
			return token, nil
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", err
	}

	token, err := generateToken()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, []byte(token+"\n"), 0o600); err != nil {
		return "", err
	}
	return token, nil
}

func fillDefaults(cfg Config) Config {
	def := Default()
	if cfg.BindAddr == "" {
		cfg.BindAddr = def.BindAddr
	}
	if cfg.DataPath == "" {
		cfg.DataPath = def.DataPath
	}
	if cfg.TokenPath == "" {
		cfg.TokenPath = def.TokenPath
	}
	if cfg.CodexCommand == "" {
		cfg.CodexCommand = def.CodexCommand
	}
	if cfg.TmuxPrefix == "" {
		cfg.TmuxPrefix = def.TmuxPrefix
	}
	return cfg
}

func writeConfig(path string, cfg Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(b, '\n'), 0o600)
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
