package common

import (
	"encoding/json"
	"fmt"
	"net/http"

	types "github.com/chainlaunch/chainlaunch/pkg/fabric/handler"
)

func (c *Client) CreateOrganization(req types.CreateOrganizationRequest) (*types.OrganizationResponse, error) {
	resp, err := c.Post("/organizations", req)
	if err != nil {
		return nil, fmt.Errorf("failed to create organization: %w", err)
	}
	defer resp.Body.Close()
	if err := CheckResponse(resp, http.StatusCreated); err != nil {
		return nil, err
	}
	var orgResp types.OrganizationResponse
	if err := json.NewDecoder(resp.Body).Decode(&orgResp); err != nil {
		return nil, fmt.Errorf("failed to decode organization response: %w", err)
	}
	return &orgResp, nil
}

func (c *Client) ListOrganizations() ([]types.OrganizationResponse, error) {
	resp, err := c.Get("/organizations")
	if err != nil {
		return nil, fmt.Errorf("failed to list organizations: %w", err)
	}
	defer resp.Body.Close()
	if err := CheckResponse(resp, http.StatusOK); err != nil {
		return nil, err
	}
	var orgs []types.OrganizationResponse
	if err := json.NewDecoder(resp.Body).Decode(&orgs); err != nil {
		return nil, fmt.Errorf("failed to decode organizations list: %w", err)
	}
	return orgs, nil
}
