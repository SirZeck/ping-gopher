package web

import (
	"embed"
	"net/http"
)

//go:embed index.html style.css app.js status.html status.js
var webFS embed.FS

// StaticHandler returns an http.Handler that serves embedded static dashboard and public status page files.
func StaticHandler() http.Handler {
	return http.FileServer(http.FS(webFS))
}
