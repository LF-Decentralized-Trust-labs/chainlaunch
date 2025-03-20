package logger

import (
	"testing"
)

func TestLogger(t *testing.T) {
	// Test default logger creation
	logger := NewDefault()
	if logger == nil {
		t.Fatal("Failed to create default logger")
	}

	// Test logging with structured fields
	logger.Info("test message",
		"key1", "value1",
		"key2", 123,
	)

	// Test with additional context
	contextLogger := logger.With(
		"requestID", "123",
		"userID", "456",
	)
	contextLogger.Info("test with context")

	// Test different log levels
	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warning message")
	logger.Error("error message")
}
