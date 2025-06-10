package ai

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/chainlaunch/chainlaunch/pkg/db"
	"github.com/chainlaunch/chainlaunch/pkg/scai/projects"
	"github.com/chainlaunch/chainlaunch/pkg/scai/sessionchanges"
	"github.com/chainlaunch/chainlaunch/pkg/scai/versionmanagement"
	"github.com/go-chi/chi/v5"
)

// Message represents a single chat message.
type Message struct {
	Role    string `json:"role"` // "user" or "assistant"
	Content string `json:"content"`
}

// AIHandler now has a ChatService field for dependency injection.
type AIHandler struct {
	OpenAIChatService *OpenAIChatService
	ChatService       *ChatService

	Projects *projects.ProjectsService // Use the correct type from the projects package
}

// RegisterRoutes registers AI API Gateway endpoints to the router
func (h *AIHandler) RegisterRoutes(r chi.Router) {
	r.Get("/api/ai/templates", h.GetTemplates)
	r.Get("/api/ai/boilerplates", h.GetBoilerplates)
	r.Route("/api/ai/{projectId}", func(r chi.Router) {
		r.Get("/models", h.GetModels)
		r.Post("/generate", h.Generate)
		r.Post("/chat", h.Chat)
		r.Get("/docs", h.GetDocs)
		r.Get("/conversations", h.GetConversations)
		r.Get("/conversations/{conversationId}", h.GetConversationMessages)
		r.Get("/conversations/{conversationId}/export", h.GetConversationDetail)
	})
}

// GetModels godoc
// @Summary      List available AI models
// @Description  Get a list of available AI models
// @Tags         ai
// @Produce      json
// @Success      200 {object} map[string]interface{}
// @Router       /api/ai/models [get]
func (h *AIHandler) GetModels(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("Not implemented: /api/ai/models"))
}

// Generate godoc
// @Summary      Generate code or content using AI
// @Description  Generate code or content using the selected AI model
// @Tags         ai
// @Accept       json
// @Produce      json
// @Success      200 {object} map[string]interface{}
// @Router       /api/ai/generate [post]
func (h *AIHandler) Generate(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("Not implemented: /api/ai/generate"))
}

// GetTemplates godoc
// @Summary      List AI instruction templates
// @Description  Get a list of available AI instruction templates
// @Tags         ai
// @Produce      json
// @Success      200 {object} map[string]interface{}
// @Router       /api/ai/templates [get]
func (h *AIHandler) GetTemplates(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("Not implemented: /api/ai/templates"))
}

// GetBoilerplates godoc
// @Summary      List available boilerplates
// @Description  Get a list of available boilerplate templates for project initialization
// @Tags         ai
// @Produce      json
// @Success      200 {object} map[string]interface{}
// @Router       /api/ai/boilerplates [get]
func (h *AIHandler) GetBoilerplates(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("Not implemented: /api/ai/boilerplates"))
}

type ChatMessagePart struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type ChatMessage struct {
	Role    string            `json:"role"`
	Content string            `json:"content"`
	Parts   []ChatMessagePart `json:"parts"`
}

type ChatRequest struct {
	ID        string        `json:"id"`
	Messages  []ChatMessage `json:"messages"`
	ProjectID string        `json:"projectId"`
}

// ChatResponse is not used for streaming, but kept for reference.
type ChatResponse struct {
	Reply string `json:"reply"`
}

// EventType is a string type for SSE event types
// (stringer is optional, but helps with enums)
//
//go:generate stringer -type=EventType
type EventType string

const (
	EventTypeLLM             EventType = "llm"
	EventTypeToolStart       EventType = "tool_start"
	EventTypeToolUpdate      EventType = "tool_update"
	EventTypeToolExecute     EventType = "tool_execute"
	EventTypeToolResult      EventType = "tool_result"
	EventTypeMaxStepsReached EventType = "max_steps_reached"
)

// Update all event structs to use EventType for the Type field

type TokenEvent struct {
	Type  EventType `json:"type"`
	Token string    `json:"token"`
}

type LLMEvent struct {
	Type    EventType `json:"type"`
	Content string    `json:"content"`
}

type ToolStartEvent struct {
	Type       EventType `json:"type"`
	ToolCallID string    `json:"toolCallID"`
	Name       string    `json:"name"`
}

type ToolUpdateEvent struct {
	Type       EventType `json:"type"`
	ToolCallID string    `json:"toolCallID"`
	Name       string    `json:"name"`
	Arguments  string    `json:"arguments"`
}

type ToolExecuteEvent struct {
	Type       EventType              `json:"type"`
	ToolCallID string                 `json:"toolCallID"`
	Name       string                 `json:"name"`
	Args       map[string]interface{} `json:"args"`
}

type ToolResultEvent struct {
	Type       EventType   `json:"type"`
	ToolCallID string      `json:"toolCallID"`
	Name       string      `json:"name"`
	Result     interface{} `json:"result"`
	Error      string      `json:"error,omitempty"`
}

type MaxStepsReachedEvent struct {
	Type EventType `json:"type"`
}

// sseAgentStepObserver streams agent step events as SSE to the client
// Needs access to http.ResponseWriter and http.Flusher
// We'll store them as fields in the struct
type sseAgentStepObserver struct {
	w       http.ResponseWriter
	flusher http.Flusher
}

func (o *sseAgentStepObserver) OnLLMContent(content string) {
	if content == "" {
		return
	}
	evt := LLMEvent{Type: EventTypeLLM, Content: content}
	data, _ := json.Marshal(evt)
	fmt.Fprintf(o.w, "data: %s\n\n", data)
	o.flusher.Flush()
}

func (o *sseAgentStepObserver) OnToolCallStart(toolCallID, name string) {
	evt := ToolStartEvent{Type: EventTypeToolStart, ToolCallID: toolCallID, Name: name}
	data, _ := json.Marshal(evt)
	fmt.Fprintf(o.w, "data: %s\n\n", data)
	o.flusher.Flush()
}

func (o *sseAgentStepObserver) OnToolCallUpdate(toolCallID, name, arguments string) {
	evt := ToolUpdateEvent{Type: EventTypeToolUpdate, ToolCallID: toolCallID, Name: name, Arguments: arguments}
	data, _ := json.Marshal(evt)
	fmt.Fprintf(o.w, "data: %s\n\n", data)
	o.flusher.Flush()
}

func (o *sseAgentStepObserver) OnToolCallExecute(toolCallID, name string, args map[string]interface{}) {
	evt := ToolExecuteEvent{Type: EventTypeToolExecute, ToolCallID: toolCallID, Name: name, Args: args}
	data, _ := json.Marshal(evt)
	fmt.Fprintf(o.w, "data: %s\n\n", data)
	o.flusher.Flush()
}

func (o *sseAgentStepObserver) OnToolCallResult(toolCallID, name string, result interface{}, err error) {
	evt := ToolResultEvent{Type: EventTypeToolResult, ToolCallID: toolCallID, Name: name, Result: result}
	if err != nil {
		evt.Error = err.Error()
	}
	data, _ := json.Marshal(evt)
	fmt.Fprintf(o.w, "data: %s\n\n", data)
	o.flusher.Flush()
}

func (o *sseAgentStepObserver) OnMaxStepsReached() {
	evt := MaxStepsReachedEvent{Type: EventTypeMaxStepsReached}
	data, _ := json.Marshal(evt)
	fmt.Fprintf(o.w, "data: %s\n\n", data)
	o.flusher.Flush()
}

// Chat godoc
// @Summary      Chat with AI assistant
// @Description  Stream a conversation with the AI assistant using Server-Sent Events (SSE)
// @Tags         ai
// @Accept       json
// @Produce      text/event-stream
// @Param        projectId path int true "Project ID"
// @Param        request body ChatRequest true "Chat request containing project ID and messages"
// @Success      200 {string} string "SSE stream of chat responses"
// @Failure      400 {string} string "Invalid request"
// @Failure      500 {string} string "Internal server error"
// @Router       /api/ai/{projectId}/chat [post]
func (h *AIHandler) Chat(w http.ResponseWriter, r *http.Request) {
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	projectIdStr := chi.URLParam(r, "projectId")
	if projectIdStr == "" {
		http.Error(w, "projectId is required", http.StatusBadRequest)
		return
	}

	// Use projectId from path or from body (prefer path param)
	projectID, err := strconv.ParseInt(projectIdStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid projectId", http.StatusBadRequest)
		return
	}

	if len(req.Messages) == 0 {
		http.Error(w, "messages are required", http.StatusBadRequest)
		return
	}

	// Use the last user message as the prompt
	var userMessage string
	for i := len(req.Messages) - 1; i >= 0; i-- {
		if req.Messages[i].Role == "user" {
			userMessage = req.Messages[i].Content
			break
		}
	}
	if userMessage == "" {
		http.Error(w, "no user message found", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	observer := &sseAgentStepObserver{w: w, flusher: flusher}

	err = h.OpenAIChatService.ChatWithPersistence(r.Context(), projectID, userMessage, observer, 0)
	if err != nil && err != io.EOF {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// After chat session, commit all changed files in the correct project directory
	files := sessionchanges.GetAndResetChanges()
	if len(files) > 0 {
		msg := "AI Chat Session: Modified files:\n- " + strings.Join(files, "\n- ")
		vm := versionmanagement.NewDefaultManager()
		ctx := r.Context()
		// Get project directory from projectID
		proj, err := h.Projects.GetProject(ctx, projectID)
		if err == nil {
			projectDir := filepath.Join(h.Projects.ProjectsDir, proj.Name)
			cwd, _ := os.Getwd()
			os.Chdir(projectDir)
			if err := vm.CommitChange(ctx, msg); err != nil {
				fmt.Printf("Failed to commit session changes: %v\n", err)
			}
			os.Chdir(cwd)
		}
	}
}

// GetDocs godoc
// @Summary      Get AI API documentation
// @Description  Get documentation for the AI API endpoints
// @Tags         ai
// @Produce      json
// @Success      200 {object} map[string]interface{}
// @Router       /api/ai/docs [get]
func (h *AIHandler) GetDocs(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("Not implemented: /api/ai/docs"))
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

// GetConversations godoc
// @Summary      Get all conversations for a project
// @Description  Returns a list of all chat conversations associated with a specific project
// @Tags         ai
// @Produce      json
// @Param        projectId path int true "Project ID"
// @Success      200 {array} ConversationResponse
// @Failure      400 {string} string "Invalid request"
// @Failure      500 {string} string "Internal server error"
// @Router       /api/ai/{projectId}/conversations [get]
func (h *AIHandler) GetConversations(w http.ResponseWriter, r *http.Request) {
	projectIdStr := chi.URLParam(r, "projectId")
	if projectIdStr == "" {
		http.Error(w, "projectId is required", http.StatusBadRequest)
		return
	}
	projectID, err := strconv.ParseInt(projectIdStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid projectId", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	convs, err := h.ChatService.Queries.ListConversationsForProject(ctx, projectID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var resp []ConversationResponse
	for _, c := range convs {
		resp = append(resp, ConversationResponse{
			ID:        c.ID,
			ProjectID: c.ProjectID,
			StartedAt: c.StartedAt.Format(time.RFC3339),
		})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
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

// Update GetConversationMessages to group messages using parent_id from the DB, not just consecutive grouping. For each message with parent_id == null, treat as a parent; for each with parent_id set, group as a child under the parent message. The response should be a list of ParentMessageDetailResponse, each with its main message and any children. Update the grouping logic accordingly.
// @Summary      Get all messages in a conversation
// @Description  Returns a list of all messages in a specific chat conversation, aggregating tool messages under their parent
// @Tags         ai
// @Produce      json
// @Param        conversationId path int true "Conversation ID"
// @Success      200 {array} ParentMessageDetailResponse
// @Failure      400 {string} string "Invalid request"
// @Failure      500 {string} string "Internal server error"
// @Router       /api/ai/{projectId}/conversations/{conversationId} [get]
func (h *AIHandler) GetConversationMessages(w http.ResponseWriter, r *http.Request) {
	conversationIdStr := chi.URLParam(r, "conversationId")
	if conversationIdStr == "" {
		http.Error(w, "conversationId is required", http.StatusBadRequest)
		return
	}
	conversationID, err := strconv.ParseInt(conversationIdStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid conversationId", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	msgs, err := h.ChatService.Queries.ListMessagesForConversation(ctx, conversationID)
	if err != nil {
		http.Error(w, "failed to get messages", http.StatusInternalServerError)
		return
	}
	// Fetch tool calls for all messages
	toolCallsByMsg := make(map[int64][]ToolCallResponse)
	for _, msg := range msgs {
		toolCalls, _ := h.ChatService.Queries.ListToolCallsForMessage(ctx, msg.ID)
		for _, tc := range toolCalls {
			toolCallsByMsg[msg.ID] = append(toolCallsByMsg[msg.ID], ToolCallResponse{
				ID:        tc.ID,
				MessageID: tc.MessageID,
				ToolName:  tc.ToolName,
				Arguments: tc.Arguments,
				Result:    tc.Result.String,
				Error:     tc.Error.String,
				CreatedAt: tc.CreatedAt.Format(time.RFC3339),
			})
		}
	}
	// Map message ID to MessageDetailResponse
	msgMap := make(map[int64]MessageDetailResponse)
	for _, msg := range msgs {
		msgMap[msg.ID] = MessageDetailResponse{
			ID:             msg.ID,
			ConversationID: msg.ConversationID,
			Sender:         msg.Sender,
			Content:        msg.Content,
			CreatedAt:      msg.CreatedAt.Format(time.RFC3339),
			ToolCalls:      toolCallsByMsg[msg.ID],
		}
	}
	// Group messages by parent_id
	parentMap := make(map[int64][]MessageDetailResponse)
	var parentOrder []int64
	for _, msg := range msgs {
		if !msg.ParentID.Valid {
			parentOrder = append(parentOrder, msg.ID)
			continue
		}
		parentMap[msg.ParentID.Int64] = append(parentMap[msg.ParentID.Int64], msgMap[msg.ID])
	}
	// Build response
	var result []ParentMessageDetailResponse
	for _, pid := range parentOrder {
		parent := msgMap[pid]
		children := parentMap[pid]
		result = append(result, ParentMessageDetailResponse{
			Message:  parent,
			Children: children,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// ConversationDetailResponse represents the full conversation export
// swagger:model
type ConversationDetailResponse struct {
	ID        int64                   `json:"id"`
	ProjectID int64                   `json:"projectId"`
	StartedAt string                  `json:"startedAt"`
	Messages  []MessageDetailResponse `json:"messages"`
}

// GetConversationDetail godoc
// @Summary      Export full conversation detail
// @Description  Returns all messages and tool calls for a conversation
// @Tags         ai
// @Produce      json
// @Param        conversationId path int true "Conversation ID"
// @Success      200 {object} ConversationDetailResponse
// @Failure      400 {string} string "Invalid request"
// @Failure      500 {string} string "Internal server error"
// @Router       /api/ai/{projectId}/conversations/{conversationId}/export [get]
func (h *AIHandler) GetConversationDetail(w http.ResponseWriter, r *http.Request) {
	conversationIdStr := chi.URLParam(r, "conversationId")
	if conversationIdStr == "" {
		http.Error(w, "conversationId is required", http.StatusBadRequest)
		return
	}
	conversationID, err := strconv.ParseInt(conversationIdStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid conversationId", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	// Only for this handler, use the concrete type
	// Get conversation info
	convs, err := h.ChatService.Queries.ListConversationsForProject(ctx, 0) // We'll get all, then filter
	var conv *db.Conversation
	for _, c := range convs {
		if c.ID == conversationID {
			conv = c
			break
		}
	}
	if conv.ID == 0 {
		http.Error(w, "conversation not found", http.StatusNotFound)
		return
	}
	msgs, err := h.ChatService.GetMessages(ctx, conversationID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var msgDetails []MessageDetailResponse
	for _, m := range msgs {
		toolCalls, err := h.ChatService.Queries.ListToolCallsForMessage(ctx, m.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		var toolCallDetails []ToolCallResponse
		for _, tc := range toolCalls {
			toolCallDetails = append(toolCallDetails, ToolCallResponse{
				ID:        tc.ID,
				MessageID: tc.MessageID,
				ToolName:  tc.ToolName,
				Arguments: tc.Arguments,
				Result:    tc.Result.String,
				Error:     tc.Error.String,
				CreatedAt: tc.CreatedAt.Format(time.RFC3339),
			})
		}
		msgDetails = append(msgDetails, MessageDetailResponse{
			ID:             m.ID,
			ConversationID: m.ConversationID,
			Sender:         m.Sender,
			Content:        m.Content,
			CreatedAt:      m.CreatedAt.Format(time.RFC3339),
			ToolCalls:      toolCallDetails,
		})
	}
	resp := ConversationDetailResponse{
		ID:        conv.ID,
		ProjectID: conv.ProjectID,
		StartedAt: conv.StartedAt.Format(time.RFC3339),
		Messages:  msgDetails,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
