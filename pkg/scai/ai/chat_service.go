package ai

import (
	"context"
	"database/sql"
	"time"

	"github.com/chainlaunch/chainlaunch/pkg/db"
)

type ChatService struct {
	Queries *db.Queries
}

type Conversation struct {
	ID        int64
	ProjectID int64
	StartedAt time.Time
}

func NewChatService(queries *db.Queries) *ChatService {
	return &ChatService{Queries: queries}
}

// EnsureConversationForProject returns the default conversation for a project, creating it if needed.
func (s *ChatService) EnsureConversationForProject(ctx context.Context, projectID int64) (Conversation, error) {
	conv, err := s.Queries.GetDefaultConversationForProject(ctx, projectID)
	if err == sql.ErrNoRows {
		// Create new conversation
		row, err := s.Queries.CreateConversation(ctx, projectID)
		if err != nil {
			return Conversation{}, err
		}
		return Conversation{
			ID:        row.ID,
			ProjectID: row.ProjectID,
			StartedAt: row.StartedAt,
		}, nil
	} else if err != nil {
		return Conversation{}, err
	}
	return Conversation{
		ID:        conv.ID,
		ProjectID: conv.ProjectID,
		StartedAt: conv.StartedAt,
	}, nil
}

// AddMessage stores a message in the conversation. Accepts optional parentID.
func (s *ChatService) AddMessage(ctx context.Context, conversationID int64, parentID *int64, sender, content string) (*db.Message, error) {
	var parentNull sql.NullInt64
	if parentID != nil {
		parentNull = sql.NullInt64{Int64: *parentID, Valid: true}
	}
	row, err := s.Queries.InsertMessage(ctx, &db.InsertMessageParams{
		ConversationID: conversationID,
		ParentID:       parentNull,
		Sender:         sender,
		Content:        content,
	})
	if err != nil {
		return nil, err
	}
	return row, nil
}

// GetMessages returns all messages for a conversation.
func (s *ChatService) GetMessages(ctx context.Context, conversationID int64) ([]*db.Message, error) {
	return s.Queries.ListMessagesForConversation(ctx, conversationID)
}

// AddToolCall stores a tool call for a message.
func (s *ChatService) AddToolCall(ctx context.Context, messageID int64, toolName, arguments, result string, errStr *string) (*db.ToolCall, error) {
	var resultNull sql.NullString
	if result != "" {
		resultNull = sql.NullString{String: result, Valid: true}
	}
	var errorNull sql.NullString
	if errStr != nil {
		errorNull = sql.NullString{String: *errStr, Valid: true}
	}
	return s.Queries.InsertToolCall(ctx, &db.InsertToolCallParams{
		MessageID: messageID,
		ToolName:  toolName,
		Arguments: arguments,
		Result:    resultNull,
		Error:     errorNull,
	})
}
