package log

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func createTestLogFile(t *testing.T) (string, func()) {
	t.Helper()

	// Create a temporary log file
	content := []string{
		"2024-03-20 10:00:00 INFO  Starting application",
		"2024-03-20 10:00:01 DEBUG Initializing database",
		"2024-03-20 10:00:02 ERROR Failed to connect to database",
		"2024-03-20 10:00:03 INFO  Retrying database connection",
		"2024-03-20 10:00:04 INFO  Database connected successfully",
	}

	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	err := os.WriteFile(logFile, []byte(strings.Join(content, "\n")), 0644)
	if err != nil {
		t.Fatalf("Failed to create test log file: %v", err)
	}

	cleanup := func() {
		os.Remove(logFile)
	}

	return logFile, cleanup
}

func TestReadLogRange(t *testing.T) {
	logFile, cleanup := createTestLogFile(t)
	defer cleanup()

	service := NewLogService()

	tests := []struct {
		name      string
		startLine int
		endLine   int
		want      int // expected number of lines
		wantErr   bool
	}{
		{
			name:      "valid range",
			startLine: 2,
			endLine:   4,
			want:      3,
			wantErr:   false,
		},
		{
			name:      "invalid start line",
			startLine: 0,
			endLine:   4,
			want:      0,
			wantErr:   true,
		},
		{
			name:      "invalid range",
			startLine: 4,
			endLine:   2,
			want:      0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := service.ReadLogRange(logFile, tt.startLine, tt.endLine)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadLogRange() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(got.Entries) != tt.want {
				t.Errorf("ReadLogRange() got %d entries, want %d", len(got.Entries), tt.want)
			}
		})
	}
}

func TestFilterLog(t *testing.T) {
	logFile, cleanup := createTestLogFile(t)
	defer cleanup()

	service := NewLogService()

	tests := []struct {
		name    string
		options FilterOptions
		want    int // expected number of lines
		wantErr bool
	}{
		{
			name: "filter by pattern",
			options: FilterOptions{
				Pattern:    "INFO",
				IgnoreCase: false,
			},
			want:    3,
			wantErr: false,
		},
		{
			name: "filter by line range",
			options: FilterOptions{
				StartLine: 1,
				EndLine:   3,
			},
			want:    3,
			wantErr: false,
		},
		{
			name: "filter by pattern and range",
			options: FilterOptions{
				Pattern:   "ERROR",
				StartLine: 1,
				EndLine:   4,
			},
			want:    1,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := service.FilterLog(logFile, tt.options)
			if (err != nil) != tt.wantErr {
				t.Errorf("FilterLog() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(got) != tt.want {
				t.Errorf("FilterLog() got %d entries, want %d", len(got), tt.want)
			}
		})
	}
}

func TestTailLog(t *testing.T) {
	logFile, cleanup := createTestLogFile(t)
	defer cleanup()

	service := NewLogService()

	tests := []struct {
		name    string
		n       int
		want    int // expected number of lines
		wantErr bool
	}{
		{
			name:    "tail 3 lines",
			n:       3,
			want:    3,
			wantErr: false,
		},
		{
			name:    "tail all lines",
			n:       10,
			want:    5, // total lines in test file
			wantErr: false,
		},
		{
			name:    "invalid line count",
			n:       0,
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := service.TailLog(logFile, tt.n)
			if (err != nil) != tt.wantErr {
				t.Errorf("TailLog() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(got) != tt.want {
				t.Errorf("TailLog() got %d entries, want %d", len(got), tt.want)
			}
		})
	}
}

func TestStreamLog(t *testing.T) {
	logFile, cleanup := createTestLogFile(t)
	defer cleanup()

	service := NewLogService()

	var buf bytes.Buffer
	err := service.StreamLog(logFile, FilterOptions{
		Pattern: "INFO",
	}, &buf)

	if err != nil {
		t.Errorf("StreamLog() error = %v", err)
		return
	}

	output := buf.String()
	count := strings.Count(output, "INFO")
	if count != 3 {
		t.Errorf("StreamLog() got %d INFO lines, want 3", count)
	}
}

func TestGetLogStats(t *testing.T) {
	logFile, cleanup := createTestLogFile(t)
	defer cleanup()

	service := NewLogService()

	stats, err := service.GetLogStats(logFile)
	if err != nil {
		t.Errorf("GetLogStats() error = %v", err)
		return
	}

	if stats["total_lines"].(int) != 5 {
		t.Errorf("GetLogStats() got %d total lines, want 5", stats["total_lines"])
	}

	if stats["size_bytes"].(int64) <= 0 {
		t.Errorf("GetLogStats() got invalid file size: %v", stats["size_bytes"])
	}
}
