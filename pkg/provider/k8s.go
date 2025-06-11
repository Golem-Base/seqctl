package provider

import (
	"context"
	"fmt"

	"github.com/golem-base/seqctl/pkg/network"
	"github.com/golem-base/seqctl/pkg/provider/k8s"
	"github.com/golem-base/seqctl/pkg/sequencer"
)

// K8sProvider discovers networks from Kubernetes
type K8sProvider struct {
	Client    *k8s.Client
	Namespace string
	Selector  string
}

// NewK8sProvider creates a new Kubernetes network provider
func NewK8sProvider(client *k8s.Client, namespace, selector string) *K8sProvider {
	return &K8sProvider{
		Client:    client,
		Namespace: namespace,
		Selector:  selector,
	}
}

// DiscoverNetworks finds all networks defined in Kubernetes
func (p *K8sProvider) DiscoverNetworks(ctx context.Context) (map[string]*network.Network, error) {
	resources, err := p.Client.DiscoverSequencers(ctx, p.Namespace, p.Selector)
	if err != nil {
		return nil, fmt.Errorf("failed to discover sequencers in namespace %s with selector %s: %w", p.Namespace, p.Selector, err)
	}

	// Group sequencers by network
	networkMap := make(map[string][]*sequencer.Sequencer)

	for _, resource := range resources {
		seq := sequencer.NewSequencer(resource.ToSequencerConfig())
		networkMap[resource.Network] = append(networkMap[resource.Network], seq)
	}

	// Create network objects
	networks := make(map[string]*network.Network)

	for networkName, sequencers := range networkMap {
		networks[networkName] = network.NewNetwork(networkName, sequencers)
	}

	return networks, nil
}

// Name returns the provider type
func (p *K8sProvider) Name() string {
	return "kubernetes"
}
