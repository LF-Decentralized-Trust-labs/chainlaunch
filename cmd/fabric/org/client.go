package org

import (
	"fmt"
	"os"

	"github.com/chainlaunch/chainlaunch/pkg/fabric/client"
	"github.com/chainlaunch/chainlaunch/pkg/fabric/handler"
	"github.com/chainlaunch/chainlaunch/pkg/logger"
)

const (
	defaultBaseURL = "http://localhost:8100/api/v1"
)

// ClientWrapper wraps the API client with additional functionality
type ClientWrapper struct {
	client *client.Client
	logger *logger.Logger
}

// NewClientWrapper creates a new client wrapper
func NewClientWrapper(logger *logger.Logger) *ClientWrapper {
	// Get base URL from environment variable
	baseURL := os.Getenv("CHAINLAUNCH_URL")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	// Get credentials from environment variables
	username := os.Getenv("CHAINLAUNCH_USER")
	password := os.Getenv("CHAINLAUNCH_PASSWORD")

	return &ClientWrapper{
		client: client.NewClient(baseURL, username, password),
		logger: logger,
	}
}

// CreateOrganization creates a new organization
func (cw *ClientWrapper) CreateOrganization(name, mspID string, providerID int64) error {
	req := handler.CreateOrganizationRequest{
		MspID:       mspID,
		Name:        name,
		Description: fmt.Sprintf("Organization %s", name),
		ProviderID:  providerID,
	}

	org, err := cw.client.CreateOrganization(req)
	if err != nil {
		return fmt.Errorf("failed to create organization: %w", err)
	}

	cw.logger.Infof("Created organization: %s (ID: %d, MSP ID: %s)", name, org.ID, org.MspID)
	return nil
}

// ListOrganizations lists all organizations
func (cw *ClientWrapper) ListOrganizations() (*client.PaginatedOrganizationsResponse, error) {
	orgs, err := cw.client.ListOrganizations()
	if err != nil {
		return nil, fmt.Errorf("failed to list organizations: %w", err)
	}

	return orgs, nil
}

// DeleteOrganization deletes an organization by MSP ID
func (cw *ClientWrapper) DeleteOrganization(mspID string) error {
	// First get the organization by MSP ID to get its ID
	org, err := cw.client.GetOrganizationByMspID(mspID)
	if err != nil {
		return fmt.Errorf("failed to get organization: %w", err)
	}

	if err := cw.client.DeleteOrganization(org.ID); err != nil {
		return fmt.Errorf("failed to delete organization: %w", err)
	}

	cw.logger.Infof("Deleted organization: %s (MSP ID: %s)", org.Description, org.MspID)
	return nil
}

// UpdateOrganization updates an organization
func (cw *ClientWrapper) UpdateOrganization(mspID, name, domain string) error {
	// First get the organization by MSP ID to get its ID
	org, err := cw.client.GetOrganizationByMspID(mspID)
	if err != nil {
		return fmt.Errorf("failed to get organization: %w", err)
	}

	req := handler.UpdateOrganizationRequest{
		Description: &name,
	}

	updatedOrg, err := cw.client.UpdateOrganization(org.ID, req)
	if err != nil {
		return fmt.Errorf("failed to update organization: %w", err)
	}

	cw.logger.Infof("Updated organization: %s (MSP ID: %s)", updatedOrg.Description, updatedOrg.MspID)
	return nil
}
