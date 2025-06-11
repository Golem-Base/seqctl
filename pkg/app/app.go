package app

import (
	"context"

	"github.com/golem-base/seqctl/pkg/config"
	"github.com/golem-base/seqctl/pkg/network"
	"github.com/golem-base/seqctl/pkg/repository"
)

// App is the main application container that holds all services and configuration
type App struct {
	Config     *config.Config
	repository repository.NetworkRepository
}

// New creates a new application container with the given configuration and repository
func New(cfg *config.Config, repo repository.NetworkRepository) *App {
	return &App{
		Config:     cfg,
		repository: repo,
	}
}

// GetNetwork returns a network by name with updated status
func (a *App) GetNetwork(ctx context.Context, networkName string) (*network.Network, error) {
	return a.repository.GetNetwork(ctx, networkName)
}

// RefreshNetworks clears the cache and re-discovers all networks
func (a *App) RefreshNetworks(ctx context.Context) error {
	return a.repository.RefreshCache(ctx)
}

// ListNetworks returns all cached networks
func (a *App) ListNetworks(ctx context.Context) (map[string]*network.Network, error) {
	return a.repository.ListNetworks(ctx)
}
