package ai

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/chainlaunch/chainlaunch/pkg/db"
	"github.com/chainlaunch/chainlaunch/pkg/logger"
	"github.com/sashabaranov/go-openai"
)

// testAgentStepObserver implements AgentStepObserver for testing.
type executedTool struct {
	Name   string
	Args   map[string]interface{}
	Result interface{}
}

type testAgentStepObserver struct {
	ToolExecuted    map[string]bool
	Executed        []executedTool
	currentArgs     map[string]interface{}
	maxStepsReached bool
}

func (o *testAgentStepObserver) OnMaxStepsReached() {
	o.maxStepsReached = true
}
func (o *testAgentStepObserver) OnLLMContent(content string)                         {}
func (o *testAgentStepObserver) OnToolCallStart(toolCallID, name string)             {}
func (o *testAgentStepObserver) OnToolCallUpdate(toolCallID, name, arguments string) {}
func (o *testAgentStepObserver) OnToolCallExecute(toolCallID, name string, args map[string]interface{}) {
	if o.ToolExecuted == nil {
		o.ToolExecuted = make(map[string]bool)
	}
	o.ToolExecuted[name] = true
	if o.currentArgs == nil {
		o.currentArgs = make(map[string]interface{})
	}
	for k, v := range args {
		o.currentArgs[k] = v
	}
}
func (o *testAgentStepObserver) OnToolCallResult(toolCallID, name string, result interface{}, err error) {
	o.Executed = append(o.Executed, executedTool{
		Name:   name,
		Args:   o.currentArgs,
		Result: result,
	})
	o.currentArgs = nil
}

func TestStreamAgentStep_EchoTool(t *testing.T) {
	// ctx := context.Background()
	contentIndexTS := `
import { readFile } from 'fs/promises';

export async function readFile(path: string) {
	return await readFile(path, 'utf8');
}
	`
	result := map[string]interface{}{"content": contentIndexTS}

	// Use the default tool schemas from ai.go
	toolSchemas := GetDefaultToolSchemas(".")
	for _, tool := range toolSchemas {
		// For testing, override the handlers to return canned results
		switch tool.Name {
		case "read_file":
			tool.Handler = func(projectRoot string, args map[string]interface{}) (interface{}, error) {
				return result, nil
			}
		case "write_file":
			tool.Handler = func(projectRoot string, args map[string]interface{}) (interface{}, error) {
				return map[string]interface{}{"result": "ok", "path": args["path"]}, nil
			}
		}
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	client := openai.NewClient(apiKey)

	observer := &testAgentStepObserver{}
	logger := logger.NewDefault()
	chatClient := &OpenAIChatService{
		Client: client,
		Logger: logger,
	}
	msgs := []Message{
		{
			Role:    "user",
			Content: "Please read the file 'index.ts' and modify the contents to add a new function 'writeFile' that writes the contents to a file.",
		},
	}
	err := chatClient.StreamChat(context.Background(), &db.Project{ID: 1, Name: "test"}, 1, msgs, observer, 0)
	if err != nil {
		t.Fatalf("StreamAgentStep failed: %v", err)
	}
	if !observer.ToolExecuted["read_file"] {
		t.Errorf("Expected read_file tool to be executed, but it was not")
	}
	if len(observer.Executed) == 0 {
		t.Fatalf("Expected at least one tool execution recorded")
	}
	exec := observer.Executed[0]
	if exec.Name != "read_file" {
		t.Errorf("Expected executed tool name to be 'read_file', got %q", exec.Name)
	}
	if exec.Args["path"] == nil {
		t.Errorf("Expected tool args to include 'path', got: %#v", exec.Args)
	}
	if exec.Result == nil {
		t.Errorf("Expected tool result to be non-nil")
	}
	resultJSON, _ := json.Marshal(result)
	actualJSON, _ := json.Marshal(exec.Result)
	if string(resultJSON) != string(actualJSON) {
		t.Errorf("Expected tool result %s, got %s", string(resultJSON), string(actualJSON))
	}

	if !observer.ToolExecuted["write_file"] {
		t.Errorf("Expected write_file tool to be executed, but it was not")
	}
	writeExec := observer.Executed[1] // Assuming write_file is called after read_file
	if writeExec.Name != "write_file" {
		t.Errorf("Expected executed tool name to be 'write_file', got %q", writeExec.Name)
	}
	if writeExec.Args["path"] == nil {
		t.Errorf("Expected tool args to include 'path', got: %#v", writeExec.Args)
	}
	if writeExec.Args["content"] == nil {
		t.Errorf("Expected tool args to include 'content', got: %#v", writeExec.Args)
	}
	if writeExec.Result == nil {
		t.Errorf("Expected tool result to be non-nil")
	}
}
