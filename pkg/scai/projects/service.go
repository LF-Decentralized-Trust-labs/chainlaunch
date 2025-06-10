package projects

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"crypto/rand"
	"encoding/hex"
	"strings"

	"github.com/chainlaunch/chainlaunch/pkg/db"
	"github.com/chainlaunch/chainlaunch/pkg/scai/projectrunner"
	"github.com/chainlaunch/chainlaunch/pkg/scai/versionmanagement"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

type ProjectsService struct {
	Queries     *db.Queries
	Runner      *projectrunner.Runner
	ProjectsDir string
}

type Project struct {
	ID            int64   `json:"id" example:"1" description:"Project ID"`
	Name          string  `json:"name" example:"myproject" description:"Project name"`
	Slug          string  `json:"slug" example:"myproject-abc12" description:"Project slug (used for proxying and folder name)"`
	Description   string  `json:"description" example:"A sample project" description:"Project description"`
	Boilerplate   string  `json:"boilerplate" example:"go-basic" description:"Boilerplate template used for scaffolding"`
	Status        string  `json:"status" example:"running" description:"Project container status (running/stopped/etc)"`
	LastStartedAt *string `json:"lastStartedAt,omitempty" description:"Last time the project was started (RFC3339)"`
	LastStoppedAt *string `json:"lastStoppedAt,omitempty" description:"Last time the project was stopped (RFC3339)"`
	ContainerPort *int    `json:"containerPort,omitempty" description:"Host port mapped to the container, if running"`
}

// ProjectProcessManager manages running server processes for projects
var projectProcessManager = struct {
	mu      sync.Mutex
	servers map[int64]*exec.Cmd
}{servers: make(map[int64]*exec.Cmd)}

func NewProjectsService(queries *db.Queries, runner *projectrunner.Runner, projectsDir string) *ProjectsService {
	return &ProjectsService{Queries: queries, Runner: runner, ProjectsDir: projectsDir}
}

func getReqID(ctx context.Context) string {
	if reqID, ok := ctx.Value(middleware.RequestIDKey).(string); ok {
		return reqID
	}
	return ""
}

// Helper to copy a directory recursively
func copyDir(src string, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		destPath := filepath.Join(dst, relPath)
		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		}
		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()
		dstFile, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
		if err != nil {
			return err
		}
		defer dstFile.Close()
		_, err = io.Copy(dstFile, srcFile)
		return err
	})
}

func generateShortGUID(n int) (string, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b)[:n], nil
}

func generateSlug(name string, queries *db.Queries, ctx context.Context) (string, error) {
	base := strings.ToLower(strings.ReplaceAll(name, " ", "-"))
	for {
		guid, err := generateShortGUID(5)
		if err != nil {
			return "", err
		}
		slug := base + "-" + guid
		// Check uniqueness
		_, err = queries.GetProjectBySlug(ctx, slug)
		if err != nil && err != sql.ErrNoRows {
			return "", err
		}
		if err == sql.ErrNoRows {
			return slug, nil
		}
		// else, collision, try again
	}
}

func (s *ProjectsService) CreateProject(ctx context.Context, name, description, boilerplate string) (Project, error) {
	slug, err := generateSlug(name, s.Queries, ctx)
	if err != nil {
		return Project{}, err
	}
	proj, err := s.Queries.CreateProject(ctx, &db.CreateProjectParams{
		Name:        name,
		Description: sql.NullString{String: description, Valid: description != ""},
		Boilerplate: sql.NullString{String: boilerplate, Valid: boilerplate != ""},
		Slug:        slug,
	})
	if err != nil {
		zap.L().Error("DB error in CreateProject", zap.String("name", name), zap.Error(err), zap.String("request_id", getReqID(ctx)))
		return Project{}, err
	}
	zap.L().Info("created project in DB", zap.Int64("id", proj.ID), zap.String("name", proj.Name), zap.String("slug", proj.Slug), zap.String("request_id", getReqID(ctx)))

	// Copy boilerplate folder if specified
	if boilerplate != "" {
		boilerplateSrc := filepath.Join("/Users/davidviejo/poc/chain-ai-v0/boilerplates", boilerplate)
		projectDst := filepath.Join(s.ProjectsDir, slug)
		if err := copyDir(boilerplateSrc, projectDst); err != nil {
			zap.L().Error("failed to copy boilerplate", zap.String("src", boilerplateSrc), zap.String("dst", projectDst), zap.Error(err))
			return Project{}, err
		}
		// Ensure git repository is initialized before committing
		gitDir := filepath.Join(projectDst, ".git")
		if _, err := os.Stat(gitDir); os.IsNotExist(err) {
			// Initialize the repo using go-git
			_, err := versionmanagement.InitRepo(projectDst)
			if err != nil {
				zap.L().Error("failed to initialize git repo", zap.Error(err))
			}
		}
		vm := versionmanagement.NewDefaultManager()
		cwd, _ := os.Getwd()
		if err := os.Chdir(projectDst); err == nil {
			err = vm.CommitChange(ctx, "Initial commit for project "+name)
			if err != nil {
				zap.L().Error("failed to commit", zap.Error(err))
			}
			if err := os.Chdir(cwd); err != nil {
				zap.L().Error("failed to return to original directory", zap.Error(err))
			}
		}
	}
	return dbProjectToAPI(proj), nil
}

func (s *ProjectsService) ListProjects(ctx context.Context) ([]Project, error) {
	dbProjects, err := s.Queries.ListProjects(ctx)
	if err != nil {
		zap.L().Error("DB error in ListProjects", zap.Error(err), zap.String("request_id", getReqID(ctx)))
		return nil, err
	}
	var projects []Project
	for _, p := range dbProjects {
		projects = append(projects, dbProjectToAPI(p))
	}
	zap.L().Info("listed projects from DB", zap.Int("count", len(projects)), zap.String("request_id", getReqID(ctx)))
	return projects, nil
}

func (s *ProjectsService) GetProject(ctx context.Context, id int64) (Project, error) {
	p, err := s.Queries.GetProject(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			zap.L().Warn("project not found in DB", zap.Int64("id", id), zap.String("request_id", getReqID(ctx)))
			return Project{}, ErrNotFound
		}
		zap.L().Error("DB error in GetProject", zap.Int64("id", id), zap.Error(err), zap.String("request_id", getReqID(ctx)))
		return Project{}, err
	}
	zap.L().Info("got project from DB", zap.Int64("id", p.ID), zap.String("name", p.Name), zap.String("request_id", getReqID(ctx)))
	return dbProjectToAPI(p), nil
}

func dbProjectToAPI(p *db.Project) Project {
	var started, stopped *string
	if p.LastStartedAt.Valid {
		ts := p.LastStartedAt.Time.UTC().Format(time.RFC3339)
		started = &ts
	}
	if p.LastStoppedAt.Valid {
		ts := p.LastStoppedAt.Time.UTC().Format(time.RFC3339)
		stopped = &ts
	}
	var containerPort *int
	if p.ContainerPort.Valid {
		v := int(p.ContainerPort.Int64)
		containerPort = &v
	}
	return Project{
		ID:            p.ID,
		Name:          p.Name,
		Slug:          p.Slug,
		Description:   p.Description.String,
		Boilerplate:   p.Boilerplate.String,
		Status:        p.Status.String,
		LastStartedAt: started,
		LastStoppedAt: stopped,
		ContainerPort: containerPort,
	}
}

var ErrNotFound = errors.New("not found")

// StartProjectServer starts the server process for a project
func (s *ProjectsService) StartProjectServer(ctx context.Context, projectID int64, boilerplate, projectName string) error {
	project, err := s.GetProject(ctx, projectID)
	if err != nil {
		return err
	}
	projectDir, err := filepath.Abs(filepath.Join(s.ProjectsDir, project.Slug))
	if err != nil {
		return err
	}
	_, args, image, ok := projectrunner.GetBoilerplateRunner(project.Boilerplate)
	if !ok {
		return errors.New("unsupported boilerplate for start")
	}
	_, err = s.Runner.Start(fmt.Sprintf("%d", projectID), projectDir, image, args...)
	return err
}

// StopProjectServer stops the server process for a project
func (s *ProjectsService) StopProjectServer(projectID int64) error {
	return s.Runner.Stop(fmt.Sprintf("%d", projectID))
}

func (s *ProjectsService) GetProjectLogs(ctx context.Context, projectID int64) (string, error) {
	return s.Runner.GetLogs(fmt.Sprintf("%d", projectID))
}

func (s *ProjectsService) StreamProjectLogs(ctx context.Context, projectID int64, onLog func([]byte)) error {
	return s.Runner.StreamLogs(ctx, fmt.Sprintf("%d", projectID), onLog)
}
