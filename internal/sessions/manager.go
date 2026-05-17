package sessions

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"codexd/internal/config"
)

var (
	ErrNotFound = errors.New("session not found")
	ErrConflict = errors.New("session already exists")
)

type Store interface {
	List() ([]Session, error)
	Get(id string) (Session, bool, error)
	Save(session Session) error
	Delete(id string) error
}

type Manager struct {
	store Store
	tmux  Tmux
	cfg   config.Config
	now   func() time.Time
}

func NewManager(store Store, tmux Tmux, cfg config.Config) *Manager {
	return &Manager{
		store: store,
		tmux:  tmux,
		cfg:   cfg,
		now:   func() time.Time { return time.Now().UTC() },
	}
}

func (m *Manager) Create(name, repoPath string) (Session, error) {
	id := SanitizeName(name)
	if id == "" {
		return Session{}, fmt.Errorf("session name must contain letters or numbers")
	}
	repoPath = strings.TrimSpace(repoPath)
	if repoPath == "" {
		return Session{}, fmt.Errorf("repoPath is required")
	}
	absPath, err := filepath.Abs(repoPath)
	if err != nil {
		return Session{}, err
	}
	if info, err := os.Stat(absPath); err != nil {
		return Session{}, err
	} else if !info.IsDir() {
		return Session{}, fmt.Errorf("repoPath must be a directory")
	}
	if _, ok, err := m.store.Get(id); err != nil {
		return Session{}, err
	} else if ok {
		return Session{}, ErrConflict
	}

	now := m.now()
	session := Session{
		ID:        id,
		Name:      strings.TrimSpace(name),
		RepoPath:  absPath,
		TmuxName:  m.cfg.TmuxPrefix + id,
		Status:    StatusRunning,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if session.Name == "" {
		session.Name = id
	}
	if err := m.tmux.Start(session.TmuxName, session.RepoPath, m.cfg.CodexCommand); err != nil {
		return Session{}, err
	}
	if err := m.store.Save(session); err != nil {
		_ = m.tmux.Kill(session.TmuxName)
		return Session{}, err
	}
	return session, nil
}

func (m *Manager) List() ([]Session, error) {
	items, err := m.store.List()
	if err != nil {
		return nil, err
	}
	for i := range items {
		items[i] = m.refresh(items[i])
	}
	return items, nil
}

func (m *Manager) Get(id string) (Session, error) {
	session, ok, err := m.store.Get(id)
	if err != nil {
		return Session{}, err
	}
	if !ok {
		return Session{}, ErrNotFound
	}
	return m.refresh(session), nil
}

func (m *Manager) Output(id string) (string, error) {
	session, err := m.Get(id)
	if err != nil {
		return "", err
	}
	if session.Status == StatusStopped {
		return "", nil
	}
	return m.tmux.Capture(session.TmuxName, 200)
}

func (m *Manager) SendInput(id, text string) error {
	text = strings.TrimRight(text, "\r\n")
	if strings.TrimSpace(text) == "" {
		return fmt.Errorf("text is required")
	}
	session, err := m.Get(id)
	if err != nil {
		return err
	}
	if session.Status == StatusStopped {
		return fmt.Errorf("session is stopped")
	}
	if err := m.tmux.Send(session.TmuxName, text); err != nil {
		return err
	}
	session.Status = StatusRunning
	session.UpdatedAt = m.now()
	return m.store.Save(session)
}

func (m *Manager) Kill(id string) error {
	session, ok, err := m.store.Get(id)
	if err != nil {
		return err
	}
	if !ok {
		return ErrNotFound
	}
	if m.tmux.Exists(session.TmuxName) {
		if err := m.tmux.Kill(session.TmuxName); err != nil {
			return err
		}
	}
	return m.store.Delete(id)
}

func (m *Manager) refresh(session Session) Session {
	exists := m.tmux.Exists(session.TmuxName)
	output := ""
	status := StatusStopped
	if exists {
		var err error
		output, err = m.tmux.Capture(session.TmuxName, 200)
		if err != nil {
			status = StatusUnknown
		} else {
			status = DetectStatus(true, output)
		}
	}
	if status != session.Status {
		session.Status = status
		session.UpdatedAt = m.now()
		_ = m.store.Save(session)
	}
	return session
}

func SanitizeName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	var b strings.Builder
	lastDash := false
	for _, r := range name {
		ok := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_'
		if ok {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			b.WriteByte('-')
			lastDash = true
		}
	}
	id := strings.Trim(b.String(), "-_")
	if len(id) > 48 {
		id = strings.TrimRight(id[:48], "-_")
	}
	return id
}
