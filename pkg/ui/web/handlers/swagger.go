package handlers

import (
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger"
)

// SwaggerHandler serves the Swagger UI
func SwaggerHandler() http.Handler {
	return httpSwagger.Handler(
		httpSwagger.URL("/api/v1/swagger/doc.json"), // The url pointing to API definition
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("list"),
		httpSwagger.DomID("swagger-ui"),
	)
}
