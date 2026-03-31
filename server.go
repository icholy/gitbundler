package main

import (
	"net/http"
	"path/filepath"
	"strings"
)

type Server struct {
	DataDir string
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Sanitize path to prevent directory traversal.
	clean := filepath.Clean(r.URL.Path)
	if !strings.HasSuffix(clean, ".bundle") {
		http.NotFound(w, r)
		return
	}
	bundlePath := filepath.Join(s.DataDir, clean)
	http.ServeFile(w, r, bundlePath)
}
