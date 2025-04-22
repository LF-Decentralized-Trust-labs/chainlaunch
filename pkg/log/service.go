package log

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"sync"
	"syscall"
)

const (
	// Size threshold for using memory mapping (100MB)
	memoryMapThreshold = 100 * 1024 * 1024
	// Size of chunks for reading large files (1MB)
	chunkSize = 1024 * 1024
	// Maximum number of lines that can be tailed
	maxTailLines = 10000
)

// LogService handles operations related to log files
type LogService struct {
	indexCache map[string]*LineIndex
	cacheMutex sync.RWMutex
}

// LogEntry represents a single log entry
type LogEntry struct {
	LineNumber int    `json:"line_number"`
	Content    string `json:"content"`
}

// LogRange represents a range of log entries
type LogRange struct {
	StartLine int        `json:"start_line"`
	EndLine   int        `json:"end_line"`
	Entries   []LogEntry `json:"entries"`
}

// FilterOptions represents options for filtering log entries
type FilterOptions struct {
	Pattern    string // Regex pattern to match
	IgnoreCase bool   // Whether to ignore case in pattern matching
	StartLine  int    // Start line number (1-based)
	EndLine    int    // End line number (1-based)
}

// NewLogService creates a new instance of LogService
func NewLogService() *LogService {
	return &LogService{
		indexCache: make(map[string]*LineIndex),
	}
}

// getOrCreateIndex gets or creates a line index for a log file
func (s *LogService) getOrCreateIndex(filePath string) (*LineIndex, error) {
	s.cacheMutex.RLock()
	index, exists := s.indexCache[filePath]
	s.cacheMutex.RUnlock()

	if exists {
		return index, nil
	}

	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	// Check again in case another goroutine created the index
	if index, exists = s.indexCache[filePath]; exists {
		return index, nil
	}

	index, err := newLineIndex(filePath)
	if err != nil {
		return nil, err
	}

	s.indexCache[filePath] = index
	return index, nil
}

// ReadLogRange reads a specific range of lines from a log file
func (s *LogService) ReadLogRange(filePath string, startLine, endLine int) (*LogRange, error) {
	if startLine < 1 {
		return nil, fmt.Errorf("start line must be >= 1")
	}
	if endLine < startLine {
		return nil, fmt.Errorf("end line must be >= start line")
	}

	// Get file info
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	// For large files, use the index and memory mapping
	if info.Size() > memoryMapThreshold {
		return s.readLogRangeMMap(filePath, startLine, endLine)
	}

	// For smaller files, use the original implementation
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	currentLine := 0
	entries := []LogEntry{}

	for scanner.Scan() {
		currentLine++
		if currentLine < startLine {
			continue
		}
		if currentLine > endLine {
			break
		}
		entries = append(entries, LogEntry{
			LineNumber: currentLine,
			Content:    scanner.Text(),
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading log file: %w", err)
	}

	return &LogRange{
		StartLine: startLine,
		EndLine:   endLine,
		Entries:   entries,
	}, nil
}

// readLogRangeMMap reads a range of lines using memory mapping for large files
func (s *LogService) readLogRangeMMap(filePath string, startLine, endLine int) (*LogRange, error) {
	// Get or create the line index
	index, err := s.getOrCreateIndex(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get line index: %w", err)
	}

	// Get file offsets for the requested lines
	startOffset, err := index.getOffset(startLine)
	if err != nil {
		return nil, err
	}

	endOffset, err := index.getOffset(endLine + 1)
	if err != nil {
		// If endLine is the last line, use the file size as endOffset
		if endLine == index.getLineCount() {
			endOffset = index.sourceSize
		} else {
			return nil, err
		}
	}

	// Open the file
	file, err := os.OpenFile(filePath, os.O_RDONLY, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	// Memory map the file
	data, err := syscall.Mmap(int(file.Fd()), startOffset, int(endOffset-startOffset),
		syscall.PROT_READ, syscall.MAP_PRIVATE)
	if err != nil {
		return nil, fmt.Errorf("failed to memory map file: %w", err)
	}
	defer syscall.Munmap(data)

	// Process the mapped data
	entries := make([]LogEntry, 0, endLine-startLine+1)
	lineNum := startLine
	start := 0

	for i := 0; i < len(data); i++ {
		if data[i] == '\n' || i == len(data)-1 {
			end := i
			if i == len(data)-1 && data[i] != '\n' {
				end = i + 1
			}
			entries = append(entries, LogEntry{
				LineNumber: lineNum,
				Content:    string(data[start:end]),
			})
			lineNum++
			start = i + 1
		}
	}

	return &LogRange{
		StartLine: startLine,
		EndLine:   endLine,
		Entries:   entries,
	}, nil
}

// FilterLog filters log entries based on provided options
func (s *LogService) FilterLog(filePath string, options FilterOptions) ([]LogEntry, error) {
	// Get file info
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	// For large files, process in chunks
	if info.Size() > memoryMapThreshold {
		return s.filterLogChunked(filePath, options)
	}

	// For smaller files, use the original implementation
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	var pattern *regexp.Regexp
	if options.Pattern != "" {
		flags := ""
		if options.IgnoreCase {
			flags = "(?i)"
		}
		pattern, err = regexp.Compile(flags + options.Pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid regex pattern: %w", err)
		}
	}

	scanner := bufio.NewScanner(file)
	currentLine := 0
	var entries []LogEntry

	for scanner.Scan() {
		currentLine++

		if options.StartLine > 0 && currentLine < options.StartLine {
			continue
		}
		if options.EndLine > 0 && currentLine > options.EndLine {
			break
		}

		line := scanner.Text()

		if pattern != nil && !pattern.MatchString(line) {
			continue
		}

		entries = append(entries, LogEntry{
			LineNumber: currentLine,
			Content:    line,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading log file: %w", err)
	}

	return entries, nil
}

// filterLogChunked processes a large log file in chunks
func (s *LogService) filterLogChunked(filePath string, options FilterOptions) ([]LogEntry, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	var pattern *regexp.Regexp
	if options.Pattern != "" {
		flags := ""
		if options.IgnoreCase {
			flags = "(?i)"
		}
		pattern, err = regexp.Compile(flags + options.Pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid regex pattern: %w", err)
		}
	}

	var entries []LogEntry
	buffer := make([]byte, chunkSize)
	lineBuffer := []byte{}
	currentLine := 0
	offset := int64(0)

	for {
		n, err := file.ReadAt(buffer, offset)
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("error reading file chunk: %w", err)
		}

		chunk := buffer[:n]
		start := 0

		for i := 0; i < len(chunk); i++ {
			if chunk[i] == '\n' || (err == io.EOF && i == len(chunk)-1) {
				// Complete the line with the buffered content
				line := append(lineBuffer, chunk[start:i]...)
				lineBuffer = lineBuffer[:0]
				currentLine++

				if options.StartLine > 0 && currentLine < options.StartLine {
					start = i + 1
					continue
				}
				if options.EndLine > 0 && currentLine > options.EndLine {
					return entries, nil
				}

				lineStr := string(line)
				if pattern == nil || pattern.MatchString(lineStr) {
					entries = append(entries, LogEntry{
						LineNumber: currentLine,
						Content:    lineStr,
					})
				}
				start = i + 1
			}
		}

		// Buffer any incomplete line
		if start < len(chunk) {
			lineBuffer = append(lineBuffer, chunk[start:]...)
		}

		offset += int64(n)
		if err == io.EOF {
			break
		}
	}

	return entries, nil
}

// TailLog returns the last n lines of a log file
func (s *LogService) TailLog(filePath string, n int) ([]LogEntry, error) {
	if n < 1 {
		return nil, fmt.Errorf("number of lines must be >= 1")
	}

	// Add maximum limit check
	if n > maxTailLines {
		return nil, fmt.Errorf("requested number of lines exceeds maximum limit of %d", maxTailLines)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	// Create a ring buffer to store the last n lines
	lines := make([]string, n)
	lineNumbers := make([]int, n)
	currentIndex := 0
	totalLines := 0

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		totalLines++
		lines[currentIndex] = scanner.Text()
		lineNumbers[currentIndex] = totalLines
		currentIndex = (currentIndex + 1) % n
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading log file: %w", err)
	}

	// Create result slice with correct order
	var entries []LogEntry
	if totalLines < n {
		// File has fewer lines than requested
		for i := 0; i < totalLines; i++ {
			entries = append(entries, LogEntry{
				LineNumber: lineNumbers[i],
				Content:    lines[i],
			})
		}
	} else {
		// File has more lines than requested
		for i := 0; i < n; i++ {
			idx := (currentIndex + i) % n
			entries = append(entries, LogEntry{
				LineNumber: lineNumbers[idx],
				Content:    lines[idx],
			})
		}
	}

	return entries, nil
}

// StreamLog streams log file content with optional filtering
func (s *LogService) StreamLog(filePath string, options FilterOptions, writer io.Writer) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	var pattern *regexp.Regexp
	if options.Pattern != "" {
		flags := ""
		if options.IgnoreCase {
			flags = "(?i)"
		}
		pattern, err = regexp.Compile(flags + options.Pattern)
		if err != nil {
			return fmt.Errorf("invalid regex pattern: %w", err)
		}
	}

	scanner := bufio.NewScanner(file)
	currentLine := 0

	for scanner.Scan() {
		currentLine++

		// Skip lines before StartLine if specified
		if options.StartLine > 0 && currentLine < options.StartLine {
			continue
		}

		// Stop after EndLine if specified
		if options.EndLine > 0 && currentLine > options.EndLine {
			break
		}

		line := scanner.Text()

		// Apply pattern filtering if pattern is specified
		if pattern != nil {
			if !pattern.MatchString(line) {
				continue
			}
		}

		// Write the line to the writer
		if _, err := fmt.Fprintln(writer, line); err != nil {
			return fmt.Errorf("error writing to output: %w", err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading log file: %w", err)
	}

	return nil
}

// GetLogStats returns statistics about a log file
func (s *LogService) GetLogStats(filePath string) (map[string]interface{}, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	stats := map[string]interface{}{
		"size_bytes": fileInfo.Size(),
		"modified":   fileInfo.ModTime(),
	}

	// Count total lines
	scanner := bufio.NewScanner(file)
	lineCount := 0
	for scanner.Scan() {
		lineCount++
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading log file: %w", err)
	}

	stats["total_lines"] = lineCount

	return stats, nil
}
