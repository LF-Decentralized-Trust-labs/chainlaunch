package types

import (
	"encoding/json"
	"fmt"
)

// Plugin represents a plugin definition
type Plugin struct {
	APIVersion string         `json:"apiVersion" yaml:"apiVersion"`
	Kind       string         `json:"kind" yaml:"kind"`
	Metadata   PluginMetadata `json:"metadata" yaml:"metadata"`
	Spec       PluginSpec     `json:"spec" yaml:"spec"`
}

// PluginMetadata contains metadata about the plugin
type PluginMetadata struct {
	Name        string            `json:"name" yaml:"name"`
	Version     string            `json:"version" yaml:"version"`
	Description string            `json:"description" yaml:"description"`
	Author      string            `json:"author" yaml:"author"`
	License     string            `json:"license" yaml:"license"`
	Tags        []string          `json:"tags" yaml:"tags"`
	Labels      map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
}

// PluginSpec contains the plugin specification
type PluginSpec struct {
	Image      string                 `json:"image" yaml:"image"`
	Parameters map[string]interface{} `json:"parameters" yaml:"parameters"`
	Env        map[string]string      `json:"env,omitempty" yaml:"env,omitempty"`
	Volumes    []PluginVolume         `json:"volumes,omitempty" yaml:"volumes,omitempty"`
	Ports      []PluginPort           `json:"ports,omitempty" yaml:"ports,omitempty"`
	Resources  *PluginResources       `json:"resources,omitempty" yaml:"resources,omitempty"`
}

// PluginVolume represents a volume mount for the plugin
type PluginVolume struct {
	Name      string `json:"name" yaml:"name"`
	MountPath string `json:"mountPath" yaml:"mountPath"`
	HostPath  string `json:"hostPath,omitempty" yaml:"hostPath,omitempty"`
}

// PluginPort represents a port mapping for the plugin
type PluginPort struct {
	Name          string `json:"name" yaml:"name"`
	ContainerPort int    `json:"containerPort" yaml:"containerPort"`
	HostPort      int    `json:"hostPort,omitempty" yaml:"hostPort,omitempty"`
	Protocol      string `json:"protocol,omitempty" yaml:"protocol,omitempty"`
}

// PluginResources represents resource requirements for the plugin
type PluginResources struct {
	CPU    string `json:"cpu,omitempty" yaml:"cpu,omitempty"`
	Memory string `json:"memory,omitempty" yaml:"memory,omitempty"`
}

// GetPluginParameters returns the plugin parameters as a JSON string
func (p *Plugin) GetPluginParameters() (string, error) {
	data, err := json.Marshal(p.Spec.Parameters)
	if err != nil {
		return "", fmt.Errorf("failed to marshal parameters: %w", err)
	}
	return string(data), nil
}

// SetPluginParameters sets the plugin parameters from a JSON string
func (p *Plugin) SetPluginParameters(parameters string) error {
	if err := json.Unmarshal([]byte(parameters), &p.Spec.Parameters); err != nil {
		return fmt.Errorf("failed to unmarshal parameters: %w", err)
	}
	return nil
}

// GetPluginEnv returns the plugin environment variables
func (p *Plugin) GetPluginEnv() map[string]string {
	if p.Spec.Env == nil {
		return make(map[string]string)
	}
	return p.Spec.Env
}

// SetPluginEnv sets the plugin environment variables
func (p *Plugin) SetPluginEnv(env map[string]string) {
	p.Spec.Env = env
}

// AddPluginVolume adds a volume mount to the plugin
func (p *Plugin) AddPluginVolume(volume PluginVolume) {
	p.Spec.Volumes = append(p.Spec.Volumes, volume)
}

// AddPluginPort adds a port mapping to the plugin
func (p *Plugin) AddPluginPort(port PluginPort) {
	p.Spec.Ports = append(p.Spec.Ports, port)
}

// SetPluginResources sets the plugin resource requirements
func (p *Plugin) SetPluginResources(resources PluginResources) {
	p.Spec.Resources = &resources
}
