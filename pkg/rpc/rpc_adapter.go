package rpc

import (
	"context"

	"github.com/ethereum/go-ethereum"
	ethrpc "github.com/ethereum/go-ethereum/rpc"
)

// RPCAdapter adapts an ethrpc.Client to the RPC interface required by sources.NewRollupClient
type RPCAdapter struct {
	client *ethrpc.Client
}

// NewRPCAdapter creates a new RPC adapter from an Ethereum RPC client
func NewRPCAdapter(client *ethrpc.Client) *RPCAdapter {
	return &RPCAdapter{client: client}
}

// CallContext performs a JSON-RPC call with the given context
func (a *RPCAdapter) CallContext(ctx context.Context, result any, method string, args ...any) error {
	return a.client.CallContext(ctx, result, method, args...)
}

// BatchCallContext performs multiple JSON-RPC calls as a batch
func (a *RPCAdapter) BatchCallContext(ctx context.Context, b []ethrpc.BatchElem) error {
	return a.client.BatchCallContext(ctx, b)
}

// Subscribe implements the Subscribe method required by the RPC interface
func (a *RPCAdapter) Subscribe(ctx context.Context, namespace string, channel any, args ...any) (ethereum.Subscription, error) {
	return a.client.Subscribe(ctx, namespace, channel, args...)
}

// Close closes the underlying RPC client connection
func (a *RPCAdapter) Close() {
	a.client.Close()
}
