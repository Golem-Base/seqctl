package rpc

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/common"
	ethrpc "github.com/ethereum/go-ethereum/rpc"

	"github.com/ethereum-optimism/optimism/op-conductor/consensus"
	cdtrpc "github.com/ethereum-optimism/optimism/op-conductor/rpc"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	seqrpc "github.com/ethereum-optimism/optimism/op-service/sources"
)

// Client provides a unified interface for conductor and node RPC operations
type Client struct {
	conductorURL string
	nodeURL      string
	timeout      time.Duration
	logger       *slog.Logger
	httpClient   *http.Client
	conductorRPC *ethrpc.Client
	sequencerRPC *ethrpc.Client
	conductor    *cdtrpc.APIClient
	sequencer    *seqrpc.RollupClient
}

// NewClient creates a new RPC client with a default context
func NewClient(conductorURL, nodeURL string, opts ...ClientOption) (*Client, error) {
	return NewClientWithContext(context.Background(), conductorURL, nodeURL, opts...)
}

// NewClientWithContext creates a new RPC client with a custom context for initialization
func NewClientWithContext(ctx context.Context, conductorURL, nodeURL string, opts ...ClientOption) (*Client, error) {
	// Validate URLs
	if conductorURL == "" {
		return nil, fmt.Errorf("conductor URL is required")
	}
	if nodeURL == "" {
		return nil, fmt.Errorf("node URL is required")
	}

	c := &Client{
		conductorURL: conductorURL,
		nodeURL:      nodeURL,
		timeout:      30 * time.Second,
		logger:       slog.Default().With(slog.String("component", "rpc-client")),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	// Apply options
	for _, opt := range opts {
		opt(c)
	}

	// Initialize connections
	if err := c.initialize(ctx); err != nil {
		return nil, err
	}

	return c, nil
}

// ClientOption configures the client
type ClientOption func(*Client)

// WithTimeout sets the timeout for RPC calls
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.timeout = timeout
	}
}

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithLogger sets a custom logger
func WithLogger(logger *slog.Logger) ClientOption {
	return func(c *Client) {
		c.logger = logger.With(slog.String("component", "rpc-client"))
	}
}

// initialize creates the RPC connections
func (c *Client) initialize(ctx context.Context) error {
	var err error

	// Use DialOptions with WithHTTPClient (non-deprecated method)
	c.conductorRPC, err = ethrpc.DialOptions(ctx, c.conductorURL, ethrpc.WithHTTPClient(c.httpClient))
	if err != nil {
		return fmt.Errorf("dial conductor: %w", err)
	}
	c.conductor = cdtrpc.NewAPIClient(c.conductorRPC)

	// Initialize node - reuse connection if same URL
	if c.nodeURL == c.conductorURL {
		c.sequencerRPC = c.conductorRPC
		c.sequencer = seqrpc.NewRollupClient(NewRPCAdapter(c.sequencerRPC))
	} else {
		c.sequencerRPC, err = ethrpc.DialOptions(ctx, c.nodeURL, ethrpc.WithHTTPClient(c.httpClient))
		if err != nil {
			c.conductorRPC.Close()
			return fmt.Errorf("dial node: %w", err)
		}
		c.sequencer = seqrpc.NewRollupClient(NewRPCAdapter(c.sequencerRPC))
	}

	return nil
}

// withTimeout creates a context with timeout
func (c *Client) withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, c.timeout)
}

// --- Conductor Status Methods ---

// Active returns whether the conductor is active
func (c *Client) Active(ctx context.Context) (bool, error) {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()
	return c.conductor.Active(ctx)
}

// Leader returns whether the conductor is the leader
func (c *Client) Leader(ctx context.Context) (bool, error) {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()
	return c.conductor.Leader(ctx)
}

// Paused returns whether the conductor is paused
func (c *Client) Paused(ctx context.Context) (bool, error) {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()
	return c.conductor.Paused(ctx)
}

// Stopped returns whether the conductor is stopped
func (c *Client) Stopped(ctx context.Context) (bool, error) {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()
	return c.conductor.Stopped(ctx)
}

// SequencerHealthy returns whether the sequencer is healthy
func (c *Client) SequencerHealthy(ctx context.Context) (bool, error) {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()
	return c.conductor.SequencerHealthy(ctx)
}

// --- Conductor Control Methods ---

// Pause pauses the conductor
func (c *Client) Pause(ctx context.Context) error {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()
	return c.conductor.Pause(ctx)
}

// Resume resumes the conductor
func (c *Client) Resume(ctx context.Context) error {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()
	return c.conductor.Resume(ctx)
}

// --- Conductor Leadership Methods ---

// TransferLeader transfers leadership to another node
func (c *Client) TransferLeader(ctx context.Context) error {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()
	return c.conductor.TransferLeader(ctx)
}

// TransferLeaderToServer transfers leadership to a specific server
func (c *Client) TransferLeaderToServer(ctx context.Context, id, addr string) error {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()
	return c.conductor.TransferLeaderToServer(ctx, id, addr)
}

// OverrideLeader overrides the leader status
func (c *Client) OverrideLeader(ctx context.Context, override bool) error {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()
	return c.conductor.OverrideLeader(ctx, override)
}

// LeaderWithID returns the current leader's server info
func (c *Client) LeaderWithID(ctx context.Context) (*consensus.ServerInfo, error) {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()
	return c.conductor.LeaderWithID(ctx)
}

// --- Conductor Cluster Management Methods ---

// ClusterMembership returns the current cluster membership
func (c *Client) ClusterMembership(ctx context.Context) (*consensus.ClusterMembership, error) {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()
	return c.conductor.ClusterMembership(ctx)
}

// AddServerAsVoter adds a server as a voting member
func (c *Client) AddServerAsVoter(ctx context.Context, id, addr string, prevIndex uint64) error {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()
	return c.conductor.AddServerAsVoter(ctx, id, addr, prevIndex)
}

// AddServerAsNonvoter adds a server as a non-voting member
func (c *Client) AddServerAsNonvoter(ctx context.Context, id, addr string, prevIndex uint64) error {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()
	return c.conductor.AddServerAsNonvoter(ctx, id, addr, prevIndex)
}

// RemoveServer removes a server from the cluster
func (c *Client) RemoveServer(ctx context.Context, id string, prevIndex uint64) error {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()
	return c.conductor.RemoveServer(ctx, id, prevIndex)
}

// --- Node Status Methods ---

// SequencerActive returns whether the sequencer is active
func (c *Client) SequencerActive(ctx context.Context) (bool, error) {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()
	return c.sequencer.SequencerActive(ctx)
}

// SyncStatus returns the sync status of the node
func (c *Client) SyncStatus(ctx context.Context) (*eth.SyncStatus, error) {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()
	return c.sequencer.SyncStatus(ctx)
}

// --- Node Control Methods ---

// StopSequencer stops the sequencer and returns the stop hash
func (c *Client) StopSequencer(ctx context.Context) (common.Hash, error) {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()
	return c.sequencer.StopSequencer(ctx)
}

// StartSequencer starts the sequencer with the given hash
func (c *Client) StartSequencer(ctx context.Context, hash common.Hash) error {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()
	return c.sequencer.StartSequencer(ctx, hash)
}

// OverrideNodeLeader overrides the node's leader status
func (c *Client) OverrideNodeLeader(ctx context.Context) error {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()
	return c.sequencer.OverrideLeader(ctx)
}

// Close closes the client connections
func (c *Client) Close() error {
	if c.conductorRPC != nil && c.conductorRPC != c.sequencerRPC {
		c.conductorRPC.Close()
	}
	if c.sequencerRPC != nil {
		c.sequencerRPC.Close()
	}
	if c.sequencer != nil {
		c.sequencer.Close()
	}
	return nil
}
