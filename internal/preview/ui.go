package preview

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed web/*.html web/*.css web/app.js web/state.js
var webAssets embed.FS

// WebHandler serves the self-contained preview interface. API routes should
// be registered before the root handler; assets require no network access.
func WebHandler() http.Handler {
	assets, err := fs.Sub(webAssets, "web")
	if err != nil {
		panic(err)
	}
	return http.FileServer(http.FS(assets))
}
