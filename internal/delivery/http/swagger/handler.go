package swagger

import (
	_ "embed"
	"net/http"
	"os"
	"path/filepath"
)

//go:embed swagger-ui.html
var swaggerUIHTML []byte

// Handler returns an HTTP handler that serves Swagger UI.
func Handler(swaggerJSONPath string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/swagger/", "/swagger":
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write(swaggerUIHTML)
		case "/swagger/api.swagger.json":
			// Serve the swagger JSON file
			data, err := os.ReadFile(swaggerJSONPath)
			if err != nil {
				http.Error(w, "Swagger spec not found", http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(data)
		default:
			http.NotFound(w, r)
		}
	})
}

// RegisterRoutes registers swagger routes on the given mux.
func RegisterRoutes(mux *http.ServeMux, baseDir string) {
	swaggerPath := filepath.Join(baseDir, "gen", "openapi", "api.swagger.json")
	mux.Handle("/swagger/", Handler(swaggerPath))
	mux.Handle("/swagger", http.RedirectHandler("/swagger/", http.StatusMovedPermanently))
}
