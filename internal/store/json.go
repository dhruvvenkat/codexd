package store

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"codexd/internal/sessions"
)

type JSONStore struct {
	path string
	mu   sync.Mutex
}

type state struct {
	Sessions []sessions.Session `json:"sessions"`
}

func NewJSONStore(path string) (*JSONStore, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, err
	}
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		if err := writeState(path, state{Sessions: []sessions.Session{}}); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	return &JSONStore{path: path}, nil
}

func (s *JSONStore) List() ([]sessions.Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	st, err := s.read()
	if err != nil {
		return nil, err
	}
	items := append([]sessions.Session{}, st.Sessions...)
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
	return items, nil
}

func (s *JSONStore) Get(id string) (sessions.Session, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	st, err := s.read()
	if err != nil {
		return sessions.Session{}, false, err
	}
	for _, item := range st.Sessions {
		if item.ID == id {
			return item, true, nil
		}
	}
	return sessions.Session{}, false, nil
}

func (s *JSONStore) Save(session sessions.Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	st, err := s.read()
	if err != nil {
		return err
	}
	for i, item := range st.Sessions {
		if item.ID == session.ID {
			st.Sessions[i] = session
			return writeState(s.path, st)
		}
	}
	st.Sessions = append(st.Sessions, session)
	return writeState(s.path, st)
}

func (s *JSONStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	st, err := s.read()
	if err != nil {
		return err
	}
	next := st.Sessions[:0]
	for _, item := range st.Sessions {
		if item.ID != id {
			next = append(next, item)
		}
	}
	st.Sessions = next
	return writeState(s.path, st)
}

func (s *JSONStore) read() (state, error) {
	b, err := os.ReadFile(s.path)
	if err != nil {
		return state{}, err
	}
	if len(b) == 0 {
		return state{}, nil
	}
	var st state
	if err := json.Unmarshal(b, &st); err != nil {
		return state{}, err
	}
	return st, nil
}

func writeState(path string, st state) error {
	b, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, append(b, '\n'), 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
