package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/chainlaunch/chainlaunch/pkg/errors"
	"github.com/chainlaunch/chainlaunch/pkg/http/response"
	"github.com/chainlaunch/chainlaunch/pkg/logger"
	"github.com/chainlaunch/chainlaunch/pkg/nodes/service"
	"github.com/chainlaunch/chainlaunch/pkg/nodes/types"
	"github.com/go-chi/chi/v5"
)

type NodeHandler struct {
	service *service.NodeService
	logger  *logger.Logger
}

func NewNodeHandler(service *service.NodeService, logger *logger.Logger) *NodeHandler {
	return &NodeHandler{
		service: service,
		logger:  logger,
	}
}

// Add these types for the response structures
type NodeEventResponse struct {
	ID        int64       `json:"id"`
	NodeID    int64       `json:"node_id"`
	Type      string      `json:"type"`
	Data      interface{} `json:"data,omitempty"`
	CreatedAt time.Time   `json:"created_at"`
}

type PaginatedNodeEventsResponse struct {
	Items []NodeEventResponse `json:"items"`
	Total int64               `json:"total"`
	Page  int                 `json:"page"`
}

// RegisterRoutes registers the node routes
func (h *NodeHandler) RegisterRoutes(r chi.Router) {
	r.Route("/nodes", func(r chi.Router) {
		r.Post("/", response.Middleware(h.CreateNode))
		r.Get("/", response.Middleware(h.ListNodes))
		r.Get("/platform/{platform}", response.Middleware(h.ListNodesByPlatform))
		r.Get("/defaults/fabric-peer", response.Middleware(h.GetFabricPeerDefaults))
		r.Get("/defaults/fabric-orderer", response.Middleware(h.GetFabricOrdererDefaults))
		r.Get("/defaults/fabric", response.Middleware(h.GetNodesDefaults))
		r.Get("/defaults/besu-node", response.Middleware(h.GetBesuNodeDefaults))
		r.Get("/{id}", response.Middleware(h.GetNode))
		r.Post("/{id}/start", response.Middleware(h.StartNode))
		r.Post("/{id}/stop", response.Middleware(h.StopNode))
		r.Post("/{id}/restart", response.Middleware(h.RestartNode))
		r.Delete("/{id}", response.Middleware(h.DeleteNode))
		r.Get("/{id}/logs", h.TailLogs)
		r.Get("/{id}/events", response.Middleware(h.GetNodeEvents))
	})
}

// CreateNode godoc
// @Summary Create a new node
// @Description Create a new node with the specified configuration
// @Tags nodes
// @Accept json
// @Produce json
// @Param request body CreateNodeRequest true "Node creation request"
// @Success 201 {object} NodeResponse
// @Failure 400 {object} response.ErrorResponse "Validation error"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /nodes [post]
func (h *NodeHandler) CreateNode(w http.ResponseWriter, r *http.Request) error {
	var req CreateNodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return errors.NewValidationError("invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Validate request
	if req.Name == "" {
		return errors.NewValidationError("name is required", nil)
	}

	if req.BlockchainPlatform == "" {
		return errors.NewValidationError("blockchain platform is required", nil)
	}

	if !isValidPlatform(types.BlockchainPlatform(req.BlockchainPlatform)) {
		return errors.NewValidationError("invalid blockchain platform", map[string]interface{}{
			"valid_platforms": []string{string(types.PlatformFabric), string(types.PlatformBesu)},
		})
	}

	serviceReq := service.CreateNodeRequest{
		Name:               req.Name,
		BlockchainPlatform: req.BlockchainPlatform,
		FabricPeer:         req.FabricPeer,
		FabricOrderer:      req.FabricOrderer,
		BesuNode:           req.BesuNode,
	}

	node, err := h.service.CreateNode(r.Context(), serviceReq)
	if err != nil {
		return errors.NewInternalError("failed to create node", err, nil)
	}

	return response.WriteJSON(w, http.StatusCreated, toNodeResponse(node))
}

// GetNode godoc
// @Summary Get a node
// @Description Get a node by ID
// @Tags nodes
// @Accept json
// @Produce json
// @Param id path int true "Node ID"
// @Success 200 {object} NodeResponse
// @Failure 400 {object} response.ErrorResponse "Validation error"
// @Failure 404 {object} response.ErrorResponse "Node not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /nodes/{id} [get]
func (h *NodeHandler) GetNode(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		return errors.NewValidationError("invalid node ID", map[string]interface{}{
			"error": err.Error(),
		})
	}

	node, err := h.service.GetNode(r.Context(), id)
	if err != nil {
		if errors.IsType(err, errors.NotFoundError) {
			return errors.NewNotFoundError("node not found", nil)
		}
		return errors.NewInternalError("failed to get node", err, nil)
	}

	return response.WriteJSON(w, http.StatusOK, toNodeResponse(node))
}

// ListNodes godoc
// @Summary List all nodes
// @Description Get a paginated list of nodes with optional platform filter
// @Tags nodes
// @Accept json
// @Produce json
// @Param platform query string false "Filter by blockchain platform"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} PaginatedNodesResponse
// @Failure 400 {object} response.ErrorResponse "Validation error"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /nodes [get]
func (h *NodeHandler) ListNodes(w http.ResponseWriter, r *http.Request) error {
	var platform *types.BlockchainPlatform
	if platformStr := r.URL.Query().Get("platform"); platformStr != "" {
		p := types.BlockchainPlatform(platformStr)
		if !isValidPlatform(p) {
			return errors.NewValidationError("invalid platform", map[string]interface{}{
				"valid_platforms": []string{string(types.PlatformFabric), string(types.PlatformBesu)},
			})
		}
		platform = &p
	}

	page := 1
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	limit := 10
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	nodes, err := h.service.ListNodes(r.Context(), platform, page, limit)
	if err != nil {
		return errors.NewInternalError("failed to list nodes", err, nil)
	}

	return response.WriteJSON(w, http.StatusOK, toPaginatedNodesResponse(nodes))
}

// ListNodesByPlatform godoc
// @Summary List nodes by platform
// @Description Get a paginated list of nodes filtered by blockchain platform
// @Tags nodes
// @Accept json
// @Produce json
// @Param platform path string true "Blockchain platform (FABRIC/BESU)" Enums(FABRIC,BESU)
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} PaginatedNodesResponse
// @Failure 400 {object} response.ErrorResponse "Validation error"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /nodes/platform/{platform} [get]
func (h *NodeHandler) ListNodesByPlatform(w http.ResponseWriter, r *http.Request) error {
	platform := types.BlockchainPlatform(chi.URLParam(r, "platform"))

	// Validate platform
	if !isValidPlatform(platform) {
		return errors.NewValidationError("invalid platform", map[string]interface{}{
			"valid_platforms": []string{string(types.PlatformFabric), string(types.PlatformBesu)},
		})
	}

	page := 1
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	limit := 10
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	nodes, err := h.service.ListNodes(r.Context(), &platform, page, limit)
	if err != nil {
		return errors.NewInternalError("failed to list nodes", err, nil)
	}

	return response.WriteJSON(w, http.StatusOK, toPaginatedNodesResponse(nodes))
}

// StartNode godoc
// @Summary Start a node
// @Description Start a node by ID
// @Tags nodes
// @Accept json
// @Produce json
// @Param id path int true "Node ID"
// @Success 200 {object} NodeResponse
// @Failure 400 {object} response.ErrorResponse "Validation error"
// @Failure 404 {object} response.ErrorResponse "Node not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /nodes/{id}/start [post]
func (h *NodeHandler) StartNode(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		return errors.NewValidationError("invalid node ID", map[string]interface{}{
			"error": err.Error(),
		})
	}

	node, err := h.service.StartNode(r.Context(), id)
	if err != nil {
		if err == service.ErrNotFound {
			return errors.NewNotFoundError("node not found", nil)
		}
		return errors.NewInternalError("failed to start node", err, nil)
	}

	return response.WriteJSON(w, http.StatusOK, toNodeResponse(node))
}

// StopNode godoc
// @Summary Stop a node
// @Description Stop a node by ID
// @Tags nodes
// @Accept json
// @Produce json
// @Param id path int true "Node ID"
// @Success 200 {object} NodeResponse
// @Failure 400 {object} response.ErrorResponse "Validation error"
// @Failure 404 {object} response.ErrorResponse "Node not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /nodes/{id}/stop [post]
func (h *NodeHandler) StopNode(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		return errors.NewValidationError("invalid node ID", map[string]interface{}{
			"error": err.Error(),
		})
	}

	node, err := h.service.StopNode(r.Context(), id)
	if err != nil {
		if err == service.ErrNotFound {
			return errors.NewNotFoundError("node not found", nil)
		}
		return errors.NewInternalError("failed to stop node", err, nil)
	}

	return response.WriteJSON(w, http.StatusOK, toNodeResponse(node))
}

// RestartNode godoc
// @Summary Restart a node
// @Description Restart a node by ID (stops and starts the node)
// @Tags nodes
// @Accept json
// @Produce json
// @Param id path int true "Node ID"
// @Success 200 {object} NodeResponse
// @Failure 400 {object} response.ErrorResponse "Validation error"
// @Failure 404 {object} response.ErrorResponse "Node not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /nodes/{id}/restart [post]
func (h *NodeHandler) RestartNode(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		return errors.NewValidationError("invalid node ID", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// First stop the node
	_, err = h.service.StopNode(r.Context(), id)
	if err != nil {
		if err == service.ErrNotFound {
			return errors.NewNotFoundError("node not found", nil)
		}
		return errors.NewInternalError("failed to stop node", err, nil)
	}

	// Then start it again
	node, err := h.service.StartNode(r.Context(), id)
	if err != nil {
		return errors.NewInternalError("failed to start node", err, nil)
	}

	return response.WriteJSON(w, http.StatusOK, toNodeResponse(node))
}

// DeleteNode godoc
// @Summary Delete a node
// @Description Delete a node by ID
// @Tags nodes
// @Accept json
// @Produce json
// @Param id path int true "Node ID"
// @Success 204 "No Content"
// @Failure 400 {object} response.ErrorResponse "Validation error"
// @Failure 404 {object} response.ErrorResponse "Node not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /nodes/{id} [delete]
func (h *NodeHandler) DeleteNode(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		return errors.NewValidationError("invalid node ID", map[string]interface{}{
			"error": err.Error(),
		})
	}

	if err := h.service.DeleteNode(r.Context(), id); err != nil {
		if err == service.ErrNotFound {
			return errors.NewNotFoundError("node not found", nil)
		}
		return errors.NewInternalError("failed to delete node", err, nil)
	}

	return response.WriteJSON(w, http.StatusNoContent, nil)
}

// GetFabricPeerDefaults godoc
// @Summary Get default values for Fabric peer node
// @Description Get default configuration values for a Fabric peer node
// @Tags nodes
// @Produce json
// @Success 200 {object} service.NodeDefaults
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /nodes/defaults/fabric-peer [get]
func (h *NodeHandler) GetFabricPeerDefaults(w http.ResponseWriter, r *http.Request) error {
	defaults := h.service.GetFabricPeerDefaults()
	return response.WriteJSON(w, http.StatusOK, defaults)
}

// GetFabricOrdererDefaults godoc
// @Summary Get default values for Fabric orderer node
// @Description Get default configuration values for a Fabric orderer node
// @Tags nodes
// @Produce json
// @Success 200 {object} service.NodeDefaults
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /nodes/defaults/fabric-orderer [get]
func (h *NodeHandler) GetFabricOrdererDefaults(w http.ResponseWriter, r *http.Request) error {
	defaults := h.service.GetFabricOrdererDefaults()
	return response.WriteJSON(w, http.StatusOK, defaults)
}

// GetNodesDefaults godoc
// @Summary Get default values for multiple Fabric nodes
// @Description Get default configuration values for multiple Fabric nodes
// @Tags nodes
// @Produce json
// @Param peerCount query int false "Number of peer nodes" default(1) minimum(0)
// @Param ordererCount query int false "Number of orderer nodes" default(1) minimum(0)
// @Param mode query string false "Deployment mode" Enums(service, docker) default(service)
// @Success 200 {object} service.NodesDefaultsResult
// @Failure 400 {object} response.ErrorResponse "Validation error"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /nodes/defaults/fabric [get]
func (h *NodeHandler) GetNodesDefaults(w http.ResponseWriter, r *http.Request) error {
	// Parse query parameters
	peerCount := 1
	if countStr := r.URL.Query().Get("peerCount"); countStr != "" {
		if count, err := strconv.Atoi(countStr); err == nil && count >= 0 {
			peerCount = count
		}
	}

	ordererCount := 1
	if countStr := r.URL.Query().Get("ordererCount"); countStr != "" {
		if count, err := strconv.Atoi(countStr); err == nil && count >= 0 {
			ordererCount = count
		}
	}

	mode := service.ModeService
	if modeStr := r.URL.Query().Get("mode"); modeStr != "" {
		mode = service.Mode(modeStr)
	}

	// Validate mode
	if mode != service.ModeService && mode != service.ModeDocker {
		return errors.NewValidationError("invalid mode", map[string]interface{}{
			"valid_modes": []string{string(service.ModeService), string(service.ModeDocker)},
		})
	}

	result, err := h.service.GetNodesDefaults(service.NodesDefaultsParams{
		PeerCount:    peerCount,
		OrdererCount: ordererCount,
		Mode:         mode,
	})
	if err != nil {
		return errors.NewInternalError("failed to get node defaults", err, nil)
	}

	return response.WriteJSON(w, http.StatusOK, result)
}

// GetBesuNodeDefaults godoc
// @Summary Get default values for Besu node
// @Description Get default configuration values for a Besu node
// @Tags nodes
// @Produce json
// @Success 200 {object} service.BesuNodeDefaults
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /nodes/defaults/besu-node [get]
func (h *NodeHandler) GetBesuNodeDefaults(w http.ResponseWriter, r *http.Request) error {
	defaults, err := h.service.GetBesuNodeDefaults()
	if err != nil {
		return errors.NewInternalError("failed to get Besu node defaults", err, nil)
	}
	return response.WriteJSON(w, http.StatusOK, defaults)
}

// TailLogs godoc
// @Summary Tail node logs
// @Description Stream logs from a specific node
// @Tags nodes
// @Accept json
// @Produce text/event-stream
// @Param id path int true "Node ID"
// @Param follow query bool false "Follow logs" default(false)
// @Param tail query int false "Number of lines to show from the end" default(100)
// @Success 200 {string} string "Log stream"
// @Failure 400 {object} response.ErrorResponse "Validation error"
// @Failure 404 {object} response.ErrorResponse "Node not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /nodes/{id}/logs [get]
func (h *NodeHandler) TailLogs(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid node ID", http.StatusBadRequest)
		return
	}

	// Parse query parameters
	follow := false
	if followStr := r.URL.Query().Get("follow"); followStr == "true" {
		follow = true
	}

	tail := 100 // default to last 100 lines
	if tailStr := r.URL.Query().Get("tail"); tailStr != "" {
		if t, err := strconv.Atoi(tailStr); err == nil && t > 0 {
			tail = t
		}
	}

	// Set headers for streaming response
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Transfer-Encoding", "chunked")

	// Create a context that's canceled when the client disconnects
	ctx := r.Context()

	// Create channel for logs
	logChan, err := h.service.TailLogs(ctx, id, tail, follow)
	if err != nil {
		if err == service.ErrNotFound {
			http.Error(w, "Node not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to tail logs: "+err.Error(), http.StatusInternalServerError)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Stream logs to client
	for {
		select {
		case <-ctx.Done():
			// Client disconnected
			return
		case logLine, ok := <-logChan:
			if !ok {
				// Channel closed
				return
			}
			// Write log line to response
			fmt.Fprintf(w, "data: %s\n\n", logLine)
			flusher.Flush()
		}
	}
}

// GetNodeEvents godoc
// @Summary Get node events
// @Description Get a paginated list of events for a specific node
// @Tags nodes
// @Accept json
// @Produce json
// @Param id path int true "Node ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} PaginatedNodeEventsResponse
// @Failure 400 {object} response.ErrorResponse "Validation error"
// @Failure 404 {object} response.ErrorResponse "Node not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /nodes/{id}/events [get]
func (h *NodeHandler) GetNodeEvents(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		return errors.NewValidationError("invalid node ID", map[string]interface{}{
			"error": err.Error(),
		})
	}

	page := 1
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	limit := 10
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	events, err := h.service.GetNodeEvents(r.Context(), id, page, limit)
	if err != nil {
		if err == service.ErrNotFound {
			return errors.NewNotFoundError("node not found", nil)
		}
		return errors.NewInternalError("failed to get node events", err, nil)
	}

	eventsResponse := PaginatedNodeEventsResponse{
		Items: make([]NodeEventResponse, len(events)),
		Page:  page,
	}

	for i, event := range events {
		eventsResponse.Items[i] = toNodeEventResponse(event)
	}

	return response.WriteJSON(w, http.StatusOK, eventsResponse)
}

// Mapping functions

func mapHTTPToServiceFabricPeerConfig(config *types.FabricPeerConfig) *types.FabricPeerConfig {
	if config == nil {
		return nil
	}
	return &types.FabricPeerConfig{
		Name:                    config.Name,
		OrganizationID:          config.OrganizationID,
		ExternalEndpoint:        config.ExternalEndpoint,
		ListenAddress:           config.ListenAddress,
		EventsAddress:           config.EventsAddress,
		OperationsListenAddress: config.OperationsListenAddress,
		ChaincodeAddress:        config.ChaincodeAddress,
		DomainNames:             config.DomainNames,
		Env:                     config.Env,
		MSPID:                   config.MSPID,
	}
}

func mapHTTPToServiceFabricOrdererConfig(config *types.FabricOrdererConfig) *types.FabricOrdererConfig {
	if config == nil {
		return nil
	}
	return &types.FabricOrdererConfig{
		Name:           config.Name,
		OrganizationID: config.OrganizationID,
		// Mode:                    "service",
		ExternalEndpoint:        config.ExternalEndpoint,
		ListenAddress:           config.ListenAddress,
		AdminAddress:            config.AdminAddress,
		OperationsListenAddress: config.OperationsListenAddress,
		DomainNames:             config.DomainNames,
		Env:                     config.Env,
		MSPID:                   config.MSPID,
	}
}

func mapHTTPToServiceBesuNodeConfig(config *types.BesuNodeConfig) *types.BesuNodeConfig {
	if config == nil {
		return nil
	}
	return &types.BesuNodeConfig{
		NetworkID: config.NetworkID,
		P2PPort:   config.P2PPort,
		RPCPort:   config.RPCPort,
		BaseNodeConfig: types.BaseNodeConfig{
			Type: "besu",
			Mode: config.Mode,
		},
		KeyID:      config.KeyID,
		P2PHost:    config.P2PHost,
		RPCHost:    config.RPCHost,
		InternalIP: config.InternalIP,
		ExternalIP: config.ExternalIP,
		Env:        config.Env,
	}
}
func mapServiceToHTTPFabricPeerDeploymentConfig(config *types.FabricPeerDeploymentConfig) *types.FabricPeerDeploymentConfig {
	if config == nil {
		return nil
	}
	return config
}

func mapServiceToHTTPNodeResponse(node *service.NodeResponse) NodeResponse {
	return NodeResponse{
		ID:                 node.ID,
		Name:               node.Name,
		BlockchainPlatform: string(node.Platform),
		NodeType:           string(node.NodeType),
		Status:             string(node.Status),
		Endpoint:           node.Endpoint,
		CreatedAt:          node.CreatedAt,
		UpdatedAt:          node.UpdatedAt,
		FabricPeer:         node.FabricPeer,
		FabricOrderer:      node.FabricOrderer,
		BesuNode:           node.BesuNode,
	}
}

func mapServiceToHTTPPaginatedResponse(nodes *service.PaginatedNodes) PaginatedNodesResponse {
	items := make([]NodeResponse, len(nodes.Items))
	for i, node := range nodes.Items {
		items[i] = mapServiceToHTTPNodeResponse(&node)
	}

	return PaginatedNodesResponse{
		Items:       items,
		Total:       nodes.Total,
		Page:        nodes.Page,
		PageCount:   nodes.PageCount,
		HasNextPage: nodes.HasNextPage,
	}
}

func toNodeResponse(node *service.NodeResponse) NodeResponse {
	return NodeResponse{
		ID:                 node.ID,
		Name:               node.Name,
		BlockchainPlatform: node.Platform,
		NodeType:           string(node.NodeType),
		Status:             string(node.Status),
		Endpoint:           node.Endpoint,
		CreatedAt:          node.CreatedAt,
		UpdatedAt:          node.UpdatedAt,
		FabricPeer:         node.FabricPeer,
		FabricOrderer:      node.FabricOrderer,
		BesuNode:           node.BesuNode,
	}
}

func isValidEventType(eventType service.NodeEventType) bool {
	switch eventType {
	case service.NodeEventStarting,
		service.NodeEventStarted,
		service.NodeEventStopping,
		service.NodeEventStopped,
		service.NodeEventError:
		return true
	}
	return false
}

// Helper function to validate platform
func isValidPlatform(platform types.BlockchainPlatform) bool {
	switch platform {
	case types.PlatformFabric, types.PlatformBesu:
		return true
	}
	return false
}

func toNodeEventResponse(event service.NodeEvent) NodeEventResponse {
	return NodeEventResponse{
		ID:        event.ID,
		NodeID:    event.NodeID,
		Type:      string(event.Type),
		Data:      event.Data,
		CreatedAt: event.CreatedAt,
	}
}
