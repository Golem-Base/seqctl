package app

import (
	"context"
	"fmt"
	"maps"
	"sync"
	"time"

	"github.com/golem-base/seqctl/pkg/config"
	"github.com/golem-base/seqctl/pkg/network"
	"github.com/golem-base/seqctl/pkg/provider"
)

// App is the main application container that holds all services and configuration
type App struct {
	Config   *config.Config
	provider provider.Provider
	networks map[string]*network.Network
	mu       sync.RWMutex
}

// New creates a new application container with the given configuration and provider
func New(cfg *config.Config, prov provider.Provider) *App {
	return &App{
		Config:   cfg,
		provider: prov,
		networks: make(map[string]*network.Network),
	}
}

// GetNetwork returns a network by name with updated status
func (a *App) GetNetwork(ctx context.Context, networkName string) (*network.Network, error) {
	// Check cache first
	a.mu.RLock()
	if network, exists := a.networks[networkName]; exists {
		a.mu.RUnlock()
		if err := a.updateNetwork(ctx, network); err != nil {
			return network, fmt.Errorf("failed to update network %s status: %w", networkName, err)
		}
		return network, nil
	}
	a.mu.RUnlock()

	// Discover and cache networks
	if err := a.RefreshNetworks(ctx); err != nil {
		return nil, fmt.Errorf("failed to discover networks: %w", err)
	}

	// Check cache again after refresh
	a.mu.RLock()
	defer a.mu.RUnlock()
	if network, exists := a.networks[networkName]; exists {
		return network, nil
	}

	return nil, fmt.Errorf("network %s not found", networkName)
}

// RefreshNetworks clears the cache and re-discovers all networks
func (a *App) RefreshNetworks(ctx context.Context) error {
	networks, err := a.provider.DiscoverNetworks(ctx)
	if err != nil {
		return fmt.Errorf("failed to discover networks using %s provider: %w", a.provider.Name(), err)
	}

	a.mu.Lock()
	a.networks = networks
	a.mu.Unlock()

	return nil
}

// ListNetworks returns all cached networks
func (a *App) ListNetworks() map[string]*network.Network {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// Return a copy to avoid race conditions
	copy := make(map[string]*network.Network, len(a.networks))
	maps.Copy(copy, a.networks)
	return copy
}

// updateNetwork updates a single network's status with timeout
func (a *App) updateNetwork(ctx context.Context, net *network.Network) error {
	// Use a timeout for the update if not already set in context
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	return net.Update(ctx)
}
