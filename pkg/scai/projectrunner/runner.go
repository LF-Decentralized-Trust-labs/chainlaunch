package projectrunner

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/chainlaunch/chainlaunch/pkg/db"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/errdefs"
	"github.com/docker/go-connections/nat"
)

type Runner struct {
	docker  *client.Client
	queries *db.Queries
}

func NewRunner(queries *db.Queries) *Runner {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	return &Runner{
		docker:  cli,
		queries: queries,
	}
}

func (r *Runner) Start(ctx context.Context, projectID string, projectDir string, imageName string, port int, env map[string]string, args ...string) (int, error) {
	containerName := fmt.Sprintf("project-%s", projectID)

	// Remove any existing container with the same name
	dockerContainer, err := r.docker.ContainerInspect(ctx, containerName)
	if err == nil {
		_ = r.docker.ContainerRemove(ctx, dockerContainer.ID, container.RemoveOptions{Force: true})
	}

	// Check if image exists locally
	_, err = r.docker.ImageInspect(ctx, imageName)
	if errdefs.IsNotFound(err) {
		// Pull the image if not found locally
		rc, err := r.docker.ImagePull(ctx, imageName, image.PullOptions{})
		if err != nil {
			return 0, fmt.Errorf("failed to pull image %s: %w", imageName, err)
		}
		_, err = io.Copy(io.Discard, rc)
		if err != nil {
			return 0, fmt.Errorf("failed to pull image %s: %w", imageName, err)
		}
	} else if err != nil {
		return 0, fmt.Errorf("failed to inspect image %s: %w", imageName, err)
	}

	// Create container host config with port binding
	containerHostConfig := &container.HostConfig{
		PortBindings: nat.PortMap{
			nat.Port("4000/tcp"): []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: strconv.Itoa(port),
				},
			},
		},
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: projectDir,
				Target: "/app",
			},
		},
	}

	// Convert environment map to slice
	envSlice := make([]string, 0, len(env))
	for k, v := range env {
		envSlice = append(envSlice, fmt.Sprintf("%s=%s", k, v))
	}

	containerConfig := &container.Config{
		Image:      imageName,
		Cmd:        args,
		Tty:        false,
		WorkingDir: "/app",
		Env:        envSlice,
		ExposedPorts: nat.PortSet{
			nat.Port("4000/tcp"): struct{}{},
		},
	}
	resp, err := r.docker.ContainerCreate(ctx, containerConfig, containerHostConfig, nil, nil, containerName)
	if err != nil {
		return 0, err
	}
	if err := r.docker.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return 0, err
	}

	// Wait for container to be ready
	time.Sleep(2 * time.Second)

	// Update DB with running status
	idInt64, _ := parseProjectID(projectID)
	now := time.Now()
	err = r.queries.UpdateProjectContainerInfo(ctx, &db.UpdateProjectContainerInfoParams{
		ContainerID:   sqlNullString(resp.ID),
		ContainerName: sqlNullString(containerName),
		Status:        sqlNullString("running"),
		LastStartedAt: sqlNullTime(now),
		LastStoppedAt: sqlNullTimeZero(),
		ContainerPort: sql.NullInt64{Int64: int64(port), Valid: true},
		ID:            idInt64,
	})
	if err != nil {
		return 0, err
	}
	return port, nil
}

func (r *Runner) Stop(projectID string) error {
	ctx := context.Background()
	idInt64, _ := parseProjectID(projectID)
	proj, err := r.queries.GetProject(ctx, idInt64)
	if err != nil {
		return err
	}
	if !proj.Status.Valid || proj.Status.String != "running" {
		return nil
	}
	timeout := 5
	if err := r.docker.ContainerStop(ctx, proj.ContainerID.String, container.StopOptions{Timeout: &timeout}); err != nil {
		return err
	}
	now := time.Now()
	return r.queries.UpdateProjectContainerInfo(ctx, &db.UpdateProjectContainerInfoParams{
		ContainerID:   proj.ContainerID,
		ContainerName: proj.ContainerName,
		Status:        sqlNullString("stopped"),
		LastStartedAt: proj.LastStartedAt,
		LastStoppedAt: sqlNullTime(now),
		ID:            idInt64,
	})
}

func (r *Runner) Restart(projectID, dir, image string, args ...string) error {
	r.Stop(projectID)
	_, err := r.Start(context.Background(), projectID, dir, image, 4000, nil, args...)
	return err
}

func (r *Runner) GetLogs(projectID string) (string, error) {
	ctx := context.Background()
	idInt64, _ := parseProjectID(projectID)
	proj, err := r.queries.GetProject(ctx, idInt64)
	if err != nil {
		return "", err
	}

	// Check if container exists
	_, err = r.docker.ContainerInspect(ctx, proj.ContainerID.String)
	if err != nil {
		return "", fmt.Errorf("container not found for project %s: %w", projectID, err)
	}

	reader, err := r.docker.ContainerLogs(ctx, proj.ContainerID.String, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Timestamps: true,
		Tail:       "1000",
	})
	if err != nil {
		return "", err
	}
	defer reader.Close()

	var logs []byte
	header := make([]byte, 8)
	for {
		// Read the 8-byte header
		_, err := io.ReadFull(reader, header)
		if err != nil {
			if err != io.EOF {
				return "", fmt.Errorf("failed to read docker log header: %w", err)
			}
			break
		}
		// Get the payload length
		length := int(uint32(header[4])<<24 | uint32(header[5])<<16 | uint32(header[6])<<8 | uint32(header[7]))
		if length == 0 {
			continue
		}
		// Read the payload
		payload := make([]byte, length)
		_, err = io.ReadFull(reader, payload)
		if err != nil {
			if err != io.EOF {
				return "", fmt.Errorf("failed to read docker log payload: %w", err)
			}
			break
		}
		logs = append(logs, payload...)
	}
	return string(logs), nil
}

func (r *Runner) StreamLogs(ctx context.Context, projectID string, onLog func([]byte)) error {
	idInt64, _ := parseProjectID(projectID)
	proj, err := r.queries.GetProject(ctx, idInt64)
	if err != nil {
		return err
	}

	// Check if container exists
	_, err = r.docker.ContainerInspect(ctx, proj.ContainerID.String)
	if err != nil {
		return fmt.Errorf("container not found for project %s: %w", projectID, err)
	}
	reader, err := r.docker.ContainerLogs(ctx, proj.ContainerID.String, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Timestamps: true,
		Tail:       "100",
		Follow:     true,
	})
	if err != nil {
		return err
	}
	defer reader.Close()

	header := make([]byte, 8)
	for {
		// Read the 8-byte header
		_, err := io.ReadFull(reader, header)
		if err != nil {
			if err != io.EOF {
				return fmt.Errorf("failed to read docker log header: %w", err)
			}
			return nil
		}
		// Get the payload length
		length := int(uint32(header[4])<<24 | uint32(header[5])<<16 | uint32(header[6])<<8 | uint32(header[7]))
		if length == 0 {
			continue
		}
		// Read the payload
		payload := make([]byte, length)
		_, err = io.ReadFull(reader, payload)
		if err != nil {
			if err != io.EOF {
				return fmt.Errorf("failed to read docker log payload: %w", err)
			}
			return nil
		}

		select {
		case <-ctx.Done():
			return nil
		default:
			onLog(payload)
		}
	}
}

// Helpers
func sqlNullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}
func sqlNullTime(t time.Time) sql.NullTime {
	return sql.NullTime{Time: t, Valid: !t.IsZero()}
}
func sqlNullTimeZero() sql.NullTime {
	return sql.NullTime{Valid: false}
}
func parseProjectID(id string) (int64, error) {
	var i int64
	_, err := fmt.Sscanf(id, "%d", &i)
	return i, err
}
