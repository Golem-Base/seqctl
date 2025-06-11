package repository

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/golem-base/seqctl/pkg/network"
	"github.com/golem-base/seqctl/pkg/provider"
)

// NetworkRepository provides access to networks with caching capabilities
type NetworkRepository interface {
	// GetNetwork returns a network by name with updated status
	GetNetwork(ctx context.Context, name string) (*network.Network, error)

	// ListNetworks returns all available networks
	ListNetworks(ctx context.Context) (map[string]*network.Network, error)

	// RefreshCache forces a cache refresh from the provider
	RefreshCache(ctx context.Context) error

	// InvalidateNetwork removes a specific network from cache
	InvalidateNetwork(name string)

	// InvalidateAll clears the entire cache
	InvalidateAll()
}

// CachedNetworkRepository implements NetworkRepository with caching
type CachedNetworkRepository struct {
	provider provider.Provider

	// Cache state
	networks      map[string]*network.Network
	lastDiscovery time.Time

	// Cache configuration
	discoveryTTL time.Duration // How long to cache network discovery
	statusTTL    time.Duration // How long before updating network status

	// Thread safety
	mu sync.RWMutex
}

// NewCachedNetworkRepository creates a new repository with caching
func NewCachedNetworkRepository(provider provider.Provider, discoveryTTL, statusTTL time.Duration) *CachedNetworkRepository {
	if discoveryTTL == 0 {
		discoveryTTL = 5 * time.Minute
	}
	if statusTTL == 0 {
		statusTTL = 10 * time.Second
	}

	return &CachedNetworkRepository{
		provider:     provider,
		networks:     make(map[string]*network.Network),
		discoveryTTL: discoveryTTL,
		statusTTL:    statusTTL,
	}
}

// GetNetwork returns a network by name with updated status
func (r *CachedNetworkRepository) GetNetwork(ctx context.Context, name string) (*network.Network, error) {
	// Check if we need to refresh discovery
	if r.shouldRefreshDiscovery() {
		if err := r.RefreshCache(ctx); err != nil {
			// Log error but continue with stale data if available
		}
	}

	r.mu.RLock()
	net, exists := r.networks[name]
	r.mu.RUnlock()

	if !exists {
		// Try one more refresh before giving up
		if err := r.RefreshCache(ctx); err != nil {
			return nil, fmt.Errorf("failed to discover networks: %w", err)
		}

		r.mu.RLock()
		net, exists = r.networks[name]
		r.mu.RUnlock()

		if !exists {
			return nil, fmt.Errorf("network %s not found", name)
		}
	}

	// Update network status if needed
	if r.shouldUpdateStatus(net) {
		if err := r.updateNetworkStatus(ctx, net); err != nil {
			return net, fmt.Errorf("failed to update network %s status: %w", name, err)
		}
	}

	return net, nil
}

// ListNetworks returns all available networks
func (r *CachedNetworkRepository) ListNetworks(ctx context.Context) (map[string]*network.Network, error) {
	if r.shouldRefreshDiscovery() {
		if err := r.RefreshCache(ctx); err != nil {
			r.mu.RLock()
			defer r.mu.RUnlock()

			if len(r.networks) == 0 {
				return nil, fmt.Errorf("failed to discover networks and cache is empty: %w", err)
			}
		}
	}

	r.mu.RLock()
	networks := make([]*network.Network, 0, len(r.networks))
	for _, net := range r.networks {
		networks = append(networks, net)
	}
	r.mu.RUnlock()

	// Update status for networks that need it
	for _, net := range networks {
		if r.shouldUpdateStatus(net) {
			// Update but don't fail if status update fails
			_ = r.updateNetworkStatus(ctx, net)
		}
	}

	// Return a copy to avoid race conditions
	result := make(map[string]*network.Network, len(networks))
	for _, net := range networks {
		result[net.Name()] = net
	}

	return result, nil
}

// RefreshCache forces a cache refresh from the provider
func (r *CachedNetworkRepository) RefreshCache(ctx context.Context) error {
	networks, err := r.provider.DiscoverNetworks(ctx)
	if err != nil {
		return fmt.Errorf("failed to discover networks using %s provider: %w", r.provider.Name(), err)
	}

	// Update status for all networks to populate timestamps
	for _, net := range networks {
		// Update in a separate goroutine with timeout to avoid blocking
		if err := r.updateNetworkStatus(ctx, net); err != nil {
			// Log error but continue with other networks
			// Networks will still be cached even if status update fails
		}
	}

	r.mu.Lock()
	r.networks = networks
	r.lastDiscovery = time.Now()
	r.mu.Unlock()

	return nil
}

// InvalidateNetwork removes a specific network from cache
func (r *CachedNetworkRepository) InvalidateNetwork(name string) {
	r.mu.Lock()
	delete(r.networks, name)
	r.mu.Unlock()
}

// InvalidateAll clears the entire cache
func (r *CachedNetworkRepository) InvalidateAll() {
	r.mu.Lock()
	r.networks = make(map[string]*network.Network)
	r.lastDiscovery = time.Time{}
	r.mu.Unlock()
}

// shouldRefreshDiscovery checks if discovery cache is stale
func (r *CachedNetworkRepository) shouldRefreshDiscovery() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.lastDiscovery.IsZero() || time.Since(r.lastDiscovery) > r.discoveryTTL
}

// shouldUpdateStatus checks if network status needs updating
func (r *CachedNetworkRepository) shouldUpdateStatus(net *network.Network) bool {
	return time.Since(net.LastUpdateTime()) > r.statusTTL
}

// updateNetworkStatus updates a single network's status
func (r *CachedNetworkRepository) updateNetworkStatus(ctx context.Context, net *network.Network) error {
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	return net.Update(ctx)
}
