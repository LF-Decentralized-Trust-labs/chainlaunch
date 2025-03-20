package utils

import (
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/chainlaunch/chainlaunch/pkg/nodes/types"
)

// LoadNodeConfig deserializes a stored config based on its type
func LoadNodeConfig(data []byte) (types.NodeConfig, error) {
	var stored types.StoredConfig
	if err := json.Unmarshal(data, &stored); err != nil {
		return nil, fmt.Errorf("failed to unmarshal stored config: %w", err)
	}

	switch stored.Type {
	case "fabric-peer":
		var config types.FabricPeerConfig
		logrus.Debug("stored.Config", "stored.Config", string(stored.Config))
		if err := json.Unmarshal(stored.Config, &config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal fabric peer config: %w", err)
		}
		logrus.Debug("config", "config", config)
		return &config, nil

	case "fabric-orderer":
		var config types.FabricOrdererConfig
		if err := json.Unmarshal(stored.Config, &config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal fabric orderer config: %w", err)
		}
		return &config, nil

	case "besu":
		var config types.BesuNodeConfig
		if err := json.Unmarshal(stored.Config, &config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal besu config: %w", err)
		}
		return &config, nil

	default:
		return nil, fmt.Errorf("unsupported node type: %s", stored.Type)
	}
}

// Add this helper function to deserialize deployment config
func DeserializeDeploymentConfig(configJSON string) (types.NodeDeploymentConfig, error) {
	// First unmarshal to get the type
	var baseConfig struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal([]byte(configJSON), &baseConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal base config: %w", err)
	}

	// Based on the type, unmarshal into the appropriate struct
	var config types.NodeDeploymentConfig
	switch baseConfig.Type {
	case "fabric-peer":
		var c types.FabricPeerDeploymentConfig
		if err := json.Unmarshal([]byte(configJSON), &c); err != nil {
			return nil, fmt.Errorf("failed to unmarshal fabric peer config: %w", err)
		}
		config = &c
	case "fabric-orderer":
		var c types.FabricOrdererDeploymentConfig
		if err := json.Unmarshal([]byte(configJSON), &c); err != nil {
			return nil, fmt.Errorf("failed to unmarshal fabric orderer config: %w", err)
		}
		config = &c
	case "besu":
		var c types.BesuNodeDeploymentConfig
		if err := json.Unmarshal([]byte(configJSON), &c); err != nil {
			return nil, fmt.Errorf("failed to unmarshal besu config: %w", err)
		}
		config = &c
	default:
		return nil, fmt.Errorf("unknown node type: %s", baseConfig.Type)
	}

	return config, nil
}

// StoreNodeConfig serializes a node config with its type information
func StoreNodeConfig(config types.NodeConfig) ([]byte, error) {
	configBytes, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	stored := types.StoredConfig{
		Type:   config.GetType(),
		Config: configBytes,
	}

	return json.Marshal(stored)
}
