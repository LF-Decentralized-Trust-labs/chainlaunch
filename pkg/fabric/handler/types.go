package handler

import (
	"time"

	"github.com/chainlaunch/chainlaunch/pkg/fabric/service"
)

// HTTP layer request/response structs
type CreateOrganizationRequest struct {
	MspID       string `json:"mspId" validate:"required"`
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
	ProviderID  int64  `json:"providerId"`
}

type UpdateOrganizationRequest struct {
	Description *string `json:"description"`
}

// OrganizationResponse represents the HTTP response structure
type OrganizationResponse struct {
	ID              int64     `json:"id"`
	MspID           string    `json:"mspId"`
	Description     string    `json:"description,omitempty"`
	SignPublicKey   string    `json:"signPublicKey"`
	SignCertificate string    `json:"signCertificate"`
	TlsPublicKey    string    `json:"tlsPublicKey"`
	TlsCertificate  string    `json:"tlsCertificate"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
	ProviderID      int64     `json:"providerId"`
	ProviderName    string    `json:"providerName,omitempty"`
}

// Convert service DTO to HTTP response
func toOrganizationResponse(dto *service.OrganizationDTO) *OrganizationResponse {
	return &OrganizationResponse{
		ID:              dto.ID,
		MspID:           dto.MspID,
		Description:     dto.Description.String,
		SignPublicKey:   dto.SignPublicKey,
		SignCertificate: dto.SignCertificate,
		TlsPublicKey:    dto.TlsPublicKey,
		TlsCertificate:  dto.TlsCertificate,
		CreatedAt:       dto.CreatedAt,
		UpdatedAt:       dto.UpdatedAt,
		ProviderID:      dto.ProviderID,
		ProviderName:    dto.ProviderName,
	}
}
