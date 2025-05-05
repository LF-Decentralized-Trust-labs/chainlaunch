package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	// Import the types package that contains the shared types
	orgtypes "github.com/chainlaunch/chainlaunch/pkg/fabric/handler"
	networktypes "github.com/chainlaunch/chainlaunch/pkg/networks/http"
	// Or wherever the shared types are defined
)

// Client represents the API client
type Client struct {
	baseURL    string
	httpClient *http.Client
	username   string
	password   string
}
type Test = networktypes.AddNodeToNetworkRequest

// Use the shared types instead of local definitions
type Organization = orgtypes.OrganizationResponse
type CreateOrganizationRequest = orgtypes.CreateOrganizationRequest
type UpdateOrganizationRequest = orgtypes.UpdateOrganizationRequest

// NewClient creates a new API client
func NewClient(baseURL, username, password string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: time.Second * 30,
		},
		username: username,
		password: password,
	}
}

// doRequest performs an HTTP request and handles the response
func (c *Client) doRequest(method, path string, body interface{}) ([]byte, error) {
	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	url := fmt.Sprintf("%s%s", c.baseURL, path)
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(c.username, c.password)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// GetOrganization retrieves an organization by ID
func (c *Client) GetOrganization(id int64) (*Organization, error) {
	respBody, err := c.doRequest("GET", fmt.Sprintf("/organizations/%d", id), nil)
	if err != nil {
		return nil, err
	}

	var org Organization
	if err := json.Unmarshal(respBody, &org); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &org, nil
}

// GetOrganizationByMspID retrieves an organization by MspID
func (c *Client) GetOrganizationByMspID(mspID string) (*Organization, error) {
	respBody, err := c.doRequest("GET", fmt.Sprintf("/organizations/by-mspid/%s", mspID), nil)
	if err != nil {
		return nil, err
	}

	var org Organization
	if err := json.Unmarshal(respBody, &org); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &org, nil
}

// CreateOrganization creates a new organization
func (c *Client) CreateOrganization(req CreateOrganizationRequest) (*Organization, error) {
	respBody, err := c.doRequest("POST", "/organizations", req)
	if err != nil {
		return nil, err
	}

	var org Organization
	if err := json.Unmarshal(respBody, &org); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &org, nil
}

func (c *Client) GetNetworkByName(name string) (*networktypes.NetworkResponse, error) {
	respBody, err := c.doRequest("GET", fmt.Sprintf("/networks/fabric/by-name/%s", name), nil)
	if err != nil {
		return nil, err
	}

	var network networktypes.NetworkResponse
	if err := json.Unmarshal(respBody, &network); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return &network, nil
}

// UpdateOrganization updates an existing organization
func (c *Client) UpdateOrganization(id int64, req UpdateOrganizationRequest) (*Organization, error) {
	respBody, err := c.doRequest("PUT", fmt.Sprintf("/organizations/%d", id), req)
	if err != nil {
		return nil, err
	}

	var org Organization
	if err := json.Unmarshal(respBody, &org); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &org, nil
}

// DeleteOrganization deletes an organization
func (c *Client) DeleteOrganization(id int64) error {
	_, err := c.doRequest("DELETE", fmt.Sprintf("/organizations/%d", id), nil)
	return err
}

// ListOrganizations retrieves all organizations
func (c *Client) ListOrganizations() ([]Organization, error) {
	respBody, err := c.doRequest("GET", "/organizations", nil)
	if err != nil {
		return nil, err
	}

	var orgs []Organization
	if err := json.Unmarshal(respBody, &orgs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return orgs, nil
}

func (c *Client) GetNetworkConfig(networkID int64, organizationID int64) ([]byte, error) {
	respBody, err := c.doRequest("GET", fmt.Sprintf("/networks/fabric/%d/organizations/%d/network-config", networkID, organizationID), nil)
	if err != nil {
		return nil, err
	}
	return respBody, nil
}

