package common

import (
	"encoding/json"
	"fmt"
	"net/http"

	httptypes "github.com/chainlaunch/chainlaunch/pkg/networks/http"
)

// CreateFabricNetwork creates a new Fabric network using the REST API
func (c *Client) CreateFabricNetwork(req *httptypes.CreateFabricNetworkRequest) (*httptypes.NetworkResponse, error) {
	resp, err := c.Post("/networks/fabric", req)
	if err != nil {
		return nil, fmt.Errorf("failed to create fabric network: %w", err)
	}
	defer resp.Body.Close()
	if err := CheckResponse(resp, http.StatusCreated); err != nil {
		return nil, err
	}
	var network httptypes.NetworkResponse
	if err := json.NewDecoder(resp.Body).Decode(&network); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &network, nil
}
