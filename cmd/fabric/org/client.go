package org

import (
	"fmt"
	"io"
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
func (cw *ClientWrapper) CreateOrganization(name, mspID, domain string) error {
	req := handler.CreateOrganizationRequest{
		MspID:       mspID,
		Name:        name,
		Description: fmt.Sprintf("Organization %s with domain %s", name, domain),
	}

	org, err := cw.client.CreateOrganization(req)
	if err != nil {
		return fmt.Errorf("failed to create organization: %w", err)
	}

	cw.logger.Infof("Created organization: %s (ID: %d, MSP ID: %s)", name, org.ID, org.MspID)
	return nil
}

// ListOrganizations lists all organizations
func (cw *ClientWrapper) ListOrganizations(out io.Writer) error {
	orgs, err := cw.client.ListOrganizations()
	if err != nil {
		return fmt.Errorf("failed to list organizations: %w", err)
	}

	if len(orgs) == 0 {
		fmt.Fprintln(out, "No organizations found")
		return nil
	}

	fmt.Fprintln(out, "Organizations:")
	for _, org := range orgs {
		fmt.Fprintf(out, "- %s (ID: %d, MSP ID: %s)\n",
			org.Description, org.ID, org.MspID)
	}

	return nil
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
