package handlers

import (
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger"
)

// SwaggerConfig contains configuration for the Swagger UI
type SwaggerConfig struct {
	JSONPath string
	DocPath  string
}

// NewSwaggerHandler creates a new Swagger handler with configuration
func NewSwaggerHandler(config SwaggerConfig) *SwaggerHandler {
	return &SwaggerHandler{
		config: config,
	}
}

// SwaggerHandler manages Swagger UI and documentation
type SwaggerHandler struct {
	config SwaggerConfig
}

// UI serves the Swagger UI
func (h *SwaggerHandler) UI() http.Handler {
	return httpSwagger.Handler(
		httpSwagger.URL(h.config.JSONPath),
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("list"),
		httpSwagger.DomID("swagger-ui"),
	)
}

// Doc serves the Swagger JSON documentation
func (h *SwaggerHandler) Doc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	http.ServeFile(w, r, h.config.DocPath)
}
