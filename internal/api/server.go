package api

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"codexd/internal/git"
	"codexd/internal/sessions"
)

type Server struct {
	manager   *sessions.Manager
	git       git.Client
	token     string
	staticDir string
}

func NewServer(manager *sessions.Manager, gitClient git.Client, token, staticDir string) *Server {
	return &Server{
		manager:   manager,
		git:       gitClient,
		token:     token,
		staticDir: staticDir,
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api/") {
		s.serveAPI(w, r)
		return
	}
	s.serveStatic(w, r)
}

func (s *Server) serveAPI(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/health" && !s.authorized(r) {
		writeError(w, http.StatusUnauthorized, "missing or invalid bearer token")
		return
	}

	switch {
	case r.URL.Path == "/api/health" && r.Method == http.MethodGet:
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	case r.URL.Path == "/api/sessions":
		s.handleSessions(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/sessions/"):
		s.handleSession(w, r)
	default:
		writeError(w, http.StatusNotFound, "not found")
	}
}

func (s *Server) serveStatic(w http.ResponseWriter, r *http.Request) {
	if s.staticDir == "" {
		http.NotFound(w, r)
		return
	}
	index := filepath.Join(s.staticDir, "index.html")
	if _, err := os.Stat(index); err != nil {
		http.NotFound(w, r)
		return
	}

	path := filepath.Clean(strings.TrimPrefix(r.URL.Path, "/"))
	if path == "." || path == "/" {
		http.ServeFile(w, r, index)
		return
	}
	full := filepath.Join(s.staticDir, path)
	if !strings.HasPrefix(full, filepath.Clean(s.staticDir)+string(os.PathSeparator)) {
		http.NotFound(w, r)
		return
	}
	if info, err := os.Stat(full); err == nil && !info.IsDir() {
		http.ServeFile(w, r, full)
		return
	}
	http.ServeFile(w, r, index)
}

func statusForError(err error) int {
	switch {
	case errors.Is(err, sessions.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, sessions.ErrConflict):
		return http.StatusConflict
	default:
		return http.StatusBadRequest
	}
}
