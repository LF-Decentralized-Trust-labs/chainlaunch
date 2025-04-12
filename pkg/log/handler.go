package log

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/chainlaunch/chainlaunch/pkg/nodes/service"
)

// LogHandler handles HTTP requests for log operations
type LogHandler struct {
	logService  *LogService
	nodeService *service.NodeService
}

// LogResponse represents the standard response format for log endpoints
type LogResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// NewLogHandler creates a new instance of LogHandler
func NewLogHandler(logService *LogService, nodeService *service.NodeService) *LogHandler {
	return &LogHandler{
		logService:  logService,
		nodeService: nodeService,
	}
}

// RegisterRoutes registers the log handler routes with the provided router
func (h *LogHandler) RegisterRoutes(router *http.ServeMux) {
	router.HandleFunc("/nodes/{nodeID}/logs", h.GetNodeLogs)
	router.HandleFunc("/nodes/{nodeID}/logs/range", h.GetLogRange)
	router.HandleFunc("/nodes/{nodeID}/logs/filter", h.FilterLogs)
	router.HandleFunc("/nodes/{nodeID}/logs/tail", h.TailLogs)
	router.HandleFunc("/nodes/{nodeID}/logs/stats", h.GetLogStats)
}

// GetNodeLogs handles requests to get all logs for a node
func (h *LogHandler) GetNodeLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	nodeID, err := h.getNodeIDFromPath(r.URL.Path)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid node ID")
		return
	}

	// Get node from node service
	node, err := h.nodeService.GetNode(r.Context(), nodeID)
	if err != nil {
		h.sendError(w, http.StatusNotFound, "node not found")
		return
	}

	// Get log file path based on node type and configuration
	logPath, err := h.nodeService.GetNodeLogPath(r.Context(), node)
	if err != nil {
		h.sendError(w, http.StatusNotFound, "log file not found")
		return
	}

	// Stream logs to response
	w.Header().Set("Content-Type", "text/plain")
	err = h.logService.StreamLog(logPath, FilterOptions{}, w)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "failed to stream logs")
		return
	}
}

// GetLogRange handles requests to get a specific range of log lines
func (h *LogHandler) GetLogRange(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	nodeID, err := h.getNodeIDFromPath(r.URL.Path)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid node ID")
		return
	}

	startLine, err := strconv.Atoi(r.URL.Query().Get("start"))
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid start line")
		return
	}

	endLine, err := strconv.Atoi(r.URL.Query().Get("end"))
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid end line")
		return
	}

	// Get node and log path
	node, err := h.nodeService.GetNode(r.Context(), nodeID)
	if err != nil {
		h.sendError(w, http.StatusNotFound, "node not found")
		return
	}

	logPath, err := h.nodeService.GetNodeLogPath(r.Context(), node)
	if err != nil {
		h.sendError(w, http.StatusNotFound, "log file not found")
		return
	}

	// Get log range
	logRange, err := h.logService.ReadLogRange(logPath, startLine, endLine)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "failed to read log range")
		return
	}

	h.sendJSON(w, http.StatusOK, logRange)
}

// FilterLogs handles requests to filter logs based on pattern and range
func (h *LogHandler) FilterLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	nodeID, err := h.getNodeIDFromPath(r.URL.Path)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid node ID")
		return
	}

	// Parse filter options from query parameters
	options := FilterOptions{
		Pattern:    r.URL.Query().Get("pattern"),
		IgnoreCase: r.URL.Query().Get("ignoreCase") == "true",
	}

	if startStr := r.URL.Query().Get("start"); startStr != "" {
		start, err := strconv.Atoi(startStr)
		if err != nil {
			h.sendError(w, http.StatusBadRequest, "invalid start line")
			return
		}
		options.StartLine = start
	}

	if endStr := r.URL.Query().Get("end"); endStr != "" {
		end, err := strconv.Atoi(endStr)
		if err != nil {
			h.sendError(w, http.StatusBadRequest, "invalid end line")
			return
		}
		options.EndLine = end
	}

	// Get node and log path
	node, err := h.nodeService.GetNode(r.Context(), nodeID)
	if err != nil {
		h.sendError(w, http.StatusNotFound, "node not found")
		return
	}

	logPath, err := h.nodeService.GetNodeLogPath(r.Context(), node)
	if err != nil {
		h.sendError(w, http.StatusNotFound, "log file not found")
		return
	}

	// Filter logs
	entries, err := h.logService.FilterLog(logPath, options)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "failed to filter logs")
		return
	}

	h.sendJSON(w, http.StatusOK, entries)
}

// TailLogs handles requests to get the last n lines of logs
func (h *LogHandler) TailLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	nodeID, err := h.getNodeIDFromPath(r.URL.Path)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid node ID")
		return
	}

	lines, err := strconv.Atoi(r.URL.Query().Get("lines"))
	if err != nil || lines < 1 {
		h.sendError(w, http.StatusBadRequest, "invalid number of lines")
		return
	}

	// Get node and log path
	node, err := h.nodeService.GetNode(r.Context(), nodeID)
	if err != nil {
		h.sendError(w, http.StatusNotFound, "node not found")
		return
	}

	logPath, err := h.nodeService.GetNodeLogPath(r.Context(), node)
	if err != nil {
		h.sendError(w, http.StatusNotFound, "log file not found")
		return
	}

	// Get tail of log
	entries, err := h.logService.TailLog(logPath, lines)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "failed to tail logs")
		return
	}

	h.sendJSON(w, http.StatusOK, entries)
}

// GetLogStats handles requests to get statistics about a log file
func (h *LogHandler) GetLogStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	nodeID, err := h.getNodeIDFromPath(r.URL.Path)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid node ID")
		return
	}

	// Get node and log path
	node, err := h.nodeService.GetNode(r.Context(), nodeID)
	if err != nil {
		h.sendError(w, http.StatusNotFound, "node not found")
		return
	}

	logPath, err := h.nodeService.GetNodeLogPath(r.Context(), node)
	if err != nil {
		h.sendError(w, http.StatusNotFound, "log file not found")
		return
	}

	// Get log stats
	stats, err := h.logService.GetLogStats(logPath)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "failed to get log stats")
		return
	}

	h.sendJSON(w, http.StatusOK, stats)
}

// getNodeIDFromPath extracts the node ID from the URL path
func (h *LogHandler) getNodeIDFromPath(path string) (int64, error) {
	// Extract nodeID from path like "/nodes/{nodeID}/logs"
	var nodeID int64
	_, err := fmt.Sscanf(path, "/nodes/%d/logs", &nodeID)
	if err != nil {
		return 0, err
	}
	return nodeID, nil
}

// sendJSON sends a JSON response
func (h *LogHandler) sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(LogResponse{
		Success: true,
		Data:    data,
	})
}

// sendError sends an error response
func (h *LogHandler) sendError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(LogResponse{
		Success: false,
		Error:   message,
	})
}
