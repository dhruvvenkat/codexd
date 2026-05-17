package main

import (
	"flag"
	"log"
	"net/http"
	"path/filepath"

	"codexd/internal/api"
	"codexd/internal/config"
	"codexd/internal/git"
	"codexd/internal/sessions"
	"codexd/internal/store"
)

func main() {
	configPath := flag.String("config", config.DefaultConfigPath(), "path to config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	token, err := config.EnsureToken(cfg.TokenPath)
	if err != nil {
		log.Fatalf("prepare token: %v", err)
	}
	sessionStore, err := store.NewJSONStore(cfg.DataPath)
	if err != nil {
		log.Fatalf("prepare state: %v", err)
	}

	manager := sessions.NewManager(sessionStore, sessions.TmuxClient{}, cfg)
	staticDir := filepath.Join("web", "dist")
	handler := api.NewServer(manager, git.Client{}, token, staticDir)

	log.Printf("codexd listening on http://%s", cfg.BindAddr)
	log.Printf("config: %s", *configPath)
	log.Printf("state: %s", cfg.DataPath)
	log.Printf("token file: %s", cfg.TokenPath)
	if err := http.ListenAndServe(cfg.BindAddr, handler); err != nil {
		log.Fatal(err)
	}
}
