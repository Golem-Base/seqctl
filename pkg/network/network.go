package network

import (
	"context"
	"sync"
	"time"

	"github.com/golem-base/seqctl/pkg/sequencer"
	"golang.org/x/sync/errgroup"
)

// Network represents a network of sequencers
type Network struct {
	name       string
	sequencers []*sequencer.Sequencer

	mu             sync.Mutex
	lastUpdateTime time.Time
	updateError    error
}

// NewNetwork creates a new network
func NewNetwork(name string, sequencers []*sequencer.Sequencer) *Network {
	return &Network{
		name:       name,
		sequencers: sequencers,
	}
}

// Name returns the network name
func (n *Network) Name() string {
	return n.name
}

// Sequencers returns the network's sequencers
func (n *Network) Sequencers() []*sequencer.Sequencer {
	return n.sequencers
}

// Update updates all sequencers in the network concurrently
func (n *Network) Update(ctx context.Context) error {
	errg, ctx := errgroup.WithContext(ctx)

	for _, seq := range n.sequencers {
		seq := seq
		errg.Go(func() error {
			// This performs the concurrent updates without holding the network-level lock.
			return seq.Update(ctx)
		})
	}

	// Wait for all updates to complete.
	err := errg.Wait()

	// Now, acquire the lock only to update the shared fields.
	n.mu.Lock()
	defer n.mu.Unlock()
	n.lastUpdateTime = time.Now()
	n.updateError = err

	return err
}

// SequencerByID returns a sequencer by its ID or nil if not found
func (n *Network) SequencerByID(id string) *sequencer.Sequencer {
	for _, seq := range n.sequencers {
		if seq.ID() == id {
			return seq
		}
	}
	return nil
}

// ConductorLeader returns the sequencer that is the conductor leader or nil if none
func (n *Network) ConductorLeader() *sequencer.Sequencer {
	for _, seq := range n.sequencers {
		if seq.ConductorLeader() {
			return seq
		}
	}
	return nil
}

// ActiveSequencer returns the sequencer that is active or nil if none
func (n *Network) ActiveSequencer() *sequencer.Sequencer {
	for _, seq := range n.sequencers {
		if seq.SequencerActive() {
			return seq
		}
	}
	return nil
}

// IsHealthy returns true if all sequencers are healthy
func (n *Network) IsHealthy() bool {
	for _, seq := range n.sequencers {
		if !seq.SequencerHealthy() {
			return false
		}
	}
	return true
}

// LastUpdateTime returns the time of the last update
func (n *Network) LastUpdateTime() time.Time {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.lastUpdateTime
}

// LastError returns the last error encountered during update
func (n *Network) LastError() error {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.updateError
}

// UpdateSuccessful returns true if the last update was successful
func (n *Network) UpdateSuccessful() bool {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.updateError == nil && !n.lastUpdateTime.IsZero()
}

// UpdatedAt returns the time of the last update (alias for LastUpdateTime)
func (n *Network) UpdatedAt() time.Time {
	return n.LastUpdateTime()
}
