package server

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/golem-base/seqctl/pkg/app"
	"github.com/golem-base/seqctl/pkg/server/handlers"
	slogchi "github.com/samber/slog-chi"
)

//go:embed all:dist
var content embed.FS

// ServerConfig holds the configuration for the web server
type ServerConfig struct {
	Address         string
	Port            int
	RefreshInterval int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	MaxHeaderBytes  int
}

// DefaultServerConfig returns the default server configuration
func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		Address:         "0.0.0.0",
		Port:            8080,
		RefreshInterval: 5,
		ReadTimeout:     15 * time.Second,
		WriteTimeout:    15 * time.Second,
		IdleTimeout:     60 * time.Second,
		MaxHeaderBytes:  1 << 20, // 1 MB
	}
}

// Server represents the web server
type Server struct {
	config     ServerConfig
	app        *app.App
	httpServer *http.Server
	logger     *slog.Logger
}

// NewServer creates a new web server instance
func NewServer(cfg ServerConfig, application *app.App) *Server {
	return &Server{
		config: cfg,
		app:    application,
		logger: slog.Default().With(slog.String("component", "web")),
	}
}

// setupRoutes configures all HTTP routes
func (s *Server) setupRoutes() http.Handler {
	r := chi.NewRouter()

	// Middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	// Use slog for request logging
	r.Use(slogchi.New(s.logger))

	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS middleware for API access
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	// Initialize handlers
	apiHandler := handlers.NewAPIHandler(s.app, s.logger)
	swaggerHandler := handlers.NewSwaggerHandler(handlers.SwaggerConfig{
		JSONPath: "/swagger/doc.json",
		DocPath:  "./pkg/server/swagger/swagger.json",
	})

	// Swagger documentation
	r.Mount("/swagger", swaggerHandler.UI())

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(middleware.SetHeader("Content-Type", "application/json"))

		// Swagger endpoint
		r.Get("/swagger/doc.json", swaggerHandler.Doc)

		// Network endpoints
		r.Get("/networks", apiHandler.ListNetworks)
		r.Get("/networks/{network}", apiHandler.GetNetwork)
		r.Get("/networks/{network}/sequencers", apiHandler.GetSequencers)

		// Sequencer actions
		r.Route("/sequencers/{id}", func(r chi.Router) {
			r.Post("/pause", apiHandler.PauseSequencer)
			r.Post("/resume", apiHandler.ResumeSequencer)
			r.Post("/transfer-leader", apiHandler.TransferLeader)
			r.Post("/resign-leader", apiHandler.ResignLeader)
			r.Post("/override-leader", apiHandler.OverrideLeader)
			r.Post("/halt", apiHandler.HaltSequencer)
			r.Post("/force-active", apiHandler.ForceActive)
			r.Delete("/membership", apiHandler.RemoveFromCluster)
			r.Put("/membership", apiHandler.UpdateMembership)
		})

		// WebSocket for real-time updates
		r.Get("/ws", apiHandler.WebSocket)
	})

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Serve React app for all non-API routes
	contentStatic, err := fs.Sub(content, "dist")
	if err != nil {
		s.logger.Error("Failed to create sub filesystem", slog.String("error", err.Error()))
		panic(err)
	}
	fileServer := http.FileServer(http.FS(contentStatic))

	// Serve React app for all other routes (SPA catch-all)
	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		// For SPA, always serve index.html for client-side routing
		if r.URL.Path != "/" {
			_, err := contentStatic.Open(r.URL.Path[1:])
			if err != nil {
				r.URL.Path = "/"
			}
		}
		fileServer.ServeHTTP(w, r)
	})

	return r
}

// Start begins serving HTTP requests
func (s *Server) Start(ctx context.Context) error {
	router := s.setupRoutes()

	s.httpServer = &http.Server{
		Addr:           fmt.Sprintf("%s:%d", s.config.Address, s.config.Port),
		Handler:        router,
		ReadTimeout:    s.config.ReadTimeout,
		WriteTimeout:   s.config.WriteTimeout,
		IdleTimeout:    s.config.IdleTimeout,
		MaxHeaderBytes: s.config.MaxHeaderBytes,
	}

	// Start server in a goroutine
	serverErr := make(chan error, 1)
	go func() {
		s.logger.Info("Starting web server", slog.String("address", s.config.Address), slog.Int("port", s.config.Port))
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("HTTP server error", slog.String("error", err.Error()))
			serverErr <- err
		}
	}()

	// Wait for context cancellation or server error
	select {
	case <-ctx.Done():
		s.logger.Info("Shutting down server...")

		// Graceful shutdown
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
			s.logger.Error("Graceful shutdown failed", slog.String("error", err.Error()))
			if closeErr := s.httpServer.Close(); closeErr != nil {
				s.logger.Error("Force close failed", slog.String("error", closeErr.Error()))
			}
			return fmt.Errorf("server shutdown error: %w", err)
		}

		s.logger.Info("Server shut down gracefully")
		return nil

	case err := <-serverErr:
		return fmt.Errorf("server error: %w", err)
	}
}
