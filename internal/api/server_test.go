package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"codexd/internal/config"
	"codexd/internal/git"
	"codexd/internal/sessions"
	"codexd/internal/store"
)

type apiFakeTmux struct {
	exists map[string]bool
	output map[string]string
	sent   []string
}

func newAPIFakeTmux() *apiFakeTmux {
	return &apiFakeTmux{exists: map[string]bool{}, output: map[string]string{}}
}

func (f *apiFakeTmux) Start(tmuxName, repoPath, codexCommand string) error {
	f.exists[tmuxName] = true
	return nil
}

func (f *apiFakeTmux) Exists(tmuxName string) bool {
	return f.exists[tmuxName]
}

func (f *apiFakeTmux) Capture(tmuxName string, lines int) (string, error) {
	return f.output[tmuxName], nil
}

func (f *apiFakeTmux) Send(tmuxName, text string) error {
	f.sent = append(f.sent, text)
	return nil
}

func (f *apiFakeTmux) Kill(tmuxName string) error {
	f.exists[tmuxName] = false
	return nil
}

func TestAuthRequiredExceptHealth(t *testing.T) {
	handler, _ := testServer(t)

	health := httptest.NewRecorder()
	handler.ServeHTTP(health, httptest.NewRequest(http.MethodGet, "/api/health", nil))
	if health.Code != http.StatusOK {
		t.Fatalf("health status = %d", health.Code)
	}

	protected := httptest.NewRecorder()
	handler.ServeHTTP(protected, httptest.NewRequest(http.MethodGet, "/api/sessions", nil))
	if protected.Code != http.StatusUnauthorized {
		t.Fatalf("protected status = %d", protected.Code)
	}
}

func TestSessionLifecycle(t *testing.T) {
	handler, tmux := testServer(t)
	repo := t.TempDir()
	body, _ := json.Marshal(map[string]string{"name": "Demo Repo", "repoPath": repo})

	create := authedRequest(http.MethodPost, "/api/sessions", bytes.NewReader(body))
	createRec := httptest.NewRecorder()
	handler.ServeHTTP(createRec, create)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("create status = %d body=%s", createRec.Code, createRec.Body.String())
	}

	tmux.output["codexd-demo-repo"] = "continue?"
	listRec := httptest.NewRecorder()
	handler.ServeHTTP(listRec, authedRequest(http.MethodGet, "/api/sessions", nil))
	if listRec.Code != http.StatusOK {
		t.Fatalf("list status = %d", listRec.Code)
	}
	if !bytes.Contains(listRec.Body.Bytes(), []byte("needs_input")) {
		t.Fatalf("expected needs_input status: %s", listRec.Body.String())
	}

	inputBody, _ := json.Marshal(map[string]string{"text": "fix tests"})
	inputRec := httptest.NewRecorder()
	handler.ServeHTTP(inputRec, authedRequest(http.MethodPost, "/api/sessions/demo-repo/input", bytes.NewReader(inputBody)))
	if inputRec.Code != http.StatusNoContent {
		t.Fatalf("input status = %d body=%s", inputRec.Code, inputRec.Body.String())
	}
	if len(tmux.sent) != 1 || tmux.sent[0] != "fix tests" {
		t.Fatalf("sent = %#v", tmux.sent)
	}

	deleteRec := httptest.NewRecorder()
	handler.ServeHTTP(deleteRec, authedRequest(http.MethodDelete, "/api/sessions/demo-repo", nil))
	if deleteRec.Code != http.StatusNoContent {
		t.Fatalf("delete status = %d body=%s", deleteRec.Code, deleteRec.Body.String())
	}
}

func testServer(t *testing.T) (http.Handler, *apiFakeTmux) {
	t.Helper()
	st, err := store.NewJSONStore(filepath.Join(t.TempDir(), "state.json"))
	if err != nil {
		t.Fatalf("NewJSONStore() error = %v", err)
	}
	tmux := newAPIFakeTmux()
	manager := sessions.NewManager(st, tmux, config.Config{
		CodexCommand: "codex",
		TmuxPrefix:   "codexd-",
	})
	return NewServer(manager, git.Client{}, "secret", ""), tmux
}

func authedRequest(method, path string, body *bytes.Reader) *http.Request {
	if body == nil {
		body = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, body)
	req.Header.Set("Authorization", "Bearer secret")
	req.Header.Set("Content-Type", "application/json")
	return req
}
