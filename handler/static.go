package handler

import (
	"net/http"
	"os"
	"path/filepath"
)

// StaticHandler serves the built frontend and falls back to index.html so that
// client-side routing keeps working on deep links.
func StaticHandler(dir string) http.HandlerFunc {
	fileServer := http.FileServer(http.Dir(dir))
	return func(w http.ResponseWriter, r *http.Request) {
		name := filepath.Join(dir, filepath.Clean("/"+r.URL.Path))
		if info, err := os.Stat(name); err == nil && !info.IsDir() {
			fileServer.ServeHTTP(w, r)
			return
		}
		http.ServeFile(w, r, filepath.Join(dir, "index.html"))
	}
}
