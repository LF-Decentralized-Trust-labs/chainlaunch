package plugin

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	plugintypes "github.com/chainlaunch/chainlaunch/pkg/plugin/types"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"
	cmdCompose "github.com/docker/compose/v2/cmd/compose"
	"github.com/docker/compose/v2/pkg/api"

	"github.com/docker/compose/v2/pkg/compose"
	"gopkg.in/yaml.v3"
)

// PluginManager handles plugin operations
type PluginManager struct {
	pluginsDir string
	compose    api.Service
	dockerCli  *command.DockerCli
}

// NewPluginManager creates a new plugin manager
func NewPluginManager(pluginsDir string) (*PluginManager, error) {
	dockerCli, err := command.NewDockerCli()
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker CLI: %w", err)
	}

	err = dockerCli.Initialize(flags.NewClientOptions())
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Docker CLI: %w", err)
	}

	// check if docker engine is running
	_, err = dockerCli.Client().Info(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to check if Docker engine is running: %w", err)
	}

	// Create Compose service
	composeService := compose.NewComposeService(dockerCli)

	return &PluginManager{
		pluginsDir: pluginsDir,
		compose:    composeService,
		dockerCli:  dockerCli,
	}, nil
}

// LoadPlugin loads a plugin from a file
func (pm *PluginManager) LoadPlugin(filePath string) (*plugintypes.Plugin, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read plugin file: %w", err)
	}

	var plugin plugintypes.Plugin
	if err := yaml.Unmarshal(data, &plugin); err != nil {
		return nil, fmt.Errorf("failed to unmarshal plugin: %w", err)
	}

	return &plugin, nil
}

// DeployPlugin deploys a plugin using docker-compose
func (pm *PluginManager) DeployPlugin(ctx context.Context, plugin *plugintypes.Plugin, parameters map[string]interface{}) error {
	// Create a temporary directory for the plugin
	tempDir, err := os.MkdirTemp("", plugin.Metadata.Name)
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Write the docker-compose contents to a file
	composePath := filepath.Join(tempDir, "docker-compose.yml")
	if err := os.WriteFile(composePath, []byte(plugin.Spec.DockerCompose.Contents), 0644); err != nil {
		return fmt.Errorf("failed to write docker-compose file: %w", err)
	}

	// Create environment variables file
	envVars := make(map[string]string)
	for name, value := range parameters {
		if strValue, ok := value.(string); ok {
			envVars[name] = strValue
		}
	}

	envPath := filepath.Join(tempDir, ".env")
	envContent := ""
	for name, value := range envVars {
		envContent += fmt.Sprintf("%s=%s\n", name, value)
	}
	if err := os.WriteFile(envPath, []byte(envContent), 0644); err != nil {
		return fmt.Errorf("failed to write environment file: %w", err)
	}

	projectOptions := cmdCompose.ProjectOptions{
		ProjectName: plugin.Metadata.Name,
		ConfigPaths: []string{composePath},
	}

	// Turn projectOptions into a project with default values
	projectType, _, err := projectOptions.ToProject(ctx, pm.dockerCli, []string{})
	if err != nil {
		return err
	}

	upOptions := api.UpOptions{
		Create: api.CreateOptions{
			RemoveOrphans: true,
			QuietPull:     true,
		},
		Start: api.StartOptions{
			Wait: true,
		},
	}

	// Load the project
	err = pm.compose.Up(ctx, projectType, upOptions)
	if err != nil {
		return fmt.Errorf("failed to load project: %w", err)
	}

	return nil
}

// SavePlugin saves a plugin to a file
func (pm *PluginManager) SavePlugin(plugin *plugintypes.Plugin) error {
	data, err := yaml.Marshal(plugin)
	if err != nil {
		return fmt.Errorf("failed to marshal plugin: %w", err)
	}

	filePath := filepath.Join(pm.pluginsDir, plugin.Metadata.Name+".yaml")
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write plugin file: %w", err)
	}

	return nil
}

// ValidatePlugin validates a plugin definition
func (pm *PluginManager) ValidatePlugin(plugin *plugintypes.Plugin) error {
	if plugin.APIVersion == "" {
		return fmt.Errorf("apiVersion is required")
	}
	if plugin.Kind == "" {
		return fmt.Errorf("kind is required")
	}
	if plugin.Metadata.Name == "" {
		return fmt.Errorf("metadata.name is required")
	}
	if plugin.Metadata.Version == "" {
		return fmt.Errorf("metadata.version is required")
	}
	if plugin.Spec.DockerCompose.Contents == "" {
		return fmt.Errorf("spec.dockerCompose.contents is required")
	}
	if plugin.Spec.Parameters.Schema == "" {
		return fmt.Errorf("spec.parameters.$schema is required")
	}
	if plugin.Spec.Parameters.Type == "" {
		return fmt.Errorf("spec.parameters.type is required")
	}
	if len(plugin.Spec.Parameters.Properties) == 0 {
		return fmt.Errorf("spec.parameters.properties is required")
	}

	// Validate that all required parameters have corresponding properties
	for _, required := range plugin.Spec.Parameters.Required {
		if _, ok := plugin.Spec.Parameters.Properties[required]; !ok {
			return fmt.Errorf("required parameter %s is not defined in properties", required)
		}
	}

	return nil
}

// StopPlugin stops a running plugin deployment
func (pm *PluginManager) StopPlugin(ctx context.Context, plugin *plugintypes.Plugin) error {
	// Create a temporary directory for the plugin
	tempDir, err := os.MkdirTemp("", plugin.Metadata.Name)
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Write the docker-compose contents to a file
	composePath := filepath.Join(tempDir, "docker-compose.yml")
	if err := os.WriteFile(composePath, []byte(plugin.Spec.DockerCompose.Contents), 0644); err != nil {
		return fmt.Errorf("failed to write docker-compose file: %w", err)
	}

	projectOptions := cmdCompose.ProjectOptions{
		ProjectName: plugin.Metadata.Name,
		ConfigPaths: []string{composePath},
	}

	// Turn projectOptions into a project with default values
	projectType, _, err := projectOptions.ToProject(ctx, pm.dockerCli, []string{})
	if err != nil {
		return err
	}

	downOptions := api.DownOptions{
		RemoveOrphans: true,
		Volumes:       true,
	}

	// Stop the project
	err = pm.compose.Down(ctx, projectType.Name, downOptions)
	if err != nil {
		return fmt.Errorf("failed to stop project: %w", err)
	}

	return nil
}

// GetPluginStatus gets the current status of a plugin deployment
func (pm *PluginManager) GetPluginStatus(ctx context.Context, plugin *plugintypes.Plugin) (*plugintypes.DeploymentStatus, error) {
	// Create a temporary directory for the plugin
	tempDir, err := os.MkdirTemp("", plugin.Metadata.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Write the docker-compose contents to a file
	composePath := filepath.Join(tempDir, "docker-compose.yml")
	if err := os.WriteFile(composePath, []byte(plugin.Spec.DockerCompose.Contents), 0644); err != nil {
		return nil, fmt.Errorf("failed to write docker-compose file: %w", err)
	}

	projectOptions := cmdCompose.ProjectOptions{
		ProjectName: plugin.Metadata.Name,
		ConfigPaths: []string{composePath},
	}

	// Turn projectOptions into a project with default values
	projectType, _, err := projectOptions.ToProject(ctx, pm.dockerCli, []string{})
	if err != nil {
		return nil, err
	}

	// Get project status
	services, err := pm.compose.Ps(ctx, projectType.Name, api.PsOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get project status: %w", err)
	}

	status := &plugintypes.DeploymentStatus{
		Status:      "running",
		StartedAt:   time.Now(),
		ProjectName: plugin.Metadata.Name,
		Services:    make([]plugintypes.Service, len(services)),
	}

	for i, service := range services {
		// Get container details
		container, err := pm.dockerCli.Client().ContainerInspect(ctx, service.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to inspect container: %w", err)
		}

		ports := make([]plugintypes.Port, len(container.NetworkSettings.Ports))
		j := 0
		for port, bindings := range container.NetworkSettings.Ports {
			if len(bindings) > 0 {
				ports[j] = plugintypes.Port{
					HostPort:      bindings[0].HostPort,
					ContainerPort: port.Port(),
					Protocol:      port.Proto(),
				}
				j++
			}
		}
		ports = ports[:j]

		status.Services[i] = plugintypes.Service{
			Name:      service.Names[0][1:], // Remove leading slash
			Status:    service.State,
			Image:     service.Image,
			Ports:     ports,
			CreatedAt: time.Unix(service.Created, 0).Format(time.RFC3339),
		}
	}

	return status, nil
}
