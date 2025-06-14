package files

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
)

type FilesService struct {
}

func NewFilesService() *FilesService {
	return &FilesService{}
}

// Placeholder for project validation
func (s *FilesService) validateProject(project string) error {
	if project == "" {
		return errors.New("project is required")
	}
	// TODO: Implement real project existence check
	return nil
}

func (s *FilesService) ListFiles(project, dir string) ([]string, error) {
	if err := s.validateProject(project); err != nil {
		return nil, err
	}
	if dir == "" {
		dir = "."
	}
	base := filepath.Join(project, dir)
	entries, err := ioutil.ReadDir(base)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}
	return files, nil
}

func (s *FilesService) ReadFile(project, path string) ([]byte, error) {
	if err := s.validateProject(project); err != nil {
		return nil, err
	}
	if path == "" {
		return nil, errors.New("path is required")
	}
	base := filepath.Join(project, path)
	return ioutil.ReadFile(base)
}

func (s *FilesService) WriteFile(project, path string, data []byte) error {
	if err := s.validateProject(project); err != nil {
		return err
	}
	if path == "" {
		return errors.New("path is required")
	}
	base := filepath.Join(project, path)
	return ioutil.WriteFile(base, data, 0644)
}

func (s *FilesService) DeleteFile(project, path string) error {
	if err := s.validateProject(project); err != nil {
		return err
	}
	if path == "" {
		return errors.New("path is required")
	}
	base := filepath.Join(project, path)
	return os.Remove(base)
}

// ListEntries returns files, directories, and skipped directories in a given directory
func (s *FilesService) ListEntries(project, dir string) (files, directories, skipped []string, err error) {
	if err := s.validateProject(project); err != nil {
		return nil, nil, nil, err
	}
	if dir == "" {
		dir = "."
	}
	base := filepath.Join(project, dir)
	entries, err := ioutil.ReadDir(base)
	if err != nil {
		return nil, nil, nil, err
	}
	var filesOut, dirsOut, skippedOut []string
	const maxEntries = 1000
	skipList := map[string]struct{}{"node_modules": {}}
	for _, entry := range entries {
		if entry.IsDir() {
			name := entry.Name()
			if _, skip := skipList[name]; skip {
				skippedOut = append(skippedOut, name)
				continue
			}
			dirPath := filepath.Join(base, name)
			dirEntries, err := ioutil.ReadDir(dirPath)
			if err == nil && len(dirEntries) > maxEntries {
				skippedOut = append(skippedOut, name)
				continue
			}
			dirsOut = append(dirsOut, name)
		} else {
			filesOut = append(filesOut, entry.Name())
		}
	}
	return filesOut, dirsOut, skippedOut, nil
}
