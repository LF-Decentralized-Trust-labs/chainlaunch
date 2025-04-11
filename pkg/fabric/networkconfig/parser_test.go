package networkconfig

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadFromFile(t *testing.T) {
	// Create a temporary test file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test-config.yaml")

	// Write test YAML content
	testYAML := `
name: test-network
version: "1.0"
client:
  organization: Org1
organizations:
  Org1:
    mspid: Org1MSP
    cryptoPath: /tmp/crypto
    users: {}
    peers: []
    orderers: []
orderers: {}
peers: {}
certificateAuthorities: {}
channels: {}
`
	err := os.WriteFile(testFile, []byte(testYAML), 0644)
	assert.NoError(t, err)

	// Test loading from file
	config, err := LoadFromFile(testFile)
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "test-network", config.Name)
	assert.Equal(t, "1.0", config.Version)
	assert.Equal(t, "Org1", config.Client.Organization)
}

func TestLoadFromBytes(t *testing.T) {
	testYAML := `
name: test-network
version: "1.0"
client:
  organization: Org1
organizations:
  Org1:
    mspid: Org1MSP
    cryptoPath: /tmp/crypto
    users: {}
    peers: []
    orderers: []
orderers: {}
peers: {}
certificateAuthorities: {}
channels: {}
`

	config, err := LoadFromBytes([]byte(testYAML))
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "test-network", config.Name)
	assert.Equal(t, "1.0", config.Version)
	assert.Equal(t, "Org1", config.Client.Organization)
}

func TestSaveToFile(t *testing.T) {
	// Create a test configuration
	config := &NetworkConfig{
		Name:    "test-network",
		Version: "1.0",
		Client: ClientConfig{
			Organization: "Org1",
		},
		Organizations:          make(map[string]Organization),
		Orderers:               make(map[string]Orderer),
		Peers:                  make(map[string]Peer),
		CertificateAuthorities: make(map[string]CertificateAuthority),
		Channels:               make(map[string]Channel),
	}

	// Save to a temporary file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test-save.yaml")
	err := config.SaveToFile(testFile)
	assert.NoError(t, err)

	// Verify the file exists
	_, err = os.Stat(testFile)
	assert.NoError(t, err)

	// Load the saved file and verify contents
	loadedConfig, err := LoadFromFile(testFile)
	assert.NoError(t, err)
	assert.NotNil(t, loadedConfig)
	assert.Equal(t, config.Name, loadedConfig.Name)
	assert.Equal(t, config.Version, loadedConfig.Version)
	assert.Equal(t, config.Client.Organization, loadedConfig.Client.Organization)
}
