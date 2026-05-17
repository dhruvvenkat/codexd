package api

import (
	"encoding/json"
	"net/http"
	"strings"
)

type createSessionRequest struct {
	Name     string `json:"name"`
	RepoPath string `json:"repoPath"`
}

type sendInputRequest struct {
	Text string `json:"text"`
}

func (s *Server) handleSessions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		items, err := s.manager.List()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"sessions": items})
	case http.MethodPost:
		var req createSessionRequest
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		session, err := s.manager.Create(req.Name, req.RepoPath)
		if err != nil {
			writeError(w, statusForError(err), err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, session)
	default:
		w.Header().Set("Allow", "GET, POST")
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) handleSession(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/sessions/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	id := parts[0]

	if len(parts) == 1 {
		switch r.Method {
		case http.MethodGet:
			session, err := s.manager.Get(id)
			if err != nil {
				writeError(w, statusForError(err), err.Error())
				return
			}
			writeJSON(w, http.StatusOK, session)
		case http.MethodDelete:
			if err := s.manager.Kill(id); err != nil {
				writeError(w, statusForError(err), err.Error())
				return
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			w.Header().Set("Allow", "GET, DELETE")
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	switch strings.Join(parts[1:], "/") {
	case "output":
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", "GET")
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		output, err := s.manager.Output(id)
		if err != nil {
			writeError(w, statusForError(err), err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"output": output})
	case "input":
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", "POST")
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		var req sendInputRequest
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if err := s.manager.SendInput(id, req.Text); err != nil {
			writeError(w, statusForError(err), err.Error())
			return
		}
		w.WriteHeader(http.StatusNoContent)
	case "git/status":
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", "GET")
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		session, err := s.manager.Get(id)
		if err != nil {
			writeError(w, statusForError(err), err.Error())
			return
		}
		status, err := s.git.Status(session.RepoPath)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": status})
	case "git/diff":
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", "GET")
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		session, err := s.manager.Get(id)
		if err != nil {
			writeError(w, statusForError(err), err.Error())
			return
		}
		diff, err := s.git.Diff(session.RepoPath)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"diff": diff})
	default:
		writeError(w, http.StatusNotFound, "not found")
	}
}

func decodeJSON(r *http.Request, v any) error {
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
