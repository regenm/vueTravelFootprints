package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func ServeSPA(staticDir string) http.HandlerFunc {
	root := filepath.Clean(staticDir)
	index := filepath.Join(root, "index.html")
	return func(w http.ResponseWriter, r *http.Request) {
		rel := strings.TrimPrefix(filepath.Clean("/"+r.URL.Path), string(os.PathSeparator))
		rel = strings.TrimPrefix(rel, "/")
		target := filepath.Join(root, filepath.FromSlash(rel))
		if target != root && !strings.HasPrefix(target, root+string(os.PathSeparator)) {
			http.NotFound(w, r)
			return
		}
		if info, err := os.Stat(target); err == nil && !info.IsDir() {
			http.ServeFile(w, r, target)
			return
		}
		if _, err := os.Stat(index); err != nil {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, index)
	}
}
