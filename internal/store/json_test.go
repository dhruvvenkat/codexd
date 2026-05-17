package store

import (
	"path/filepath"
	"testing"
	"time"

	"codexd/internal/sessions"
)

func TestJSONStoreSaveListGetDelete(t *testing.T) {
	st, err := NewJSONStore(filepath.Join(t.TempDir(), "state.json"))
	if err != nil {
		t.Fatalf("NewJSONStore() error = %v", err)
	}

	items, err := st.List()
	if err != nil {
		t.Fatalf("initial List() error = %v", err)
	}
	if items == nil || len(items) != 0 {
		t.Fatalf("initial List() = %#v", items)
	}

	session := sessions.Session{
		ID:        "demo",
		Name:      "demo",
		RepoPath:  "/tmp/demo",
		TmuxName:  "codexd-demo",
		Status:    sessions.StatusRunning,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	if err := st.Save(session); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	got, ok, err := st.Get("demo")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if !ok || got.ID != "demo" {
		t.Fatalf("Get() = %#v, %v", got, ok)
	}
	items, err = st.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("List() length = %d", len(items))
	}
	if err := st.Delete("demo"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, ok, err := st.Get("demo"); err != nil || ok {
		t.Fatalf("Get() after delete ok=%v err=%v", ok, err)
	}
}
