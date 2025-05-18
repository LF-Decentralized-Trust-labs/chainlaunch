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

// CreateBesuNetwork creates a new Besu network using the API and returns the BesuNetworkResponse.
func (c *Client) CreateBesuNetwork(req *httptypes.CreateBesuNetworkRequest) (*httptypes.BesuNetworkResponse, error) {
	resp, err := c.Post("/networks/besu", req)
	if err != nil {
		return nil, fmt.Errorf("failed to create besu network: %w", err)
	}
	if err := CheckResponse(resp, 200, 201); err != nil {
		return nil, err
	}
	var netResp httptypes.BesuNetworkResponse
	body, err := ReadBody(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	if err := json.Unmarshal(body, &netResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &netResp, nil
}
