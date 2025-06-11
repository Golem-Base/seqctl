package provider

import (
	"context"

	"github.com/golem-base/seqctl/pkg/network"
)

// Provider defines the interface for discovering sequencer infrastructure
type Provider interface {
	// DiscoverNetworks returns all available networks with their sequencers
	DiscoverNetworks(ctx context.Context) (map[string]*network.Network, error)

	// Name returns the provider type
	Name() string
}
