package common

import (
	"encoding/json"
	"fmt"

	"github.com/chainlaunch/chainlaunch/pkg/keymanagement/models"
)

// CreateKey creates a new cryptographic key using the API and returns the KeyResponse.
func (c *Client) CreateKey(req *models.CreateKeyRequest) (*models.KeyResponse, error) {
	resp, err := c.Post("/keys", req)
	if err != nil {
		return nil, fmt.Errorf("failed to create key: %w", err)
	}
	if err := CheckResponse(resp, 201); err != nil {
		return nil, err
	}
	var keyResp models.KeyResponse
	body, err := ReadBody(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	if err := json.Unmarshal(body, &keyResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &keyResp, nil
}

// GetKey retrieves a cryptographic key by ID using the API and returns the KeyResponse.
func (c *Client) GetKey(keyID string) (*models.KeyResponse, error) {
	resp, err := c.Get(fmt.Sprintf("/keys/%s", keyID))
	if err != nil {
		return nil, fmt.Errorf("failed to get key: %w", err)
	}
	if err := CheckResponse(resp, 200); err != nil {
		return nil, err
	}
	var keyResp models.KeyResponse
	body, err := ReadBody(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	if err := json.Unmarshal(body, &keyResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &keyResp, nil
}
