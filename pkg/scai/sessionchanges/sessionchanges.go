package sessionchanges

import (
	"sync"
)

var (
	mu            sync.Mutex
	modifiedFiles = make(map[string]struct{})
)

// RegisterChange registers a file as changed (created/modified/deleted) during the session.
func RegisterChange(filePath string) {
	mu.Lock()
	defer mu.Unlock()
	modifiedFiles[filePath] = struct{}{}
}

// GetAndResetChanges returns the list of changed files and resets the tracker.
func GetAndResetChanges() []string {
	mu.Lock()
	defer mu.Unlock()
	files := make([]string, 0, len(modifiedFiles))
	for f := range modifiedFiles {
		files = append(files, f)
	}
	modifiedFiles = make(map[string]struct{})
	return files
}
