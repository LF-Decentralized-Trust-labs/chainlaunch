package projects

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/chainlaunch/chainlaunch/pkg/db"
	"github.com/chainlaunch/chainlaunch/pkg/scai/versionmanagement"
)

const ProjectsRoot = "./data/projects"

func ListProjects(q *db.Queries, ctx context.Context) ([]*db.Project, error) {
	return q.ListProjects(ctx)
}

func CreateProject(q *db.Queries, ctx context.Context, name, description string) (*db.Project, error) {
	proj, err := q.CreateProject(ctx, &db.CreateProjectParams{Name: name, Description: sql.NullString{String: description, Valid: description != ""}})
	if err != nil {
		return proj, err
	}
	projDir := filepath.Join(ProjectsRoot, name)
	if err := os.MkdirAll(projDir, 0755); err != nil {
		return proj, fmt.Errorf("failed to create project directory: %w", err)
	}

	// Use versionmanagement to initialize the repo and make the initial commit
	vm := versionmanagement.NewDefaultManager()
	// Create a .gitkeep file so there's something to commit
	gitkeepPath := filepath.Join(projDir, ".gitkeep")
	if err := os.WriteFile(gitkeepPath, []byte{}, 0644); err != nil {
		return proj, fmt.Errorf("failed to create .gitkeep: %w", err)
	}
	// Initialize the repo (if not already initialized)
	if _, err := os.Stat(filepath.Join(projDir, ".git")); os.IsNotExist(err) {
		if err := os.Chdir(projDir); err != nil {
			return proj, fmt.Errorf("failed to change dir: %w", err)
		}
		if err := vm.CommitChange(ctx, "Initial repository"); err != nil {
			return proj, fmt.Errorf("failed to initialize version management: %w", err)
		}
		if err := os.Chdir("../../.."); err != nil { // Return to root
			return proj, fmt.Errorf("failed to return to root dir: %w", err)
		}
	}
	return proj, nil
}

func DeleteProject(q *db.Queries, ctx context.Context, id int64, name string) error {
	if err := q.DeleteProject(ctx, id); err != nil {
		return err
	}
	projDir := filepath.Join(ProjectsRoot, name)
	if err := os.RemoveAll(projDir); err != nil {
		return fmt.Errorf("failed to remove project directory: %w", err)
	}
	// Commit the deletion using versionmanagement
	vm := versionmanagement.NewDefaultManager()
	cwd, _ := os.Getwd()
	if err := os.Chdir(ProjectsRoot); err == nil {
		_ = vm.CommitChange(ctx, "Deleted project "+name)
		_ = os.Chdir(cwd)
	}
	return nil
}
