package sequencer

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/sync/errgroup"

	"github.com/ethereum-optimism/optimism/op-conductor/consensus"
	"github.com/ethereum-optimism/optimism/op-service/eth"

	"github.com/golem-base/seqctl/pkg/rpc"
)

// Status represents the current status of a sequencer
type Status struct {
	ConductorActive  bool
	ConductorLeader  bool
	ConductorPaused  bool
	ConductorStopped bool
	SequencerHealthy bool
	SequencerActive  bool
	UnsafeL2         *eth.L2BlockRef
	LastUpdateTime   time.Time
}

// Config holds the configuration for a sequencer
type Config struct {
	ID           string
	RaftAddr     string
	ConductorURL string
	NodeURL      string
	Voting       bool
	Network      string
}

// Sequencer represents a sequencer in a network
type Sequencer struct {
	// Immutable configuration
	config Config

	// Mutable state - atomic for lock-free reads
	status atomic.Pointer[Status]

	// RPC client
	client *rpc.Client

	// Error tracking (still needs mutex)
	mu            sync.Mutex
	lastError     error
	lastErrorTime time.Time
}

// New creates a new initialized sequencer instance
func New(ctx context.Context, cfg Config, rpcOpts ...rpc.ClientOption) (*Sequencer, error) {
	// Create RPC client immediately
	client, err := rpc.NewClientWithContext(ctx, cfg.ConductorURL, cfg.NodeURL, rpcOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create RPC client for sequencer %s: %w", cfg.ID, err)
	}

	s := &Sequencer{
		config: cfg,
		client: client,
	}

	// Initialize with empty status
	s.status.Store(&Status{})

	slog.Debug("Sequencer created and initialized",
		"sequencer", cfg.ID,
		"conductorURL", cfg.ConductorURL,
		"nodeURL", cfg.NodeURL)

	return s, nil
}

// Update fetches the current status of the sequencer
func (s *Sequencer) Update(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	slog.Debug("Updating sequencer status", "sequencer", s.config.ID)

	// Fetch status concurrently
	g, ctx := errgroup.WithContext(ctx)

	var status Status
	g.Go(func() error {
		active, err := s.client.Active(ctx)
		if err != nil {
			slog.Debug("Conductor active check failed",
				"sequencer", s.config.ID,
				"error", err)
			return fmt.Errorf("conductor active check failed for sequencer %s: %w", s.config.ID, err)
		}
		status.ConductorActive = active
		return nil
	})

	g.Go(func() error {
		leader, err := s.client.Leader(ctx)
		if err != nil {
			slog.Debug("Conductor leader check failed",
				"sequencer", s.config.ID,
				"error", err)
			return fmt.Errorf("conductor leader check failed for sequencer %s: %w", s.config.ID, err)
		}
		status.ConductorLeader = leader
		return nil
	})

	g.Go(func() error {
		paused, err := s.client.Paused(ctx)
		if err != nil {
			slog.Debug("Conductor paused check failed",
				"sequencer", s.config.ID,
				"error", err)
			return fmt.Errorf("conductor paused check failed for sequencer %s: %w", s.config.ID, err)
		}
		status.ConductorPaused = paused
		return nil
	})

	g.Go(func() error {
		stopped, err := s.client.Stopped(ctx)
		if err != nil {
			slog.Debug("Conductor stopped check failed",
				"sequencer", s.config.ID,
				"error", err)
			return fmt.Errorf("conductor stopped check failed for sequencer %s: %w", s.config.ID, err)
		}
		status.ConductorStopped = stopped
		return nil
	})

	g.Go(func() error {
		healthy, err := s.client.SequencerHealthy(ctx)
		if err != nil {
			slog.Debug("Sequencer healthy check failed",
				"sequencer", s.config.ID,
				"error", err)
			return fmt.Errorf("sequencer healthy check failed for sequencer %s: %w", s.config.ID, err)
		}
		status.SequencerHealthy = healthy
		return nil
	})

	g.Go(func() error {
		active, err := s.client.SequencerActive(ctx)
		if err != nil {
			slog.Debug("Sequencer active check failed",
				"sequencer", s.config.ID,
				"error", err)
			return fmt.Errorf("sequencer active check failed for sequencer %s: %w", s.config.ID, err)
		}
		status.SequencerActive = active
		return nil
	})

	g.Go(func() error {
		syncStatus, err := s.client.SyncStatus(ctx)
		if err != nil {
			slog.Debug("Sync status check failed",
				"sequencer", s.config.ID,
				"error", err)
			return fmt.Errorf("sync status check failed for sequencer %s: %w", s.config.ID, err)
		}

		if syncStatus != nil {
			status.UnsafeL2 = &syncStatus.UnsafeL2
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		s.lastError = err
		s.lastErrorTime = time.Now()
		slog.Error("Failed to update sequencer status",
			"sequencer", s.config.ID,
			"error", err)
		return err
	}

	// Update status and track update time
	status.LastUpdateTime = time.Now()
	s.status.Store(&status)
	s.lastError = nil
	s.lastErrorTime = time.Time{}

	slog.Debug("Sequencer status updated successfully",
		"sequencer", s.config.ID,
		"active", status.ConductorActive,
		"leader", status.ConductorLeader,
		"healthy", status.SequencerHealthy,
		"sequencing", status.SequencerActive)

	return nil
}

// LastError returns the last error encountered during update
func (s *Sequencer) LastError() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastError
}

// ClearError explicitly clears the last error
func (s *Sequencer) ClearError() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastError = nil
	s.lastErrorTime = time.Time{}
}

// ResetClients forces the clients to be reinitialized on the next operation
func (s *Sequencer) ResetClients() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client != nil {
		s.client.Close()
		s.client = nil
	}
}

// GetClusterMembership returns the cluster membership
func (s *Sequencer) GetClusterMembership(ctx context.Context) (*consensus.ClusterMembership, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	membership, err := s.client.ClusterMembership(ctx)
	if err != nil {
		slog.Error("Failed to get cluster membership",
			"sequencer", s.config.ID,
			"error", err)
		return nil, fmt.Errorf("failed to get cluster membership for sequencer %s: %w", s.config.ID, err)
	}
	return membership, nil
}

// Pause pauses the conductor
func (s *Sequencer) Pause(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.client.Pause(ctx); err != nil {
		slog.Error("Failed to pause conductor",
			"sequencer", s.config.ID,
			"error", err)
		return fmt.Errorf("failed to pause conductor for sequencer %s: %w", s.config.ID, err)
	}
	return nil
}

// Resume resumes the conductor
func (s *Sequencer) Resume(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.client.Resume(ctx); err != nil {
		slog.Error("Failed to resume conductor",
			"sequencer", s.config.ID,
			"error", err)
		return fmt.Errorf("failed to resume conductor for sequencer %s: %w", s.config.ID, err)
	}
	return nil
}

// TransferLeaderToServer transfers leadership to another server
func (s *Sequencer) TransferLeaderToServer(ctx context.Context, id, addr string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.client.TransferLeaderToServer(ctx, id, addr); err != nil {
		slog.Error("Failed to transfer leadership",
			"from", s.config.ID,
			"to", id,
			"addr", addr,
			"error", err)
		return fmt.Errorf("failed to transfer leadership from %s to %s: %w", s.config.ID, id, err)
	}
	return nil
}

// OverrideLeader overrides the conductor leader status
func (s *Sequencer) OverrideLeader(ctx context.Context, override bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.client.OverrideLeader(ctx, override); err != nil {
		slog.Error("Failed to override leader status",
			"sequencer", s.config.ID,
			"override", override,
			"error", err)
		return fmt.Errorf("failed to override leader status for sequencer %s (override=%t): %w", s.config.ID, override, err)
	}
	return nil
}

// AddServerAsVoter adds a server as a voter
func (s *Sequencer) AddServerAsVoter(ctx context.Context, id, addr string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.client.AddServerAsVoter(ctx, id, addr, 0); err != nil {
		slog.Error("Failed to add server as voter",
			"sequencer", s.config.ID,
			"server", id,
			"addr", addr,
			"error", err)
		return fmt.Errorf("failed to add server %s as voter to sequencer %s: %w", id, s.config.ID, err)
	}
	return nil
}

// AddServerAsNonvoter adds a server as a non-voter
func (s *Sequencer) AddServerAsNonvoter(ctx context.Context, id, addr string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.client.AddServerAsNonvoter(ctx, id, addr, 0); err != nil {
		slog.Error("Failed to add server as non-voter",
			"sequencer", s.config.ID,
			"server", id,
			"addr", addr,
			"error", err)
		return fmt.Errorf("failed to add server %s as non-voter to sequencer %s: %w", id, s.config.ID, err)
	}
	return nil
}

// RemoveServer removes a server from the cluster
func (s *Sequencer) RemoveServer(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.client.RemoveServer(ctx, id, 0); err != nil {
		slog.Error("Failed to remove server",
			"sequencer", s.config.ID,
			"server", id,
			"error", err)
		return fmt.Errorf("failed to remove server %s from sequencer %s: %w", id, s.config.ID, err)
	}
	return nil
}

// StopSequencer stops the sequencer
func (s *Sequencer) StopSequencer(ctx context.Context) (common.Hash, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	hash, err := s.client.StopSequencer(ctx)
	if err != nil {
		slog.Error("Failed to stop sequencer",
			"sequencer", s.config.ID,
			"error", err)
		return common.Hash{}, fmt.Errorf("failed to stop sequencer %s: %w", s.config.ID, err)
	}
	return hash, nil
}

// StartSequencer starts the sequencer with the given hash
func (s *Sequencer) StartSequencer(ctx context.Context, hash common.Hash) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.client.StartSequencer(ctx, hash); err != nil {
		slog.Error("Failed to start sequencer",
			"sequencer", s.config.ID,
			"hash", hash.String(),
			"error", err)
		return fmt.Errorf("failed to start sequencer %s with hash %s: %w", s.config.ID, hash.String(), err)
	}
	return nil
}

// OverrideNodeLeader overrides the node leader status
func (s *Sequencer) OverrideNodeLeader(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.client.OverrideNodeLeader(ctx); err != nil {
		slog.Error("Failed to override node leader",
			"sequencer", s.config.ID,
			"error", err)
		return fmt.Errorf("failed to override node leader for sequencer %s: %w", s.config.ID, err)
	}
	return nil
}

// IsPaused checks if the conductor is paused
func (s *Sequencer) IsPaused(ctx context.Context) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	paused, err := s.client.Paused(ctx)
	if err != nil {
		slog.Error("Failed to check paused status",
			"sequencer", s.config.ID,
			"error", err)
		return false, fmt.Errorf("failed to check paused status for sequencer %s: %w", s.config.ID, err)
	}
	return paused, nil
}

// IsStopped checks if the conductor is stopped
func (s *Sequencer) IsStopped(ctx context.Context) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	stopped, err := s.client.Stopped(ctx)
	if err != nil {
		slog.Error("Failed to check stopped status",
			"sequencer", s.config.ID,
			"error", err)
		return false, fmt.Errorf("failed to check stopped status for sequencer %s: %w", s.config.ID, err)
	}
	return stopped, nil
}

// GetLeaderWithID returns the current leader's server info
func (s *Sequencer) GetLeaderWithID(ctx context.Context) (*consensus.ServerInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	info, err := s.client.LeaderWithID(ctx)
	if err != nil {
		slog.Error("Failed to get leader with ID",
			"sequencer", s.config.ID,
			"error", err)
		return nil, fmt.Errorf("failed to get leader with ID for sequencer %s: %w", s.config.ID, err)
	}
	return info, nil
}

// TransferLeader transfers leadership (resigns from leadership)
func (s *Sequencer) TransferLeader(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.client.TransferLeader(ctx); err != nil {
		slog.Error("Failed to transfer leadership",
			"sequencer", s.config.ID,
			"error", err)
		return fmt.Errorf("failed to transfer leadership for sequencer %s: %w", s.config.ID, err)
	}
	return nil
}

// ID returns the sequencer ID
func (s *Sequencer) ID() string {
	return s.config.ID
}

// RaftAddr returns the raft address
func (s *Sequencer) RaftAddr() string {
	return s.config.RaftAddr
}

// Voting returns the voting status
func (s *Sequencer) Voting() bool {
	return s.config.Voting
}

// Network returns the network name
func (s *Sequencer) Network() string {
	return s.config.Network
}

// Status returns a copy of the current status for safe concurrent access
func (s *Sequencer) Status() Status {
	if status := s.status.Load(); status != nil {
		return *status
	}
	return Status{}
}

// ConductorActive returns the conductor active status
func (s *Sequencer) ConductorActive() bool {
	return s.Status().ConductorActive
}

// ConductorLeader returns the conductor leader status
func (s *Sequencer) ConductorLeader() bool {
	return s.Status().ConductorLeader
}

// SequencerHealthy returns the sequencer healthy status
func (s *Sequencer) SequencerHealthy() bool {
	return s.Status().SequencerHealthy
}

// SequencerActive returns the sequencer active status
func (s *Sequencer) SequencerActive() bool {
	return s.Status().SequencerActive
}

// UnsafeL2 returns the unsafe L2 block number
func (s *Sequencer) UnsafeL2() uint64 {
	status := s.Status()
	if status.UnsafeL2 != nil {
		return status.UnsafeL2.Number
	}
	return 0
}

// ConductorPaused returns the conductor paused status
func (s *Sequencer) ConductorPaused() bool {
	return s.Status().ConductorPaused
}

// ConductorStopped returns the conductor stopped status
func (s *Sequencer) ConductorStopped() bool {
	return s.Status().ConductorStopped
}
