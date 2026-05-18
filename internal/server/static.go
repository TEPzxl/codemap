package server

import (
	"embed"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

//go:embed all:static
var webAssets embed.FS

func webHandler() http.Handler {
	staticFS, err := fs.Sub(webAssets, "static")
	if err != nil {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			writeError(w, http.StatusInternalServerError, "web assets are not available")
		})
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		requestPath := cleanWebPath(r.URL.Path)
		if requestPath == "/" {
			requestPath = "/index.html"
		}

		if !assetExists(staticFS, strings.TrimPrefix(requestPath, "/")) {
			if isStaticAssetRequest(requestPath) {
				http.NotFound(w, r)
				return
			}
			if !assetExists(staticFS, "index.html") {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("<!doctype html><title>codemap</title><main><h1>codemap</h1><p>Web assets are not built. Run <code>make web-build</code> before building the release binary.</p></main>"))
				return
			}
			requestPath = "/index.html"
		}

		http.ServeFileFS(w, r, staticFS, strings.TrimPrefix(requestPath, "/"))
	})
}

func cleanWebPath(rawPath string) string {
	cleaned := path.Clean("/" + rawPath)
	if cleaned == "." {
		return "/"
	}
	return cleaned
}

func assetExists(files fs.FS, name string) bool {
	if name == "" {
		return false
	}
	cleaned := strings.TrimPrefix(path.Clean("/"+name), "/")
	if cleaned == "." || cleaned != name {
		return false
	}

	file, err := files.Open(name)
	if err != nil {
		return false
	}
	_ = file.Close()
	return true
}

func isStaticAssetRequest(requestPath string) bool {
	return strings.HasPrefix(requestPath, "/_next/") || path.Ext(requestPath) != ""
}
