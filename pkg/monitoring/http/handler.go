package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/chainlaunch/chainlaunch/pkg/monitoring"
	"github.com/go-chi/chi/v5"
)

// NodeStatusResponse represents the JSON response for a node status request
type NodeStatusResponse struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	URL          string    `json:"url"`
	Status       string    `json:"status"`
	LastChecked  time.Time `json:"last_checked"`
	ResponseTime string    `json:"response_time,omitempty"`
	Error        string    `json:"error,omitempty"`
	FailureCount int       `json:"failure_count,omitempty"`
	Since        time.Time `json:"status_since,omitempty"`
}

// Handler handles HTTP requests for the monitoring service
type Handler struct {
	service monitoring.Service
}

// NewHandler creates a new monitoring HTTP handler
func NewHandler(service monitoring.Service) *Handler {
	return &Handler{
		service: service,
	}
}

// RegisterRoutes registers the monitoring routes with the provided router
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/monitoring", func(r chi.Router) {
		r.Get("/nodes", h.GetAllNodeStatuses)
		r.Get("/nodes/{nodeID}", h.GetNodeStatus)
		r.Post("/nodes", h.AddNode)
		r.Delete("/nodes/{nodeID}", h.RemoveNode)
	})
}

// GetAllNodeStatuses returns the status of all monitored nodes
func (h *Handler) GetAllNodeStatuses(w http.ResponseWriter, r *http.Request) {
	statuses := h.service.GetAllNodeStatuses()

	// Convert to response format
	response := make([]NodeStatusResponse, 0, len(statuses))
	for _, status := range statuses {
		var errStr string
		if status.Error != nil {
			errStr = status.Error.Error()
		}

		respTime := status.ResponseTime.String()

		resp := NodeStatusResponse{
			ID:           status.Node.ID,
			Name:         status.Node.Name,
			URL:          status.Node.URL,
			Status:       string(status.Status),
			LastChecked:  status.Timestamp,
			ResponseTime: respTime,
			Error:        errStr,
			FailureCount: status.Node.FailureCount,
			Since:        status.Node.LastStatusChange,
		}
		response = append(response, resp)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// GetNodeStatus returns the status of a specific node
func (h *Handler) GetNodeStatus(w http.ResponseWriter, r *http.Request) {
	nodeIDStr := chi.URLParam(r, "nodeID")
	nodeID, err := strconv.ParseInt(nodeIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid node ID", http.StatusBadRequest)
		return
	}

	status, err := h.service.GetNodeStatus(nodeID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	var errStr string
	if status.Error != nil {
		errStr = status.Error.Error()
	}

	respTime := status.ResponseTime.String()

	response := NodeStatusResponse{
		ID:           status.Node.ID,
		Name:         status.Node.Name,
		URL:          status.Node.URL,
		Status:       string(status.Status),
		LastChecked:  status.Timestamp,
		ResponseTime: respTime,
		Error:        errStr,
		FailureCount: status.Node.FailureCount,
		Since:        status.Node.LastStatusChange,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// AddNodeRequest represents the request to add a node to monitoring
type AddNodeRequest struct {
	ID               int64         `json:"id"`
	Name             string        `json:"name"`
	URL              string        `json:"url"`
	CheckInterval    time.Duration `json:"check_interval_ms"`
	Timeout          time.Duration `json:"timeout_ms"`
	FailureThreshold int           `json:"failure_threshold"`
}

// AddNode adds a new node to be monitored
func (h *Handler) AddNode(w http.ResponseWriter, r *http.Request) {
	var req AddNodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.ID == 0 || req.URL == "" {
		http.Error(w, "ID and URL are required", http.StatusBadRequest)
		return
	}

	node := &monitoring.Node{
		ID:               req.ID,
		Name:             req.Name,
		URL:              req.URL,
		CheckInterval:    req.CheckInterval,
		Timeout:          req.Timeout,
		FailureThreshold: req.FailureThreshold,
	}

	if err := h.service.AddNode(node); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// RemoveNode removes a node from monitoring
func (h *Handler) RemoveNode(w http.ResponseWriter, r *http.Request) {
	nodeIDStr := chi.URLParam(r, "nodeID")
	nodeID, err := strconv.ParseInt(nodeIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid node ID", http.StatusBadRequest)
		return
	}

	if err := h.service.RemoveNode(nodeID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
