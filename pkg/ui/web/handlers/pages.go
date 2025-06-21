package handlers

import (
	"log/slog"
	"net/http"

	chi "github.com/go-chi/chi/v5"
	"github.com/golem-base/seqctl/pkg/app"
	"github.com/golem-base/seqctl/pkg/ui/web/templates"
)

// PageHandler handles page requests
type PageHandler struct {
	app             *app.App
	logger          *slog.Logger
	refreshInterval int
}

// NewPageHandler creates a new page handler
func NewPageHandler(application *app.App, logger *slog.Logger, refreshInterval int) *PageHandler {
	return &PageHandler{
		app:             application,
		logger:          logger.With(slog.String("component", "pages")),
		refreshInterval: refreshInterval,
	}
}

// Index serves the main page listing all networks
func (h *PageHandler) Index(w http.ResponseWriter, r *http.Request) {
	networks, err := h.app.ListNetworks(r.Context())
	if err != nil {
		h.logger.Error("Failed to list networks", "error", err)
		http.Error(w, "Failed to list networks", http.StatusInternalServerError)
		return
	}

	h.logger.Debug("Rendering index page", slog.Int("networks", len(networks)))

	// Render the index template
	if err := templates.Index(networks, h.refreshInterval).Render(r.Context(), w); err != nil {
		h.logger.Error("Failed to render index page", slog.String("error", err.Error()))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// NetworkDetail serves the network detail page
func (h *PageHandler) NetworkDetail(w http.ResponseWriter, r *http.Request) {
	networkName := chi.URLParam(r, "network")

	network, err := h.app.GetNetwork(r.Context(), networkName)
	if err != nil {
		h.logger.Error("Network not found",
			slog.String("network", networkName),
			slog.String("error", err.Error()),
		)
		http.Error(w, "Network not found", http.StatusNotFound)
		return
	}

	h.logger.Debug("Rendering network detail page",
		slog.String("network", networkName),
		slog.Int("sequencers", len(network.Sequencers())),
	)

	// Render the network detail template
	if err := templates.NetworkDetail(network, h.refreshInterval).Render(r.Context(), w); err != nil {
		h.logger.Error("Failed to render network detail page", slog.String("error", err.Error()))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
