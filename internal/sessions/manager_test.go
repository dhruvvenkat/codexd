package sessions

import (
	"path/filepath"
	"testing"
	"time"

	"codexd/internal/config"
)

type memoryStore struct {
	items map[string]Session
}

func newMemoryStore() *memoryStore {
	return &memoryStore{items: map[string]Session{}}
}

func (s *memoryStore) List() ([]Session, error) {
	items := make([]Session, 0, len(s.items))
	for _, item := range s.items {
		items = append(items, item)
	}
	return items, nil
}

func (s *memoryStore) Get(id string) (Session, bool, error) {
	item, ok := s.items[id]
	return item, ok, nil
}

func (s *memoryStore) Save(session Session) error {
	s.items[session.ID] = session
	return nil
}

func (s *memoryStore) Delete(id string) error {
	delete(s.items, id)
	return nil
}

type fakeTmux struct {
	exists map[string]bool
	output map[string]string
	sent   []string
}

func newFakeTmux() *fakeTmux {
	return &fakeTmux{exists: map[string]bool{}, output: map[string]string{}}
}

func (f *fakeTmux) Start(tmuxName, repoPath, codexCommand string) error {
	f.exists[tmuxName] = true
	return nil
}

func (f *fakeTmux) Exists(tmuxName string) bool {
	return f.exists[tmuxName]
}

func (f *fakeTmux) Capture(tmuxName string, lines int) (string, error) {
	return f.output[tmuxName], nil
}

func (f *fakeTmux) Send(tmuxName, text string) error {
	f.sent = append(f.sent, text)
	return nil
}

func (f *fakeTmux) Kill(tmuxName string) error {
	f.exists[tmuxName] = false
	return nil
}

func TestCreateSanitizesNameAndStartsTmux(t *testing.T) {
	repo := t.TempDir()
	tmux := newFakeTmux()
	manager := NewManager(newMemoryStore(), tmux, config.Config{
		CodexCommand: "codex",
		TmuxPrefix:   "codexd-",
	})
	now := time.Date(2026, 5, 17, 9, 0, 0, 0, time.UTC)
	manager.now = func() time.Time { return now }

	session, err := manager.Create("My Repo!", repo)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if session.ID != "my-repo" {
		t.Fatalf("ID = %q", session.ID)
	}
	if session.RepoPath != filepath.Clean(repo) {
		t.Fatalf("RepoPath = %q", session.RepoPath)
	}
	if !tmux.exists["codexd-my-repo"] {
		t.Fatalf("tmux session was not started")
	}
}

func TestListRefreshesStatus(t *testing.T) {
	repo := t.TempDir()
	tmux := newFakeTmux()
	store := newMemoryStore()
	manager := NewManager(store, tmux, config.Config{CodexCommand: "codex", TmuxPrefix: "codexd-"})

	session, err := manager.Create("needs input", repo)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	tmux.output[session.TmuxName] = "continue?"

	items, err := manager.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(items) != 1 || items[0].Status != StatusNeedsInput {
		t.Fatalf("status = %#v", items)
	}
}
