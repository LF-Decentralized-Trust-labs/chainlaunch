package log

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"
	"sync"
)

// LineIndex stores the byte offsets of lines in a log file
type LineIndex struct {
	offsets    []int64 // Byte offsets for each line
	indexPath  string  // Path to the index file
	sourceSize int64   // Size of the source file when indexed
	mutex      sync.RWMutex
}

// newLineIndex creates a new line index for a log file
func newLineIndex(logPath string) (*LineIndex, error) {
	indexPath := logPath + ".idx"
	index := &LineIndex{
		indexPath: indexPath,
	}

	// Try to load existing index
	if err := index.load(); err == nil {
		// Verify if the index is still valid
		if info, err := os.Stat(logPath); err == nil {
			if info.Size() == index.sourceSize {
				return index, nil
			}
		}
	}

	// Create new index if loading failed or index is invalid
	if err := index.build(logPath); err != nil {
		return nil, err
	}

	return index, nil
}

// build creates a new index for the log file
func (idx *LineIndex) build(logPath string) error {
	idx.mutex.Lock()
	defer idx.mutex.Unlock()

	file, err := os.Open(logPath)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	// Get file size
	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}
	idx.sourceSize = info.Size()

	// Create a buffered reader
	reader := bufio.NewReader(file)
	var offset int64
	idx.offsets = make([]int64, 0, 1000) // Pre-allocate space for 1000 lines

	// Record the offset of the first line
	idx.offsets = append(idx.offsets, 0)

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			break // End of file or error
		}
		offset += int64(len(line))
		idx.offsets = append(idx.offsets, offset)
	}

	// Save the index to disk
	return idx.save()
}

// save writes the index to disk
func (idx *LineIndex) save() error {
	file, err := os.Create(idx.indexPath)
	if err != nil {
		return fmt.Errorf("failed to create index file: %w", err)
	}
	defer file.Close()

	// Write source file size
	if err := binary.Write(file, binary.LittleEndian, idx.sourceSize); err != nil {
		return fmt.Errorf("failed to write source size: %w", err)
	}

	// Write number of offsets
	numOffsets := int64(len(idx.offsets))
	if err := binary.Write(file, binary.LittleEndian, numOffsets); err != nil {
		return fmt.Errorf("failed to write offset count: %w", err)
	}

	// Write offsets
	for _, offset := range idx.offsets {
		if err := binary.Write(file, binary.LittleEndian, offset); err != nil {
			return fmt.Errorf("failed to write offset: %w", err)
		}
	}

	return nil
}

// load reads the index from disk
func (idx *LineIndex) load() error {
	idx.mutex.Lock()
	defer idx.mutex.Unlock()

	file, err := os.Open(idx.indexPath)
	if err != nil {
		return fmt.Errorf("failed to open index file: %w", err)
	}
	defer file.Close()

	// Read source file size
	if err := binary.Read(file, binary.LittleEndian, &idx.sourceSize); err != nil {
		return fmt.Errorf("failed to read source size: %w", err)
	}

	// Read number of offsets
	var numOffsets int64
	if err := binary.Read(file, binary.LittleEndian, &numOffsets); err != nil {
		return fmt.Errorf("failed to read offset count: %w", err)
	}

	// Read offsets
	idx.offsets = make([]int64, numOffsets)
	for i := range idx.offsets {
		if err := binary.Read(file, binary.LittleEndian, &idx.offsets[i]); err != nil {
			return fmt.Errorf("failed to read offset: %w", err)
		}
	}

	return nil
}

// getOffset returns the byte offset for a given line number (1-based)
func (idx *LineIndex) getOffset(lineNum int) (int64, error) {
	idx.mutex.RLock()
	defer idx.mutex.RUnlock()

	if lineNum < 1 || lineNum > len(idx.offsets) {
		return 0, fmt.Errorf("line number out of range")
	}
	return idx.offsets[lineNum-1], nil
}

// getLineCount returns the total number of lines in the indexed file
func (idx *LineIndex) getLineCount() int {
	idx.mutex.RLock()
	defer idx.mutex.RUnlock()
	return len(idx.offsets)
}
