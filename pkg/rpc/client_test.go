package rpc

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

// mockRPCServer creates a test server that responds to RPC calls
func mockRPCServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parse JSON-RPC request
		var req struct {
			Method string          `json:"method"`
			ID     json.RawMessage `json:"id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Logf("Failed to decode request: %v", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		// Prepare response
		resp := map[string]any{
			"jsonrpc": "2.0",
			"id":      req.ID,
		}

		// Handle different methods
		switch req.Method {
		case "web3_clientVersion":
			resp["result"] = "test-client/v1.0.0"
		case "conductor_active":
			resp["result"] = true
		case "conductor_leader":
			resp["result"] = false
		case "conductor_paused":
			resp["result"] = false
		case "optimism_syncStatus":
			resp["result"] = map[string]any{
				"head_l1":         map[string]any{"number": "0x100"},
				"safe_l1":         map[string]any{"number": "0xff"},
				"finalized_l1":    map[string]any{"number": "0xfe"},
				"unsafe_l2":       map[string]any{"number": "0x200"},
				"safe_l2":         map[string]any{"number": "0x1ff"},
				"finalized_l2":    map[string]any{"number": "0x1fe"},
				"pending_safe_l2": map[string]any{"number": "0x1ff"},
			}
		default:
			resp["error"] = map[string]any{
				"code":    -32601,
				"message": "Method not found",
			}
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
}

func TestClient_Initialization(t *testing.T) {
	server := mockRPCServer(t)
	defer server.Close()

	// Test successful initialization
	client, err := NewClient(server.URL, server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Test basic operation
	ctx := context.Background()
	active, err := client.Active(ctx)
	if err != nil {
		t.Fatalf("Failed to check active status: %v", err)
	}
	if !active {
		t.Error("Expected active to be true")
	}
}

func TestClient_DifferentURLs(t *testing.T) {
	server1 := mockRPCServer(t)
	defer server1.Close()

	server2 := mockRPCServer(t)
	defer server2.Close()

	client, err := NewClient(server1.URL, server2.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()
}

func TestClient_WithOptions(t *testing.T) {
	server := mockRPCServer(t)
	defer server.Close()

	customTimeout := 5 * time.Second
	customHTTPClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	client, err := NewClient(
		server.URL,
		server.URL,
		WithTimeout(customTimeout),
		WithHTTPClient(customHTTPClient),
	)
	if err != nil {
		t.Fatalf("Failed to create client with options: %v", err)
	}
	defer client.Close()

	if client.timeout != customTimeout {
		t.Errorf("Expected timeout %v, got %v", customTimeout, client.timeout)
	}
	if client.httpClient != customHTTPClient {
		t.Error("Custom HTTP client not set")
	}
}

func TestClient_ConcurrentAccess(t *testing.T) {
	server := mockRPCServer(t)
	defer server.Close()

	client, err := NewClient(server.URL, server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Launch concurrent operations
	const numGoroutines = 10
	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines)

	for range numGoroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := client.Active(ctx)
			errors <- err
		}()
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		if err != nil {
			t.Errorf("Concurrent access error: %v", err)
		}
	}
}

func TestClient_Timeout(t *testing.T) {
	// Server that delays response
	slowServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer slowServer.Close()

	// This should fail during initialization due to timeout
	_, err := NewClient(
		slowServer.URL,
		slowServer.URL,
		WithTimeout(100*time.Millisecond),
	)
	if err == nil {
		t.Error("Expected timeout error during initialization")
	}
}

func TestClient_InvalidURL(t *testing.T) {
	_, err := NewClient("invalid://url", "invalid://url")
	if err == nil {
		t.Error("Expected error with invalid URL")
	}
}
