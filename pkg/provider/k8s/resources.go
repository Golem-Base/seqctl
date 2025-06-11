package k8s

import (
	"fmt"
	"net/http"
	"time"

	"github.com/golem-base/seqctl/pkg/sequencer"
)

const (
	// Default timeout for sequencer operations
	DefaultSequencerTimeout = 10 * time.Second
)

// SequencerResource represents a discovered Kubernetes sequencer
type SequencerResource struct {
	Name        string
	Namespace   string
	Network     string
	Role        string
	RaftAddr    string
	RPCURLs     map[string]string
	IsBootstrap bool
	IsVoting    bool
	HTTPClient  *http.Client
}

// ToSequencerConfig converts a Kubernetes resource to a sequencer config
func (r *SequencerResource) ToSequencerConfig() sequencer.Config {
	return sequencer.Config{
		ID:              r.Name,
		RaftAddr:        r.RaftAddr,
		ConductorRPCURL: r.RPCURLs[PortConductorRPC],
		NodeRPCURL:      r.RPCURLs[PortNodeRPC],
		Voting:          r.IsVoting,
		Timeout:         DefaultSequencerTimeout,
		HTTPClient:      r.HTTPClient,
	}
}

// String returns a string representation of the sequencer resource
func (r *SequencerResource) String() string {
	roleTag := ""
	if r.IsBootstrap {
		roleTag = " (bootstrap)"
	}
	votingTag := ""
	if r.IsVoting {
		votingTag = " (voting)"
	}

	return fmt.Sprintf("%s/%s [%s]%s%s", r.Namespace, r.Name, r.Network, roleTag, votingTag)
}
