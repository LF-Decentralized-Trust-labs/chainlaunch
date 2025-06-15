package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/chainlaunch/chainlaunch/pkg/db"
	"github.com/chainlaunch/chainlaunch/pkg/logger"
	"github.com/chainlaunch/chainlaunch/pkg/scai/sessionchanges"
	"github.com/sashabaranov/go-openai"
)

// ToolSchema defines a tool with its JSON schema and handler.
type ToolSchema struct {
	Name        string
	Description string
	Parameters  map[string]interface{} // JSON schema
	Handler     func(projectRoot string, args map[string]interface{}) (interface{}, error)
}

// GetDefaultToolSchemas returns all registered tools with their schemas and handlers, scoped to a project root.
func GetDefaultToolSchemas(projectRoot string) []ToolSchema {
	return []ToolSchema{
		{
			Name:        "read_file",
			Description: "Read the contents of a file.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{"type": "string", "description": "Path to the file (relative to project root)"},
				},
				"required": []string{"path"},
			},
			Handler: func(funcName string, args map[string]interface{}) (interface{}, error) {
				path, _ := args["path"].(string)
				absPath := filepath.Join(projectRoot, path)
				data, err := os.ReadFile(absPath)
				if err != nil {
					return nil, err
				}
				return map[string]interface{}{"content": string(data)}, nil
			},
		},
		{
			Name:        "write_file",
			Description: "Write content to a file.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path":    map[string]interface{}{"type": "string", "description": "Path to the file (relative to project root)"},
					"content": map[string]interface{}{"type": "string", "description": "Content to write"},
				},
				"required": []string{"path", "content"},
			},
			Handler: func(funcName string, args map[string]interface{}) (interface{}, error) {
				path, _ := args["path"].(string)
				content, _ := args["content"].(string)
				absPath := filepath.Join(projectRoot, path)
				if err := os.WriteFile(absPath, []byte(content), 0644); err != nil {
					return nil, err
				}
				// Register the change with the global tracker for backward compatibility
				sessionchanges.RegisterChange(absPath)
				return map[string]interface{}{"result": "file written successfully"}, nil
			},
		},
	}
}

// getToolSchemas returns all registered tools with their schemas and handlers.
func getToolSchemas(projectRoot string) []ToolSchema {
	return GetDefaultToolSchemas(projectRoot)
}

// OpenAIChatService implements ChatServiceInterface using OpenAI's API and function-calling tools.
type OpenAIChatService struct {
	Client      *openai.Client
	Logger      *logger.Logger
	ChatService *ChatService
	Queries     *db.Queries
	ProjectsDir string
}

func NewOpenAIChatService(apiKey string, logger *logger.Logger, chatService *ChatService, queries *db.Queries, projectsDir string) *OpenAIChatService {
	return &OpenAIChatService{
		Client:      openai.NewClient(apiKey),
		Logger:      logger,
		ChatService: chatService,
		Queries:     queries,
		ProjectsDir: projectsDir,
	}
}

// getProjectStructurePrompt generates a system prompt with the project structure and file contents.
func getProjectStructurePrompt(projectRoot string) string {
	ignored := map[string]bool{
		"node_modules": true,
		".git":         true,
		".DS_Store":    true,
	}
	var sb strings.Builder
	sb.WriteString(`
You are an expert AI coding agent.
All projects use Bun (TypeScript) as the runtime and build system.
Here is the current project structure and contents.

Be proactive, read and write files as needed, your goal is to progress in the project and write the code to achieve the goal. Including fixing issues.
`)
	filepath.Walk(projectRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		rel, _ := filepath.Rel(projectRoot, path)
		parts := strings.Split(rel, string(os.PathSeparator))
		for _, part := range parts {
			if ignored[part] {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}
		if info.IsDir() {
			return nil
		}
		// Only include files < 32KB
		if info.Size() < 32*1024 {
			data, err := os.ReadFile(path)
			if err == nil {
				sb.WriteString("\n---\nFile: " + rel + "\n" + string(data) + "\n---\n")
			}
		} else {
			sb.WriteString("\n---\nFile: " + rel + " (too large to display)\n---\n")
		}
		return nil
	})
	return sb.String()
}

const maxAgentSteps = 10

// handleToolCall executes a tool call and returns the result as a string.
func (s *OpenAIChatService) handleToolCall(toolCall openai.ToolCall, projectRoot string) string {
	toolSchemas := getToolSchemas(projectRoot)
	var tool ToolSchema
	ok := false
	for _, t := range toolSchemas {
		if t.Name == toolCall.Function.Name {
			tool = t
			ok = true
			break
		}
	}
	if !ok {
		return `{"error": "Unknown tool function: ` + toolCall.Function.Name + `"}`
	}
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
		return `{"error": "Failed to parse arguments: ` + err.Error() + `"}`
	}
	result, err := tool.Handler(projectRoot, args)
	if err != nil {
		return `{"error": "Tool error: ` + err.Error() + `"}`
	}
	resultJson, _ := json.Marshal(result)
	return string(resultJson)
}

// StreamChat uses a multi-step tool execution loop with OpenAI function-calling.
func (s *OpenAIChatService) StreamChat(
	ctx context.Context,
	project *db.ChaincodeProject,
	conversationID int64,
	messages []Message,
	observer AgentStepObserver,
	maxSteps int,
	sessionTracker *sessionchanges.Tracker,
) error {
	var chatMsgs []openai.ChatCompletionMessage
	projectID := project.ID
	projectSlug := project.Slug
	projectRoot := filepath.Join(s.ProjectsDir, projectSlug)
	systemPrompt := getProjectStructurePrompt(projectRoot)
	s.Logger.Debugf("[StreamChat] projectID: %s", projectID)
	s.Logger.Debugf("[StreamChat] projectRoot: %s", projectRoot)
	s.Logger.Debugf("[StreamChat] systemPrompt: %s", systemPrompt)
	chatMsgs = append(chatMsgs, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: systemPrompt,
	})
	var lastParentMsgID *int64
	for _, m := range messages {
		role := openai.ChatMessageRoleUser
		if m.Sender == "assistant" {
			role = openai.ChatMessageRoleAssistant
		}
		chatMsgs = append(chatMsgs, openai.ChatCompletionMessage{
			Role:    role,
			Content: m.Content,
		})
	}

	// Update the tool schemas to use the session tracker
	toolSchemas := getToolSchemas(projectRoot)
	for i := range toolSchemas {
		originalHandler := toolSchemas[i].Handler
		toolSchemas[i].Handler = func(name string, args map[string]interface{}) (interface{}, error) {
			result, err := originalHandler(name, args)
			if err == nil && sessionTracker != nil {
				// If the tool call was successful and we have a session tracker,
				// register any file changes
				if filePath, ok := args["path"].(string); ok {
					absPath := filepath.Join(projectRoot, filePath)
					sessionTracker.RegisterChange(absPath)
				}
			}
			return result, err
		}
	}

	toolSchemasMap := make(map[string]ToolSchema)
	for _, tool := range toolSchemas {
		toolSchemasMap[tool.Name] = tool
	}
	tools := []openai.Tool{}
	for _, tool := range toolSchemas {
		tools = append(tools, openai.Tool{
			Type: "function",
			Function: &openai.FunctionDefinition{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters:  tool.Parameters,
			},
		})
	}

	if maxSteps <= 0 {
		maxSteps = maxAgentSteps
	}

	for step := 0; step < maxSteps; step++ {
		s.Logger.Debugf("[StreamChat] Agent step: %d", step)
		msg, err := StreamAgentStep(
			ctx,
			s.Client,
			chatMsgs,
			"gpt-4o",
			tools,
			toolSchemasMap,
			observer,
		)
		if err != nil {
			s.Logger.Debugf("[StreamChat] Error in StreamAgentStep: %v", err)
			return err
		}

		s.Logger.Debugf("[StreamChat] Agent step: %d, assistant message: %s", step, msg.Content)
		if len(msg.ToolCalls) > 0 {
			s.Logger.Debugf("[StreamChat] Tool calls in step: %d, %v", step, msg.ToolCalls)
		}

		chatMsgs = append(chatMsgs, msg)

		// If no tool calls, we're done
		if len(msg.ToolCalls) == 0 {
			s.Logger.Debugf("[StreamChat] No tool calls in step: %d - finishing", step)
			return nil
		}

		// Process all tool calls in this step
		for _, toolCall := range msg.ToolCalls {
			s.Logger.Debugf("[StreamChat] Handling tool call: %s, args: %s", toolCall.Function.Name, toolCall.Function.Arguments)
			resultObj, _ := s.executeAndSerializeToolCall(toolCall, projectRoot)
			resultStr := resultObj.resultStr
			errStr := resultObj.errStr
			argsStr := resultObj.argsStr
			s.Logger.Debugf("[StreamChat] Tool result for: %s, %v", toolCall.Function.Name, resultStr)

			// Add tool result message to DB and get its ID, set parentID to lastParentMsgID
			toolMsg, err := s.ChatService.AddMessage(ctx, conversationID, lastParentMsgID, "tool", resultStr)
			if err != nil {
				s.Logger.Debugf("[StreamChat] Failed to persist tool message: %v", err)
				continue
			}
			// Persist tool call
			_, err = s.ChatService.AddToolCall(ctx, toolMsg.ID, toolCall.Function.Name, argsStr, resultStr, errStr)
			if err != nil {
				s.Logger.Debugf("[StreamChat] Failed to persist tool call: %v", err)
			}
			// Add tool result message to chatMsgs for next step
			chatMsgs = append(chatMsgs, openai.ChatCompletionMessage{
				Role:       openai.ChatMessageRoleTool,
				Content:    resultStr,
				ToolCallID: toolCall.ID,
			})
		}
	}

	// If we reach max steps, notify observer and make one final call and stream the response
	if observer != nil {
		observer.OnMaxStepsReached()
	}
	s.Logger.Debugf("[StreamChat] Reached maxSteps, making final call")
	msg, err := StreamAgentStep(
		ctx,
		s.Client,
		chatMsgs,
		"gpt-4o",
		tools,
		toolSchemasMap,
		observer,
	)
	if err != nil {
		s.Logger.Debugf("[StreamChat] Error in final StreamAgentStep: %v", err)
		return err
	}
	chatMsgs = append(chatMsgs, msg)
	s.Logger.Debugf("[StreamChat] Final assistant message: %s", msg.Content)
	if len(msg.ToolCalls) > 0 {
		s.Logger.Debugf("[StreamChat] Final tool calls: %v", msg.ToolCalls)
	}

	return nil
}

// Helper to execute a tool call and serialize args/result/error
func (s *OpenAIChatService) executeAndSerializeToolCall(toolCall openai.ToolCall, projectRoot string) (struct {
	resultStr, argsStr string
	errStr             *string
}, error) {
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
		errMsg := err.Error()
		return struct {
			resultStr, argsStr string
			errStr             *string
		}{"", toolCall.Function.Arguments, &errMsg}, err
	}
	result, err := getToolSchemas(projectRoot)[0].Handler(projectRoot, args) // Find the correct handler
	var resultStr string
	if result != nil {
		b, _ := json.Marshal(result)
		resultStr = string(b)
	}
	var errStr *string
	if err != nil {
		errMsg := err.Error()
		errStr = &errMsg
	}
	argsStr, _ := json.Marshal(args)
	return struct {
		resultStr, argsStr string
		errStr             *string
	}{resultStr, string(argsStr), errStr}, nil
}

// AgentStepObserver defines hooks for observing agent step events.
type AgentStepObserver interface {
	OnLLMContent(content string)
	OnToolCallStart(toolCallID, name string)
	OnToolCallUpdate(toolCallID, name, arguments string)
	OnToolCallExecute(toolCallID, name string, args map[string]interface{})
	OnToolCallResult(toolCallID, name string, result interface{}, err error)
	OnMaxStepsReached()
}

// StreamAgentStep streams the assistant's response for a single agent step, executes tool calls if present, and streams tool execution progress.
func StreamAgentStep(
	ctx context.Context,
	client *openai.Client,
	messages []openai.ChatCompletionMessage,
	model string,
	tools []openai.Tool,
	toolSchemas map[string]ToolSchema,
	observer AgentStepObserver, // new observer argument, can be nil
) (openai.ChatCompletionMessage, error) {
	var contentBuilder strings.Builder
	toolCallsMap := map[string]*openai.ToolCall{} // toolCallID -> ToolCall
	var lastToolCallID string                     // Track the last tool call ID for argument accumulation

	stream, err := client.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
		Model:    model,
		Messages: messages,
		Tools:    tools,
		Stream:   true,
	})
	if err != nil {
		return openai.ChatCompletionMessage{}, err
	}
	defer stream.Close()

	for {
		response, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return openai.ChatCompletionMessage{}, err
		}
		for _, choice := range response.Choices {
			// Stream assistant text
			if choice.Delta.Content != "" {
				contentBuilder.WriteString(choice.Delta.Content)
				if observer != nil {
					observer.OnLLMContent(choice.Delta.Content)
				}
			}

			// Handle tool call deltas robustly
			for _, tc := range choice.Delta.ToolCalls {
				if tc.ID != "" {
					// New tool call or new chunk for an existing one
					lastToolCallID = tc.ID
					if _, ok := toolCallsMap[tc.ID]; !ok {
						toolCallsMap[tc.ID] = &openai.ToolCall{
							ID:       tc.ID,
							Type:     tc.Type,
							Function: openai.FunctionCall{},
						}
						if observer != nil {
							observer.OnToolCallStart(tc.ID, tc.Function.Name)
						}
					}
				}
				// Use lastToolCallID for argument accumulation
				if lastToolCallID != "" {
					toolCall := toolCallsMap[lastToolCallID]
					updated := false
					if tc.Function.Name != "" && toolCall.Function.Name != tc.Function.Name {
						toolCall.Function.Name = tc.Function.Name
						updated = true
					}
					if tc.Function.Arguments != "" {
						toolCall.Function.Arguments += tc.Function.Arguments
						updated = true
					}
					if observer != nil && updated {
						observer.OnToolCallUpdate(lastToolCallID, toolCall.Function.Name, toolCall.Function.Arguments)
					}
				}
			}

			// If we get a tool calls finish reason, break out of the stream and reset state
			if choice.FinishReason == openai.FinishReasonToolCalls {
				lastToolCallID = ""
				break
			}
		}
	}

	// After stream, reconstruct tool calls
	var toolCalls []openai.ToolCall
	for _, tc := range toolCallsMap {
		toolCalls = append(toolCalls, *tc)
	}
	assistantMsg := openai.ChatCompletionMessage{
		Role:      openai.ChatMessageRoleAssistant,
		Content:   contentBuilder.String(),
		ToolCalls: toolCalls,
	}

	// If there are tool calls, execute them and stream progress
	for _, toolCall := range toolCalls {
		toolSchema, ok := toolSchemas[toolCall.Function.Name]
		if !ok {
			if observer != nil {
				observer.OnToolCallResult(toolCall.ID, toolCall.Function.Name, nil,
					fmt.Errorf("Unknown tool function: %s", toolCall.Function.Name))
			}
			continue
		}
		var args map[string]interface{}
		err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args)
		if err != nil {
			if observer != nil {
				observer.OnToolCallResult(toolCall.ID, toolCall.Function.Name, nil, err)
			}
			continue
		}
		if observer != nil {
			observer.OnToolCallExecute(toolCall.ID, toolCall.Function.Name, args)
		}
		result, err := toolSchema.Handler(toolCall.Function.Name, args)
		if observer != nil {
			observer.OnToolCallResult(toolCall.ID, toolCall.Function.Name, result, err)
		}
		if err != nil {
			continue
		}
		// resultJson, _ := json.Marshal(result)
	}

	return assistantMsg, nil
}

// streamingObserver wraps an AgentStepObserver and captures assistant tokens
// for persistence after streaming.
type streamingObserver struct {
	AgentStepObserver
	onAssistantToken func(token string)
}

func (o *streamingObserver) OnLLMContent(content string) {
	if o.AgentStepObserver != nil {
		o.AgentStepObserver.OnLLMContent(content)
	}
	if o.onAssistantToken != nil {
		o.onAssistantToken(content)
	}
}

// ChatWithPersistence handles chat with DB persistence for a project.
func (s *OpenAIChatService) ChatWithPersistence(
	ctx context.Context,
	projectID int64,
	userMessage string,
	observer AgentStepObserver,
	maxSteps int,
	sessionTracker *sessionchanges.Tracker,
) error {
	project, err := s.Queries.GetProject(ctx, projectID)
	if err != nil {
		return err
	}
	if s.ChatService == nil {
		return fmt.Errorf("ChatService is not configured")
	}
	// 1. Ensure conversation exists
	conv, err := s.ChatService.EnsureConversationForProject(ctx, projectID)
	if err != nil {
		return err
	}

	// 2. Add the new user message to the DB
	_, err = s.ChatService.AddMessage(ctx, conv.ID, nil, "user", userMessage)
	if err != nil {
		return err
	}

	// 3. Fetch all messages again (now includes the new user message)
	dbMessages, err := s.ChatService.GetMessages(ctx, conv.ID)
	if err != nil {
		return err
	}
	var messages []Message
	for _, m := range dbMessages {
		messages = append(messages, Message{
			Sender:  m.Sender,
			Content: m.Content,
		})
	}

	// 4. Call the streaming chat logic (this will stream and also generate the assistant reply)
	var assistantReply strings.Builder
	streamObserver := &streamingObserver{
		AgentStepObserver: observer,
		onAssistantToken: func(token string) {
			assistantReply.WriteString(token)
		},
	}
	err = s.StreamChat(ctx, project, conv.ID, messages, streamObserver, maxSteps, sessionTracker)
	if err != nil {
		return err
	}

	// 5. Store the assistant's reply in the DB
	_, err = s.ChatService.AddMessage(ctx, conv.ID, nil, "assistant", assistantReply.String())
	return err
}
