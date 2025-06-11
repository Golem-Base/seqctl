package sequencer

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	ethrpc "github.com/ethereum/go-ethereum/rpc"
	"golang.org/x/sync/errgroup"

	"github.com/ethereum-optimism/optimism/op-conductor/consensus"
	"github.com/ethereum-optimism/optimism/op-conductor/rpc"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/sources"
)

// Config holds the configuration for a sequencer
type Config struct {
	ID              string `koanf:"sequencer_id"`
	RaftAddr        string `koanf:"raft_addr"`
	ConductorRPCURL string `koanf:"conductor_rpc_url"`
	NodeRPCURL      string `koanf:"node_rpc_url"`
	Voting          bool   `koanf:"voting"`
	Timeout         time.Duration
	HTTPClient      *http.Client
}

// Status represents the current status of a sequencer
type Status struct {
	ConductorActive  bool
	ConductorLeader  bool
	SequencerHealthy bool
	SequencerActive  bool
	UnsafeL2         *eth.L2BlockRef
	LastUpdateTime   time.Time
}

// Sequencer represents a sequencer in a network
type Sequencer struct {
	Config Config
	Status Status

	conductorClient *rpc.APIClient
	nodeClient      *sources.RollupClient

	mu            sync.Mutex
	lastError     error
	lastErrorTime time.Time
}

// NewSequencer creates a new sequencer instance
func NewSequencer(config Config) *Sequencer {
	// Set default timeout if not specified
	if config.Timeout == 0 {
		config.Timeout = 5 * time.Second
	}

	return &Sequencer{
		Config: config,
	}
}

// Initialize establishes connections to the conductor and node RPC endpoints
func (s *Sequencer) Initialize(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Initialize conductor client
	if s.conductorClient == nil {
		var rpcClient *ethrpc.Client
		var err error

		if s.Config.HTTPClient != nil {
			slog.Debug("Using authenticated HTTP client",
				"sequencer", s.Config.ID,
				"url", s.Config.ConductorRPCURL)

			rpcClient, err = ethrpc.DialHTTPWithClient(s.Config.ConductorRPCURL, s.Config.HTTPClient)
		} else {
			slog.Debug("Using default HTTP client",
				"sequencer", s.Config.ID,
				"url", s.Config.ConductorRPCURL)

			rpcClient, err = ethrpc.DialContext(ctx, s.Config.ConductorRPCURL)
		}

		if err != nil {
			slog.Error("Failed to connect to conductor RPC",
				"sequencer", s.Config.ID,
				"url", s.Config.ConductorRPCURL,
				"error", err)
			return fmt.Errorf("failed to connect to conductor RPC for sequencer %s at %s: %w", s.Config.ID, s.Config.ConductorRPCURL, err)
		}

		slog.Debug("Connected to conductor RPC", "sequencer", s.Config.ID)
		s.conductorClient = rpc.NewAPIClient(rpcClient)
	}

	// Initialize node client
	if s.nodeClient == nil {
		var rpcClient *ethrpc.Client
		var err error

		if s.Config.HTTPClient != nil {
			slog.Debug("Using authenticated HTTP client for node RPC",
				"sequencer", s.Config.ID,
				"url", s.Config.NodeRPCURL)

			rpcClient, err = ethrpc.DialHTTPWithClient(s.Config.NodeRPCURL, s.Config.HTTPClient)
		} else {
			slog.Debug("Using default HTTP client for node RPC",
				"sequencer", s.Config.ID,
				"url", s.Config.NodeRPCURL)

			rpcClient, err = ethrpc.DialContext(ctx, s.Config.NodeRPCURL)
		}

		if err != nil {
			slog.Error("Failed to connect to node RPC",
				"sequencer", s.Config.ID,
				"url", s.Config.NodeRPCURL,
				"error", err)
			return fmt.Errorf("failed to connect to node RPC for sequencer %s at %s: %w", s.Config.ID, s.Config.NodeRPCURL, err)
		}

		// Use our adapter to make the rpcClient compatible with the sources.RollupClient
		s.nodeClient = sources.NewRollupClient(NewRPCAdapter(rpcClient))
	}

	return nil
}

// ensureInitialized checks if clients are initialized and returns an error if not
func (s *Sequencer) ensureInitialized() error {
	if s.conductorClient == nil || s.nodeClient == nil {
		return fmt.Errorf("sequencer %s not initialized, call Initialize first", s.Config.ID)
	}
	return nil
}

// Update fetches the current status of the sequencer
func (s *Sequencer) Update(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.ensureInitialized(); err != nil {
		return err
	}

	slog.Debug("Updating sequencer status", "sequencer", s.Config.ID)

	// Fetch status concurrently
	g, ctx := errgroup.WithContext(ctx)

	var status Status
	g.Go(func() error {
		active, err := s.conductorClient.Active(ctx)
		if err != nil {
			slog.Debug("Conductor active check failed",
				"sequencer", s.Config.ID,
				"error", err)
			return fmt.Errorf("conductor active check failed for sequencer %s: %w", s.Config.ID, err)
		}
		status.ConductorActive = active
		return nil
	})

	g.Go(func() error {
		leader, err := s.conductorClient.Leader(ctx)
		if err != nil {
			slog.Debug("Conductor leader check failed",
				"sequencer", s.Config.ID,
				"error", err)
			return fmt.Errorf("conductor leader check failed for sequencer %s: %w", s.Config.ID, err)
		}
		status.ConductorLeader = leader
		return nil
	})

	g.Go(func() error {
		healthy, err := s.conductorClient.SequencerHealthy(ctx)
		if err != nil {
			slog.Debug("Sequencer healthy check failed",
				"sequencer", s.Config.ID,
				"error", err)
			return fmt.Errorf("sequencer healthy check failed for sequencer %s: %w", s.Config.ID, err)
		}
		status.SequencerHealthy = healthy
		return nil
	})

	g.Go(func() error {
		active, err := s.nodeClient.SequencerActive(ctx)
		if err != nil {
			slog.Debug("Sequencer active check failed",
				"sequencer", s.Config.ID,
				"error", err)
			return fmt.Errorf("sequencer active check failed for sequencer %s: %w", s.Config.ID, err)
		}
		status.SequencerActive = active
		return nil
	})

	g.Go(func() error {
		syncStatus, err := s.nodeClient.SyncStatus(ctx)
		if err != nil {
			slog.Debug("Sync status check failed",
				"sequencer", s.Config.ID,
				"error", err)
			return fmt.Errorf("sync status check failed for sequencer %s: %w", s.Config.ID, err)
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
			"sequencer", s.Config.ID,
			"error", err)
		return err
	}

	// Update status and track update time
	status.LastUpdateTime = time.Now()
	s.Status = status
	s.lastError = nil
	s.lastErrorTime = time.Time{}

	slog.Debug("Sequencer status updated successfully",
		"sequencer", s.Config.ID,
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

	if s.conductorClient != nil {
		s.conductorClient = nil
	}

	if s.nodeClient != nil {
		s.nodeClient.Close()
		s.nodeClient = nil
	}
}

// GetClusterMembership returns the cluster membership
func (s *Sequencer) GetClusterMembership(ctx context.Context) (*consensus.ClusterMembership, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.ensureInitialized(); err != nil {
		return nil, err
	}

	membership, err := s.conductorClient.ClusterMembership(ctx)
	if err != nil {
		slog.Error("Failed to get cluster membership",
			"sequencer", s.Config.ID,
			"error", err)
		return nil, fmt.Errorf("failed to get cluster membership for sequencer %s: %w", s.Config.ID, err)
	}
	return membership, nil
}

// Pause pauses the conductor
func (s *Sequencer) Pause(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.ensureInitialized(); err != nil {
		return err
	}

	if err := s.conductorClient.Pause(ctx); err != nil {
		slog.Error("Failed to pause conductor",
			"sequencer", s.Config.ID,
			"error", err)
		return fmt.Errorf("failed to pause conductor for sequencer %s: %w", s.Config.ID, err)
	}
	return nil
}

// Resume resumes the conductor
func (s *Sequencer) Resume(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.ensureInitialized(); err != nil {
		return err
	}

	if err := s.conductorClient.Resume(ctx); err != nil {
		slog.Error("Failed to resume conductor",
			"sequencer", s.Config.ID,
			"error", err)
		return fmt.Errorf("failed to resume conductor for sequencer %s: %w", s.Config.ID, err)
	}
	return nil
}

// TransferLeaderToServer transfers leadership to another server
func (s *Sequencer) TransferLeaderToServer(ctx context.Context, id, addr string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.ensureInitialized(); err != nil {
		return err
	}

	if err := s.conductorClient.TransferLeaderToServer(ctx, id, addr); err != nil {
		slog.Error("Failed to transfer leadership",
			"from", s.Config.ID,
			"to", id,
			"addr", addr,
			"error", err)
		return fmt.Errorf("failed to transfer leadership from %s to %s: %w", s.Config.ID, id, err)
	}
	return nil
}

// OverrideLeader overrides the conductor leader status
func (s *Sequencer) OverrideLeader(ctx context.Context, override bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.ensureInitialized(); err != nil {
		return err
	}

	if err := s.conductorClient.OverrideLeader(ctx, override); err != nil {
		slog.Error("Failed to override leader status",
			"sequencer", s.Config.ID,
			"override", override,
			"error", err)
		return fmt.Errorf("failed to override leader status for sequencer %s (override=%t): %w", s.Config.ID, override, err)
	}
	return nil
}

// AddServerAsVoter adds a server as a voter
func (s *Sequencer) AddServerAsVoter(ctx context.Context, id, addr string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.ensureInitialized(); err != nil {
		return err
	}

	if err := s.conductorClient.AddServerAsVoter(ctx, id, addr, 0); err != nil {
		slog.Error("Failed to add server as voter",
			"sequencer", s.Config.ID,
			"server", id,
			"addr", addr,
			"error", err)
		return fmt.Errorf("failed to add server %s as voter to sequencer %s: %w", id, s.Config.ID, err)
	}
	return nil
}

// AddServerAsNonvoter adds a server as a non-voter
func (s *Sequencer) AddServerAsNonvoter(ctx context.Context, id, addr string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.ensureInitialized(); err != nil {
		return err
	}

	if err := s.conductorClient.AddServerAsNonvoter(ctx, id, addr, 0); err != nil {
		slog.Error("Failed to add server as non-voter",
			"sequencer", s.Config.ID,
			"server", id,
			"addr", addr,
			"error", err)
		return fmt.Errorf("failed to add server %s as non-voter to sequencer %s: %w", id, s.Config.ID, err)
	}
	return nil
}

// RemoveServer removes a server from the cluster
func (s *Sequencer) RemoveServer(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.ensureInitialized(); err != nil {
		return err
	}

	if err := s.conductorClient.RemoveServer(ctx, id, 0); err != nil {
		slog.Error("Failed to remove server",
			"sequencer", s.Config.ID,
			"server", id,
			"error", err)
		return fmt.Errorf("failed to remove server %s from sequencer %s: %w", id, s.Config.ID, err)
	}
	return nil
}

// StopSequencer stops the sequencer
func (s *Sequencer) StopSequencer(ctx context.Context) (common.Hash, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.ensureInitialized(); err != nil {
		return common.Hash{}, err
	}

	hash, err := s.nodeClient.StopSequencer(ctx)
	if err != nil {
		slog.Error("Failed to stop sequencer",
			"sequencer", s.Config.ID,
			"error", err)
		return common.Hash{}, fmt.Errorf("failed to stop sequencer %s: %w", s.Config.ID, err)
	}
	return hash, nil
}

// StartSequencer starts the sequencer with the given hash
func (s *Sequencer) StartSequencer(ctx context.Context, hash common.Hash) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.ensureInitialized(); err != nil {
		return err
	}

	if err := s.nodeClient.StartSequencer(ctx, hash); err != nil {
		slog.Error("Failed to start sequencer",
			"sequencer", s.Config.ID,
			"hash", hash.String(),
			"error", err)
		return fmt.Errorf("failed to start sequencer %s with hash %s: %w", s.Config.ID, hash.String(), err)
	}
	return nil
}

// OverrideNodeLeader overrides the node leader status
func (s *Sequencer) OverrideNodeLeader(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.ensureInitialized(); err != nil {
		return err
	}

	if err := s.nodeClient.OverrideLeader(ctx); err != nil {
		slog.Error("Failed to override node leader",
			"sequencer", s.Config.ID,
			"error", err)
		return fmt.Errorf("failed to override node leader for sequencer %s: %w", s.Config.ID, err)
	}
	return nil
}
