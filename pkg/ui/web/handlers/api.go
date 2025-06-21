package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-chi/chi/v5"
	"github.com/golem-base/seqctl/pkg/app"
	"github.com/golem-base/seqctl/pkg/network"
	"github.com/golem-base/seqctl/pkg/sequencer"
	"github.com/gorilla/websocket"
)

// APIHandler handles API requests
type APIHandler struct {
	app      *app.App
	logger   *slog.Logger
	upgrader websocket.Upgrader
}

// NewAPIHandler creates a new API handler
func NewAPIHandler(application *app.App, logger *slog.Logger) *APIHandler {
	return &APIHandler{
		app:    application,
		logger: logger.With(slog.String("component", "api")),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for now
			},
		},
	}
}

// ErrorResponse represents an error response following RFC 7807
type ErrorResponse struct {
	Type     string         `json:"type"`
	Title    string         `json:"title"`
	Status   int            `json:"status"`
	Detail   string         `json:"detail,omitempty"`
	Instance string         `json:"instance,omitempty"`
	Errors   map[string]any `json:"errors,omitempty"`
}

// NetworkResponse represents a network in API responses
type NetworkResponse struct {
	ID         string              `json:"id"`
	Name       string              `json:"name"`
	Healthy    bool                `json:"healthy"`
	Sequencers []SequencerResponse `json:"sequencers"`
	UpdatedAt  time.Time           `json:"updated_at"`
	Links      NetworkLinks        `json:"_links"`
}

// NetworkLinks represents HATEOAS links for a network
type NetworkLinks struct {
	Self       Link `json:"self"`
	Sequencers Link `json:"sequencers"`
}

// SequencerResponse represents a sequencer in API responses
type SequencerResponse struct {
	ID               string         `json:"id"`
	NetworkID        string         `json:"network_id"`
	RaftAddr         string         `json:"raft_addr"`
	ConductorActive  bool           `json:"conductor_active"`
	ConductorLeader  bool           `json:"conductor_leader"`
	ConductorPaused  bool           `json:"conductor_paused"`
	ConductorStopped bool           `json:"conductor_stopped"`
	SequencerHealthy bool           `json:"sequencer_healthy"`
	SequencerActive  bool           `json:"sequencer_active"`
	UnsafeL2         uint64         `json:"unsafe_l2"`
	Voting           bool           `json:"voting"`
	UpdatedAt        time.Time      `json:"updated_at"`
	Links            SequencerLinks `json:"_links"`
}

// SequencerLinks represents HATEOAS links for a sequencer
type SequencerLinks struct {
	Self           Link  `json:"self"`
	Network        Link  `json:"network"`
	Pause          *Link `json:"pause,omitempty"`
	Resume         *Link `json:"resume,omitempty"`
	TransferLeader *Link `json:"transfer_leader,omitempty"`
	ResignLeader   *Link `json:"resign_leader,omitempty"`
	OverrideLeader *Link `json:"override_leader,omitempty"`
	Halt           *Link `json:"halt,omitempty"`
	ForceActive    *Link `json:"force_active,omitempty"`
	RemoveMember   *Link `json:"remove_member,omitempty"`
	UpdateMember   *Link `json:"update_member,omitempty"`
}

// Link represents a HATEOAS link
type Link struct {
	Href   string `json:"href"`
	Method string `json:"method,omitempty"`
}

// PaginatedResponse represents a paginated response
type PaginatedResponse struct {
	Data       any        `json:"data"`
	Pagination Pagination `json:"pagination"`
}

// Pagination represents pagination metadata
type Pagination struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// sendJSON sends a JSON response
func (h *APIHandler) sendJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("Failed to encode response", slog.String("error", err.Error()))
	}
}

// sendError sends an error response following REST best practices
func (h *APIHandler) sendError(w http.ResponseWriter, status int, title string, detail string) {
	h.logger.Error("API error",
		slog.Int("status", status),
		slog.String("title", title),
		slog.String("detail", detail),
	)

	errorType := "about:blank"
	switch status {
	case http.StatusBadRequest:
		errorType = "/errors/bad-request"
	case http.StatusNotFound:
		errorType = "/errors/not-found"
	case http.StatusConflict:
		errorType = "/errors/conflict"
	case http.StatusUnprocessableEntity:
		errorType = "/errors/validation-failed"
	case http.StatusInternalServerError:
		errorType = "/errors/internal-server-error"
	}

	h.sendJSON(w, status, ErrorResponse{
		Type:   errorType,
		Title:  title,
		Status: status,
		Detail: detail,
	})
}

// ListNetworks returns all available networks
// @Summary List all networks
// @Description Get a list of all sequencer networks in the environment
// @Tags Networks
// @Accept json
// @Produce json
// @Success 200 {array} NetworkResponse "List of networks"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /networks [get]
func (h *APIHandler) ListNetworks(w http.ResponseWriter, r *http.Request) {
	networks, err := h.app.ListNetworks(r.Context())
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "Failed to list networks", err.Error())
		return
	}

	response := make([]NetworkResponse, 0, len(networks))
	for _, net := range networks {
		response = append(response, h.networkToResponse(net))
	}

	h.sendJSON(w, http.StatusOK, response)
}

// GetNetwork returns a specific network
// @Summary Get network details
// @Description Get detailed information about a specific network
// @Tags Networks
// @Accept json
// @Produce json
// @Param network path string true "Network name"
// @Success 200 {object} NetworkResponse "Network details"
// @Failure 404 {object} ErrorResponse "Network not found"
// @Router /networks/{network} [get]
func (h *APIHandler) GetNetwork(w http.ResponseWriter, r *http.Request) {
	networkName := chi.URLParam(r, "network")

	net, err := h.app.GetNetwork(r.Context(), networkName)
	if err != nil {
		h.sendError(w, http.StatusNotFound, "Network not found",
			fmt.Sprintf("Network '%s' does not exist", networkName))
		return
	}

	h.sendJSON(w, http.StatusOK, h.networkToResponse(net))
}

// GetSequencers returns all sequencers in a network
// @Summary List network sequencers
// @Description Get all sequencers belonging to a specific network
// @Tags Networks,Sequencers
// @Accept json
// @Produce json
// @Param network path string true "Network name"
// @Success 200 {array} SequencerResponse "List of sequencers"
// @Failure 404 {object} ErrorResponse "Network not found"
// @Router /networks/{network}/sequencers [get]
func (h *APIHandler) GetSequencers(w http.ResponseWriter, r *http.Request) {
	networkName := chi.URLParam(r, "network")

	net, err := h.app.GetNetwork(r.Context(), networkName)
	if err != nil {
		h.sendError(w, http.StatusNotFound, "Network not found",
			fmt.Sprintf("Network '%s' does not exist", networkName))
		return
	}

	sequencers := make([]SequencerResponse, 0, len(net.Sequencers()))
	for _, seq := range net.Sequencers() {
		sequencers = append(sequencers, h.sequencerToResponse(seq, networkName))
	}

	h.sendJSON(w, http.StatusOK, sequencers)
}

// PauseSequencer pauses a sequencer's conductor
// @Summary Pause conductor
// @Description Pause the conductor service on a sequencer, stopping it from participating in consensus
// @Tags Actions
// @Accept json
// @Produce json
// @Param id path string true "Sequencer ID"
// @Success 200 {object} SequencerResponse "Updated sequencer state"
// @Failure 404 {object} ErrorResponse "Sequencer not found"
// @Failure 409 {object} ErrorResponse "Conductor already paused"
// @Failure 500 {object} ErrorResponse "Operation failed"
// @Router /sequencers/{id}/pause [post]
func (h *APIHandler) PauseSequencer(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	seq, network, err := h.getSequencer(ctx, chi.URLParam(r, "id"))
	if err != nil {
		h.sendError(w, http.StatusNotFound, "Sequencer not found", err.Error())
		return
	}

	if !seq.ConductorActive() {
		h.sendError(w, http.StatusConflict, "Invalid state",
			"Conductor is already paused")
		return
	}

	if err := seq.Pause(ctx); err != nil {
		h.sendError(w, http.StatusInternalServerError, "Operation failed",
			fmt.Sprintf("Failed to pause conductor: %v", err))
		return
	}

	// Return updated sequencer state
	// State will be updated on next refresh
	h.sendJSON(w, http.StatusOK, h.sequencerToResponse(seq, network))
}

// ResumeSequencer resumes a sequencer's conductor
// @Summary Resume conductor
// @Description Resume the conductor service on a sequencer, allowing it to participate in consensus again
// @Tags Actions
// @Accept json
// @Produce json
// @Param id path string true "Sequencer ID"
// @Success 200 {object} SequencerResponse "Updated sequencer state"
// @Failure 404 {object} ErrorResponse "Sequencer not found"
// @Failure 409 {object} ErrorResponse "Conductor already active"
// @Failure 500 {object} ErrorResponse "Operation failed"
// @Router /sequencers/{id}/resume [post]
func (h *APIHandler) ResumeSequencer(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	seq, network, err := h.getSequencer(ctx, chi.URLParam(r, "id"))
	if err != nil {
		h.sendError(w, http.StatusNotFound, "Sequencer not found", err.Error())
		return
	}

	if seq.ConductorActive() {
		h.sendError(w, http.StatusConflict, "Invalid state",
			"Conductor is already active")
		return
	}

	if err := seq.Resume(ctx); err != nil {
		h.sendError(w, http.StatusInternalServerError, "Operation failed",
			fmt.Sprintf("Failed to resume conductor: %v", err))
		return
	}

	// Return updated sequencer state
	// State will be updated on next refresh
	h.sendJSON(w, http.StatusOK, h.sequencerToResponse(seq, network))
}

// TransferLeaderRequest represents the request body for leader transfer
type TransferLeaderRequest struct {
	TargetID   string `json:"target_id" validate:"required"`
	TargetAddr string `json:"target_addr" validate:"required"`
}

// TransferLeader transfers leadership to another sequencer
// @Summary Transfer leadership
// @Description Transfer Raft leadership from the current leader to a specified target sequencer
// @Tags Actions
// @Accept json
// @Produce json
// @Param id path string true "Sequencer ID"
// @Param request body TransferLeaderRequest true "Transfer target details"
// @Success 202 {object} map[string]interface{} "Leadership transfer initiated"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 404 {object} ErrorResponse "Sequencer not found"
// @Failure 409 {object} ErrorResponse "Cannot transfer from current leader"
// @Failure 422 {object} ErrorResponse "Validation failed"
// @Failure 500 {object} ErrorResponse "Operation failed"
// @Router /sequencers/{id}/transfer-leader [post]
func (h *APIHandler) TransferLeader(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	seq, _, err := h.getSequencer(ctx, chi.URLParam(r, "id"))
	if err != nil {
		h.sendError(w, http.StatusNotFound, "Sequencer not found", err.Error())
		return
	}

	if seq.ConductorLeader() {
		h.sendError(w, http.StatusConflict, "Invalid state",
			"Cannot transfer leadership from current leader")
		return
	}

	var req TransferLeaderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if req.TargetID == "" || req.TargetAddr == "" {
		h.sendError(w, http.StatusUnprocessableEntity, "Validation failed",
			"target_id and target_addr are required")
		return
	}

	if err := seq.TransferLeaderToServer(ctx, req.TargetID, req.TargetAddr); err != nil {
		h.sendError(w, http.StatusInternalServerError, "Operation failed",
			fmt.Sprintf("Failed to transfer leadership: %v", err))
		return
	}

	h.sendJSON(w, http.StatusAccepted, map[string]any{
		"message":     "Leadership transfer initiated",
		"target_id":   req.TargetID,
		"target_addr": req.TargetAddr,
	})
}

// ResignLeader causes the current leader to resign
// @Summary Resign leadership
// @Description Make the current leader sequencer resign, triggering a new leader election
// @Tags Actions
// @Accept json
// @Produce json
// @Param id path string true "Sequencer ID"
// @Success 202 {object} SequencerResponse "Leadership resignation accepted"
// @Failure 404 {object} ErrorResponse "Sequencer not found"
// @Failure 409 {object} ErrorResponse "Sequencer is not the current leader"
// @Failure 500 {object} ErrorResponse "Operation failed"
// @Router /sequencers/{id}/resign-leader [post]
func (h *APIHandler) ResignLeader(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	seq, network, err := h.getSequencer(ctx, chi.URLParam(r, "id"))
	if err != nil {
		h.sendError(w, http.StatusNotFound, "Sequencer not found", err.Error())
		return
	}

	if !seq.ConductorLeader() {
		h.sendError(w, http.StatusConflict, "Invalid state",
			"Sequencer is not the current leader")
		return
	}

	if err := seq.TransferLeader(ctx); err != nil {
		h.sendError(w, http.StatusInternalServerError, "Operation failed",
			fmt.Sprintf("Failed to resign leadership: %v", err))
		return
	}

	h.sendJSON(w, http.StatusAccepted, h.sequencerToResponse(seq, network))
}

// OverrideLeaderRequest represents the request body for leader override
type OverrideLeaderRequest struct {
	Override bool `json:"override"`
}

// OverrideLeader overrides the leader status
// @Summary Override leader status
// @Description Force override the leader status of a sequencer (WARNING: Can cause split-brain)
// @Tags Actions
// @Accept json
// @Produce json
// @Param id path string true "Sequencer ID"
// @Param request body OverrideLeaderRequest true "Override configuration"
// @Success 200 {object} SequencerResponse "Leader status overridden"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 404 {object} ErrorResponse "Sequencer not found"
// @Failure 500 {object} ErrorResponse "Operation failed"
// @Router /sequencers/{id}/override-leader [post]
func (h *APIHandler) OverrideLeader(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	seq, network, err := h.getSequencer(ctx, chi.URLParam(r, "id"))
	if err != nil {
		h.sendError(w, http.StatusNotFound, "Sequencer not found", err.Error())
		return
	}

	var req OverrideLeaderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if err := seq.OverrideLeader(ctx, req.Override); err != nil {
		h.sendError(w, http.StatusInternalServerError, "Operation failed",
			fmt.Sprintf("Failed to override leader: %v", err))
		return
	}

	h.sendJSON(w, http.StatusOK, h.sequencerToResponse(seq, network))
}

// HaltSequencer halts a sequencer
// @Summary Halt sequencer
// @Description Stop a sequencer from processing transactions
// @Tags Actions
// @Accept json
// @Produce json
// @Param id path string true "Sequencer ID"
// @Success 200 {object} SequencerResponse "Sequencer halted"
// @Failure 404 {object} ErrorResponse "Sequencer not found"
// @Failure 409 {object} ErrorResponse "Sequencer already halted"
// @Failure 500 {object} ErrorResponse "Operation failed"
// @Router /sequencers/{id}/halt [post]
func (h *APIHandler) HaltSequencer(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	seq, network, err := h.getSequencer(ctx, chi.URLParam(r, "id"))
	if err != nil {
		h.sendError(w, http.StatusNotFound, "Sequencer not found", err.Error())
		return
	}

	if !seq.SequencerActive() {
		h.sendError(w, http.StatusConflict, "Invalid state",
			"Sequencer is already halted")
		return
	}

	if _, err := seq.StopSequencer(ctx); err != nil {
		h.sendError(w, http.StatusInternalServerError, "Operation failed",
			fmt.Sprintf("Failed to halt sequencer: %v", err))
		return
	}

	// Return updated sequencer state
	// State will be updated on next refresh
	h.sendJSON(w, http.StatusOK, h.sequencerToResponse(seq, network))
}

// ForceActiveRequest represents the request body for forcing a sequencer active
type ForceActiveRequest struct {
	BlockHash string `json:"block_hash,omitempty"`
}

// ForceActive forces a sequencer to become active
// @Summary Force sequencer active
// @Description Force a sequencer to become the active sequencer (WARNING: Use only in emergencies)
// @Tags Actions
// @Accept json
// @Produce json
// @Param id path string true "Sequencer ID"
// @Param request body ForceActiveRequest false "Optional block hash to start from"
// @Success 200 {object} SequencerResponse "Sequencer activated"
// @Failure 404 {object} ErrorResponse "Sequencer not found"
// @Failure 409 {object} ErrorResponse "Sequencer already active"
// @Failure 500 {object} ErrorResponse "Operation failed"
// @Router /sequencers/{id}/force-active [post]
func (h *APIHandler) ForceActive(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	seq, network, err := h.getSequencer(ctx, chi.URLParam(r, "id"))
	if err != nil {
		h.sendError(w, http.StatusNotFound, "Sequencer not found", err.Error())
		return
	}

	if seq.SequencerActive() {
		h.sendError(w, http.StatusConflict, "Invalid state",
			"Sequencer is already active")
		return
	}

	var req ForceActiveRequest
	// Allow empty body - will use zero hash
	json.NewDecoder(r.Body).Decode(&req)

	var hash common.Hash
	if req.BlockHash != "" {
		hash = common.HexToHash(req.BlockHash)
	}

	if err := seq.StartSequencer(ctx, hash); err != nil {
		h.sendError(w, http.StatusInternalServerError, "Operation failed",
			fmt.Sprintf("Failed to activate sequencer: %v", err))
		return
	}

	// Return updated sequencer state
	// State will be updated on next refresh
	h.sendJSON(w, http.StatusOK, h.sequencerToResponse(seq, network))
}

// RemoveMemberRequest represents the request body for removing a member
type RemoveMemberRequest struct {
	ServerID string `json:"server_id" validate:"required"`
}

// RemoveFromCluster removes a sequencer from the cluster
// @Summary Remove server from cluster
// @Description Remove a server from the Raft cluster membership
// @Tags Actions
// @Accept json
// @Produce json
// @Param id path string true "Sequencer ID (must be leader)"
// @Param request body RemoveMemberRequest true "Server to remove"
// @Success 204 "Server removed successfully"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 404 {object} ErrorResponse "Sequencer not found"
// @Failure 422 {object} ErrorResponse "Validation failed"
// @Failure 500 {object} ErrorResponse "Operation failed"
// @Router /sequencers/{id}/membership [delete]
func (h *APIHandler) RemoveFromCluster(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	seq, _, err := h.getSequencer(ctx, chi.URLParam(r, "id"))
	if err != nil {
		h.sendError(w, http.StatusNotFound, "Sequencer not found", err.Error())
		return
	}

	var req RemoveMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if req.ServerID == "" {
		h.sendError(w, http.StatusUnprocessableEntity, "Validation failed",
			"server_id is required")
		return
	}

	if err := seq.RemoveServer(ctx, req.ServerID); err != nil {
		h.sendError(w, http.StatusInternalServerError, "Operation failed",
			fmt.Sprintf("Failed to remove server from cluster: %v", err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UpdateMembershipRequest represents the request body for updating membership
type UpdateMembershipRequest struct {
	ServerID   string `json:"server_id" validate:"required"`
	ServerAddr string `json:"server_addr" validate:"required"`
	Voting     bool   `json:"voting"`
}

// UpdateMembership updates cluster membership
// @Summary Update cluster membership
// @Description Add a new server to the Raft cluster as either a voting or non-voting member
// @Tags Actions
// @Accept json
// @Produce json
// @Param id path string true "Sequencer ID (must be leader)"
// @Param request body UpdateMembershipRequest true "New member details"
// @Success 200 {object} SequencerResponse "Membership updated"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 404 {object} ErrorResponse "Sequencer not found"
// @Failure 422 {object} ErrorResponse "Validation failed"
// @Failure 500 {object} ErrorResponse "Operation failed"
// @Router /sequencers/{id}/membership [put]
func (h *APIHandler) UpdateMembership(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	seq, network, err := h.getSequencer(ctx, chi.URLParam(r, "id"))
	if err != nil {
		h.sendError(w, http.StatusNotFound, "Sequencer not found", err.Error())
		return
	}

	var req UpdateMembershipRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if req.ServerID == "" || req.ServerAddr == "" {
		h.sendError(w, http.StatusUnprocessableEntity, "Validation failed",
			"server_id and server_addr are required")
		return
	}

	if req.Voting {
		err = seq.AddServerAsVoter(ctx, req.ServerID, req.ServerAddr)
	} else {
		err = seq.AddServerAsNonvoter(ctx, req.ServerID, req.ServerAddr)
	}

	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "Operation failed",
			fmt.Sprintf("Failed to update membership: %v", err))
		return
	}

	h.sendJSON(w, http.StatusOK, h.sequencerToResponse(seq, network))
}

// WebSocket handles WebSocket connections for real-time updates
func (h *APIHandler) WebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("WebSocket upgrade failed", slog.String("error", err.Error()))
		return
	}
	defer conn.Close()

	// TODO: Implement WebSocket handling for real-time updates
	h.logger.Info("WebSocket connection established")
}

// Helper methods

func (h *APIHandler) getSequencer(ctx context.Context, sequencerID string) (*sequencer.Sequencer, string, error) {
	networks, err := h.app.ListNetworks(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list networks: %w", err)
	}

	for _, net := range networks {
		for _, seq := range net.Sequencers() {
			if seq.ID() == sequencerID {
				return seq, net.Name(), nil
			}
		}
	}

	return nil, "", fmt.Errorf("sequencer not found: %s", sequencerID)
}

func (h *APIHandler) networkToResponse(net *network.Network) NetworkResponse {
	sequencers := make([]SequencerResponse, 0, len(net.Sequencers()))
	for _, seq := range net.Sequencers() {
		sequencers = append(sequencers, h.sequencerToResponse(seq, net.Name()))
	}

	return NetworkResponse{
		ID:         net.Name(),
		Name:       net.Name(),
		Healthy:    net.IsHealthy(),
		Sequencers: sequencers,
		UpdatedAt:  net.UpdatedAt(),
		Links: NetworkLinks{
			Self:       Link{Href: fmt.Sprintf("/api/v1/networks/%s", net.Name())},
			Sequencers: Link{Href: fmt.Sprintf("/api/v1/networks/%s/sequencers", net.Name())},
		},
	}
}

func (h *APIHandler) sequencerToResponse(seq *sequencer.Sequencer, networkName string) SequencerResponse {
	resp := SequencerResponse{
		ID:               seq.ID(),
		NetworkID:        networkName,
		RaftAddr:         seq.RaftAddr(),
		ConductorActive:  seq.ConductorActive(),
		ConductorLeader:  seq.ConductorLeader(),
		ConductorPaused:  seq.ConductorPaused(),
		ConductorStopped: seq.ConductorStopped(),
		SequencerHealthy: seq.SequencerHealthy(),
		SequencerActive:  seq.SequencerActive(),
		UnsafeL2:         seq.UnsafeL2(),
		Voting:           seq.Voting(),
		UpdatedAt:        time.Now(),
		Links: SequencerLinks{
			Self:    Link{Href: fmt.Sprintf("/api/v1/sequencers/%s", seq.ID())},
			Network: Link{Href: fmt.Sprintf("/api/v1/networks/%s", networkName)},
		},
	}

	// Add action links based on current state
	baseURL := fmt.Sprintf("/api/v1/sequencers/%s", seq.ID())

	if seq.ConductorActive() {
		resp.Links.Pause = &Link{Href: baseURL + "/pause", Method: "POST"}
	} else {
		resp.Links.Resume = &Link{Href: baseURL + "/resume", Method: "POST"}
	}

	if !seq.ConductorLeader() {
		resp.Links.TransferLeader = &Link{Href: baseURL + "/transfer-leader", Method: "POST"}
	}

	resp.Links.OverrideLeader = &Link{Href: baseURL + "/override-leader", Method: "POST"}

	if seq.SequencerActive() {
		resp.Links.Halt = &Link{Href: baseURL + "/halt", Method: "POST"}
	} else {
		resp.Links.ForceActive = &Link{Href: baseURL + "/force-active", Method: "POST"}
	}

	if seq.ConductorLeader() {
		resp.Links.ResignLeader = &Link{Href: baseURL + "/resign-leader", Method: "POST"}
		resp.Links.UpdateMember = &Link{Href: baseURL + "/membership", Method: "PUT"}
		resp.Links.RemoveMember = &Link{Href: baseURL + "/membership", Method: "DELETE"}
	}

	return resp
}
