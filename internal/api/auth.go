package api

import (
	"crypto/subtle"
	"net/http"
	"strings"
)

func (s *Server) authorized(r *http.Request) bool {
	header := r.Header.Get("Authorization")
	if !strings.HasPrefix(header, "Bearer ") {
		return false
	}
	got := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
	if got == "" || s.token == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(got), []byte(s.token)) == 1
}
