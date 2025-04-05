package networkconfig

import (
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

// LoadFromFile loads a network configuration from a YAML file
func LoadFromFile(path string) (*NetworkConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return LoadFromReader(file)
}

// LoadFromReader loads a network configuration from an io.Reader
func LoadFromReader(reader io.Reader) (*NetworkConfig, error) {
	var config NetworkConfig
	decoder := yaml.NewDecoder(reader)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}
	return &config, nil
}

// LoadFromBytes loads a network configuration from a byte slice
func LoadFromBytes(data []byte) (*NetworkConfig, error) {
	var config NetworkConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// SaveToFile saves a network configuration to a YAML file
func (c *NetworkConfig) SaveToFile(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// SaveToWriter saves a network configuration to an io.Writer
func (c *NetworkConfig) SaveToWriter(writer io.Writer) error {
	encoder := yaml.NewEncoder(writer)
	return encoder.Encode(c)
}

// SaveToBytes converts a network configuration to a byte slice
func (c *NetworkConfig) SaveToBytes() ([]byte, error) {
	return yaml.Marshal(c)
}
