package provider

import (
	"fmt"

	"github.com/golem-base/seqctl/pkg/config"
)

// NewProvider creates a provider based on the configuration
func NewProvider(cfg *config.Config) (Provider, error) {
	provider, err := NewK8sProvider(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes provider: %w", err)
	}

	return provider, nil
}
