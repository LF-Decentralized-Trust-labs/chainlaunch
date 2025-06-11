package ai

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/chainlaunch/chainlaunch/pkg/errors"
	"github.com/chainlaunch/chainlaunch/pkg/http/response"
	"github.com/chainlaunch/chainlaunch/pkg/scai/boilerplates"
	"github.com/chainlaunch/chainlaunch/pkg/scai/projects"
	"github.com/go-chi/chi/v5"
)

// Model represents an AI model
type Model struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	MaxTokens   int    `json:"maxTokens"`
}

// Template represents a project template
type Template struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// GenerateRequest represents a code generation request
type GenerateRequest struct {
	ProjectID int64  `json:"projectId"`
	Prompt    string `json:"prompt"`
}

// GenerateResponse represents a code generation response
type GenerateResponse struct {
	Code string `json:"code"`
}

// NewAIHandler creates a new instance of AIHandler with the required dependencies
func NewAIHandler(openAIService *OpenAIChatService, chatService *ChatService, projectsService *projects.ProjectsService, boilerplateService *boilerplates.BoilerplateService) *AIHandler {
	return &AIHandler{
		OpenAIChatService: openAIService,
		ChatService:       chatService,
		Projects:          projectsService,
		Boilerplates:      boilerplateService,
	}
}

// AIHandler now has a ChatService field for dependency injection.
type AIHandler struct {
	OpenAIChatService *OpenAIChatService
	ChatService       *ChatService
	Projects          *projects.ProjectsService
	Boilerplates      *boilerplates.BoilerplateService
}

// RegisterRoutes registers all AI-related routes
func (h *AIHandler) RegisterRoutes(r chi.Router) {
	r.Route("/ai", func(r chi.Router) {
		r.Get("/boilerplates", response.Middleware(h.GetBoilerplates))
		r.Get("/models", response.Middleware(h.GetModels))
		r.Post("/generate", response.Middleware(h.Generate))
		r.Get("/{projectId}/conversations", response.Middleware(h.GetConversations))
		r.Get("/{projectId}/conversations/{conversationId}", response.Middleware(h.GetConversationMessages))
		r.Get("/{projectId}/conversations/{conversationId}/export", response.Middleware(h.GetConversationDetail))
	})
}

// GetBoilerplates godoc
// @Summary      Get available boilerplates
// @Description  Returns a list of available boilerplates filtered by network platform
// @Tags         ai
// @Produce      json
// @Param        network_id query int true "Network ID to filter boilerplates by platform"
// @Success      200 {array} Boilerplate
// @Failure      400 {object} response.ErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /api/v1/ai/boilerplates [get]
func (h *AIHandler) GetBoilerplates(w http.ResponseWriter, r *http.Request) error {
	networkIDStr := r.URL.Query().Get("network_id")
	if networkIDStr == "" {
		return errors.NewValidationError("network_id is required", nil)
	}

	networkID, err := strconv.ParseInt(networkIDStr, 10, 64)
	if err != nil {
		return errors.NewValidationError("invalid network_id", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Get boilerplates for the network
	boilerplates, err := h.Boilerplates.GetBoilerplatesByNetworkID(r.Context(), networkID)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return errors.NewNotFoundError("network not found", nil)
		}
		return errors.NewInternalError("failed to get boilerplates", err, nil)
	}

	return response.WriteJSON(w, http.StatusOK, boilerplates)
}

// GetModels godoc
// @Summary      Get available AI models
// @Description  Returns a list of available AI models for code generation
// @Tags         ai
// @Produce      json
// @Success      200 {array} Model
// @Failure      500 {object} response.ErrorResponse
// @Router       /api/v1/ai/models [get]
func (h *AIHandler) GetModels(w http.ResponseWriter, r *http.Request) error {
	models := []Model{
		{
			Name:        "GPT-4",
			Description: "Most capable model, best for complex tasks",
			MaxTokens:   8192,
		},
		{
			Name:        "GPT-3.5",
			Description: "Fast and efficient model for simpler tasks",
			MaxTokens:   4096,
		},
	}
	return response.WriteJSON(w, http.StatusOK, models)
}

// Generate godoc
// @Summary      Generate code
// @Description  Generates code based on the provided prompt and project context
// @Tags         ai
// @Accept       json
// @Produce      json
// @Param        request body GenerateRequest true "Generation request"
// @Success      200 {object} GenerateResponse
// @Failure      400 {object} response.ErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /api/v1/ai/generate [post]
func (h *AIHandler) Generate(w http.ResponseWriter, r *http.Request) error {
	var req GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return errors.NewValidationError("invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Get project directly from the database
	project, err := h.Projects.Queries.GetProject(r.Context(), req.ProjectID)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return errors.NewNotFoundError("project not found", nil)
		}
		return errors.NewInternalError("failed to get project", err, nil)
	}

	code, err := h.ChatService.GenerateCode(r.Context(), req.Prompt, project)
	if err != nil {
		return errors.NewInternalError("failed to generate code", err, nil)
	}

	return response.WriteJSON(w, http.StatusOK, GenerateResponse{
		Code: code,
	})
}

// GetConversations godoc
// @Summary      Get all conversations for a project
// @Description  Returns a list of all chat conversations associated with a specific project
// @Tags         ai
// @Produce      json
// @Param        projectId path int true "Project ID"
// @Success      200 {array} ConversationResponse
// @Failure      400 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /api/v1/ai/{projectId}/conversations [get]
func (h *AIHandler) GetConversations(w http.ResponseWriter, r *http.Request) error {
	projectID, err := strconv.ParseInt(chi.URLParam(r, "projectId"), 10, 64)
	if err != nil {
		return errors.NewValidationError("invalid project ID", map[string]interface{}{
			"error": err.Error(),
		})
	}

	convs, err := h.ChatService.Queries.ListConversationsForProject(r.Context(), projectID)
	if err != nil {
		return errors.NewInternalError("failed to get conversations", err, nil)
	}

	var resp []ConversationResponse
	for _, c := range convs {
		resp = append(resp, ConversationResponse{
			ID:        c.ID,
			ProjectID: c.ProjectID,
			StartedAt: c.StartedAt.Format(time.RFC3339),
		})
	}

	return response.WriteJSON(w, http.StatusOK, resp)
}

// GetConversationMessages godoc
// @Summary      Get conversation messages
// @Description  Get all messages in a conversation
// @Tags         ai
// @Produce      json
// @Param        projectId path int true "Project ID"
// @Param        conversationId path int true "Conversation ID"
// @Success      200 {array} Message
// @Failure      400 {object} response.ErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /api/v1/ai/{projectId}/conversations/{conversationId} [get]
func (h *AIHandler) GetConversationMessages(w http.ResponseWriter, r *http.Request) error {
	projectID, err := strconv.ParseInt(chi.URLParam(r, "projectId"), 10, 64)
	if err != nil {
		return errors.NewValidationError("invalid project ID", map[string]interface{}{
			"error": err.Error(),
		})
	}

	conversationID, err := strconv.ParseInt(chi.URLParam(r, "conversationId"), 10, 64)
	if err != nil {
		return errors.NewValidationError("invalid conversation ID", map[string]interface{}{
			"error": err.Error(),
		})
	}

	messages, err := h.ChatService.GetConversationMessages(r.Context(), projectID, conversationID)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return errors.NewNotFoundError("conversation not found", nil)
		}
		return errors.NewInternalError("failed to get conversation messages", err, nil)
	}

	return response.WriteJSON(w, http.StatusOK, messages)
}

// GetConversationDetail godoc
// @Summary      Get conversation detail
// @Description  Get detailed information about a conversation including all messages and metadata
// @Tags         ai
// @Produce      json
// @Param        projectId path int true "Project ID"
// @Param        conversationId path int true "Conversation ID"
// @Success      200 {object} ConversationDetail
// @Failure      400 {object} response.ErrorResponse
// @Failure      404 {object} response.ErrorResponse
// @Failure      500 {object} response.ErrorResponse
// @Router       /api/v1/ai/{projectId}/conversations/{conversationId}/export [get]
func (h *AIHandler) GetConversationDetail(w http.ResponseWriter, r *http.Request) error {
	projectID, err := strconv.ParseInt(chi.URLParam(r, "projectId"), 10, 64)
	if err != nil {
		return errors.NewValidationError("invalid project ID", map[string]interface{}{
			"error": err.Error(),
		})
	}

	conversationID, err := strconv.ParseInt(chi.URLParam(r, "conversationId"), 10, 64)
	if err != nil {
		return errors.NewValidationError("invalid conversation ID", map[string]interface{}{
			"error": err.Error(),
		})
	}

	detail, err := h.ChatService.GetConversationDetail(r.Context(), projectID, conversationID)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return errors.NewNotFoundError("conversation not found", nil)
		}
		return errors.NewInternalError("failed to get conversation detail", err, nil)
	}

	return response.WriteJSON(w, http.StatusOK, detail)
}

// ConversationResponse represents a conversation for API responses
// swagger:model
type ConversationResponse struct {
	ID        int64  `json:"id"`
	ProjectID int64  `json:"projectId"`
	StartedAt string `json:"startedAt"`
}

// MessageResponse represents a message for API responses
// swagger:model
type MessageResponse struct {
	ID             int64  `json:"id"`
	ConversationID int64  `json:"conversationId"`
	Sender         string `json:"sender"`
	Content        string `json:"content"`
	CreatedAt      string `json:"createdAt"`
}

// MessageDetailResponse represents a message with tool calls
// swagger:model
type MessageDetailResponse struct {
	ID             int64              `json:"id"`
	ConversationID int64              `json:"conversationId"`
	Sender         string             `json:"sender"`
	Content        string             `json:"content"`
	CreatedAt      string             `json:"createdAt"`
	ToolCalls      []ToolCallResponse `json:"toolCalls"`
}

// ToolCallResponse represents a tool call for API responses
// swagger:model
type ToolCallResponse struct {
	ID        int64  `json:"id"`
	MessageID int64  `json:"messageId"`
	ToolName  string `json:"toolName"`
	Arguments string `json:"arguments"`
	Result    string `json:"result"`
	Error     string `json:"error"`
	CreatedAt string `json:"createdAt"`
}

// ParentMessageDetailResponse represents a parent message with its children (tool calls, etc.)
// swagger:model
type ParentMessageDetailResponse struct {
	Message  MessageDetailResponse   `json:"message"`
	Children []MessageDetailResponse `json:"children"`
}
