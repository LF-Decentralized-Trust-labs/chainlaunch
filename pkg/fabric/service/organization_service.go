package service

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"database/sql"
	"encoding/asn1"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chainlaunch/chainlaunch/pkg/config"
	"github.com/chainlaunch/chainlaunch/pkg/db"
	"github.com/chainlaunch/chainlaunch/pkg/keymanagement/models"
	keymanagement "github.com/chainlaunch/chainlaunch/pkg/keymanagement/service"
	gwidentity "github.com/hyperledger/fabric-gateway/pkg/identity"
)

// OrganizationDTO represents the service layer data structure
type OrganizationDTO struct {
	ID              int64          `json:"id"`
	MspID           string         `json:"mspId"`
	Description     sql.NullString `json:"description"`
	SignKeyID       sql.NullInt64  `json:"signKeyId"`
	TlsRootKeyID    sql.NullInt64  `json:"tlsRootKeyId"`
	SignPublicKey   string         `json:"signPublicKey"`
	SignCertificate string         `json:"signCertificate"`
	TlsPublicKey    string         `json:"tlsPublicKey"`
	TlsCertificate  string         `json:"tlsCertificate"`
	CreatedAt       time.Time      `json:"createdAt"`
	UpdatedAt       time.Time      `json:"updatedAt"`
	AdminTlsKeyID   sql.NullInt64  `json:"adminTlsKeyId"`
	AdminSignKeyID  sql.NullInt64  `json:"adminSignKeyId"`
	ClientSignKeyID sql.NullInt64  `json:"clientSignKeyId"`
	ProviderID      int64          `json:"providerId"`
	ProviderName    string         `json:"providerName"`
}

// CreateOrganizationParams represents the service layer input parameters
type CreateOrganizationParams struct {
	MspID       string `validate:"required"`
	Name        string `validate:"required"`
	Description string
	ProviderID  int64
}

// UpdateOrganizationParams represents the service layer update parameters
type UpdateOrganizationParams struct {
	Description *string
}

type OrganizationService struct {
	queries       *db.Queries
	keyManagement *keymanagement.KeyManagementService
	configService *config.ConfigService
}

func NewOrganizationService(queries *db.Queries, keyManagement *keymanagement.KeyManagementService, configService *config.ConfigService) *OrganizationService {
	return &OrganizationService{
		queries:       queries,
		keyManagement: keyManagement,
		configService: configService,
	}
}

func mapDBOrganizationToServiceOrganization(org *db.GetFabricOrganizationByMspIDRow) *OrganizationDTO {
	providerName := ""
	if org.ProviderName.Valid {
		providerName = org.ProviderName.String
	}

	return &OrganizationDTO{
		ID:              org.ID,
		MspID:           org.MspID,
		Description:     org.Description,
		SignKeyID:       org.SignKeyID,
		TlsRootKeyID:    org.TlsRootKeyID,
		SignPublicKey:   org.SignPublicKey.String,
		SignCertificate: org.SignCertificate.String,
		TlsPublicKey:    org.TlsPublicKey.String,
		TlsCertificate:  org.TlsCertificate.String,
		CreatedAt:       org.CreatedAt,
		UpdatedAt:       org.UpdatedAt.Time,
		AdminTlsKeyID:   org.AdminTlsKeyID,
		AdminSignKeyID:  org.AdminSignKeyID,
		ClientSignKeyID: org.ClientSignKeyID,
		ProviderID:      org.ProviderID.Int64,
		ProviderName:    providerName,
	}
}

// Convert database model to DTO for single organization
func toOrganizationDTO(org *db.GetFabricOrganizationWithKeysRow) *OrganizationDTO {
	providerName := ""
	if org.ProviderName.Valid {
		providerName = org.ProviderName.String
	}

	return &OrganizationDTO{
		ID:              org.ID,
		MspID:           org.MspID,
		Description:     org.Description,
		SignKeyID:       org.SignKeyID,
		TlsRootKeyID:    org.TlsRootKeyID,
		SignPublicKey:   org.SignPublicKey.String,
		SignCertificate: org.SignCertificate.String,
		TlsPublicKey:    org.TlsPublicKey.String,
		TlsCertificate:  org.TlsCertificate.String,
		CreatedAt:       org.CreatedAt,
		UpdatedAt:       org.UpdatedAt.Time,
		AdminTlsKeyID:   org.AdminTlsKeyID,
		AdminSignKeyID:  org.AdminSignKeyID,
		ClientSignKeyID: org.ClientSignKeyID,
		ProviderID:      org.ProviderID.Int64,
		ProviderName:    providerName,
	}
}

// Convert database model to DTO for list of organizations
func toOrganizationListDTO(org *db.ListFabricOrganizationsWithKeysRow) *OrganizationDTO {
	providerName := ""
	if org.ProviderName.Valid {
		providerName = org.ProviderName.String
	}

	return &OrganizationDTO{
		ID:              org.ID,
		MspID:           org.MspID,
		Description:     org.Description,
		SignKeyID:       org.SignKeyID,
		TlsRootKeyID:    org.TlsRootKeyID,
		SignPublicKey:   org.SignPublicKey.String,
		SignCertificate: org.SignCertificate.String,
		TlsPublicKey:    org.TlsPublicKey.String,
		TlsCertificate:  org.TlsCertificate.String,
		CreatedAt:       org.CreatedAt,
		UpdatedAt:       org.UpdatedAt.Time,
		ProviderID:      org.ProviderID.Int64,
		ProviderName:    providerName,
	}
}

func (s *OrganizationService) CreateOrganization(ctx context.Context, params CreateOrganizationParams) (*OrganizationDTO, error) {
	description := fmt.Sprintf("Sign key for organization %s", params.MspID)
	curve := models.ECCurveP256
	// Create SIGN key
	providerID := int(params.ProviderID)
	isCA := 1
	signKeyReq := models.CreateKeyRequest{
		Name:        fmt.Sprintf("%s-sign-ca", params.MspID),
		Description: &description,
		Algorithm:   models.KeyAlgorithmEC,
		KeySize:     nil,
		Curve:       &curve,
		ProviderID:  &providerID,
		IsCA:        &isCA,
		Certificate: &models.CertificateRequest{
			CommonName:         fmt.Sprintf("%s-sign-ca", params.MspID),
			Organization:       []string{params.Name},
			OrganizationalUnit: []string{"SIGN"},
			Country:            []string{"US"},
			Locality:           []string{"San Francisco"},
			Province:           []string{"California"},
		},
	}
	signKey, err := s.keyManagement.CreateKey(ctx, signKeyReq, providerID)
	if err != nil {
		return nil, fmt.Errorf("failed to create SIGN key: %w", err)
	}

	// Create SIGN admin key
	isCA = 0
	signAdminKeyReq := models.CreateKeyRequest{
		Name:        fmt.Sprintf("%s-sign-admin", params.MspID),
		Description: &description,
		Algorithm:   models.KeyAlgorithmEC,
		KeySize:     nil,
		Curve:       &curve,
		ProviderID:  &providerID,
		IsCA:        &isCA,
	}
	signAdminKey, err := s.keyManagement.CreateKey(ctx, signAdminKeyReq, providerID)
	if err != nil {
		_ = s.keyManagement.DeleteKey(ctx, signKey.ID)
		return nil, fmt.Errorf("failed to create SIGN admin key: %w", err)
	}

	// Sign the admin key with the CA
	signAdminKey, err = s.keyManagement.SignCertificate(ctx, signAdminKey.ID, signKey.ID, models.CertificateRequest{
		CommonName:         fmt.Sprintf("%s-sign-admin", params.MspID),
		Organization:       []string{params.Name},
		OrganizationalUnit: []string{"admin"},
		Country:            []string{"US"},
		Locality:           []string{"San Francisco"},
		Province:           []string{"California"},
	})
	if err != nil {
		_ = s.keyManagement.DeleteKey(ctx, signKey.ID)
		_ = s.keyManagement.DeleteKey(ctx, signAdminKey.ID)
		return nil, fmt.Errorf("failed to sign admin certificate: %w", err)
	}

	// Create SIGN client key
	signClientKeyReq := models.CreateKeyRequest{
		Name:        fmt.Sprintf("%s-sign-client", params.MspID),
		Description: &description,
		Algorithm:   models.KeyAlgorithmEC,
		KeySize:     nil,
		Curve:       &curve,
		ProviderID:  &providerID,
		IsCA:        &isCA,
	}
	signClientKey, err := s.keyManagement.CreateKey(ctx, signClientKeyReq, providerID)
	if err != nil {
		_ = s.keyManagement.DeleteKey(ctx, signKey.ID)
		_ = s.keyManagement.DeleteKey(ctx, signAdminKey.ID)
		return nil, fmt.Errorf("failed to create SIGN client key: %w", err)
	}

	// Sign the client key with the CA
	signClientKey, err = s.keyManagement.SignCertificate(ctx, signClientKey.ID, signKey.ID, models.CertificateRequest{
		CommonName:         fmt.Sprintf("%s-sign-client", params.MspID),
		Organization:       []string{params.Name},
		OrganizationalUnit: []string{"client"},
		Country:            []string{"US"},
		Locality:           []string{"San Francisco"},
		Province:           []string{"California"},
	})
	if err != nil {
		_ = s.keyManagement.DeleteKey(ctx, signKey.ID)
		_ = s.keyManagement.DeleteKey(ctx, signAdminKey.ID)
		_ = s.keyManagement.DeleteKey(ctx, signClientKey.ID)
		return nil, fmt.Errorf("failed to sign client certificate: %w", err)
	}

	// Create TLS key
	isCA = 1
	tlsKeyReq := models.CreateKeyRequest{
		Name:        fmt.Sprintf("%s-tls-ca", params.MspID),
		Description: &description,
		Algorithm:   models.KeyAlgorithmEC,
		KeySize:     nil,
		Curve:       &curve,
		ProviderID:  &providerID,
		IsCA:        &isCA,
		Certificate: &models.CertificateRequest{
			CommonName:         fmt.Sprintf("%s-tls-ca", params.MspID),
			Organization:       []string{params.Name},
			OrganizationalUnit: []string{"TLS"},
			Country:            []string{"US"},
			Locality:           []string{"San Francisco"},
			Province:           []string{"California"},
		},
	}
	tlsKey, err := s.keyManagement.CreateKey(ctx, tlsKeyReq, providerID)
	if err != nil {
		_ = s.keyManagement.DeleteKey(ctx, signKey.ID)
		_ = s.keyManagement.DeleteKey(ctx, signAdminKey.ID)
		_ = s.keyManagement.DeleteKey(ctx, signClientKey.ID)
		return nil, fmt.Errorf("failed to create TLS key: %w", err)
	}

	// Create TLS admin key
	isCA = 0
	tlsAdminKeyReq := models.CreateKeyRequest{
		Name:        fmt.Sprintf("%s-tls-admin", params.MspID),
		Description: &description,
		Algorithm:   models.KeyAlgorithmEC,
		KeySize:     nil,
		Curve:       &curve,
		ProviderID:  &providerID,
		IsCA:        &isCA,
	}
	tlsAdminKey, err := s.keyManagement.CreateKey(ctx, tlsAdminKeyReq, providerID)
	if err != nil {
		_ = s.keyManagement.DeleteKey(ctx, signKey.ID)
		_ = s.keyManagement.DeleteKey(ctx, signAdminKey.ID)
		_ = s.keyManagement.DeleteKey(ctx, signClientKey.ID)
		_ = s.keyManagement.DeleteKey(ctx, tlsKey.ID)
		return nil, fmt.Errorf("failed to create TLS admin key: %w", err)
	}

	// Sign the TLS admin key with the CA
	tlsAdminKey, err = s.keyManagement.SignCertificate(ctx, tlsAdminKey.ID, tlsKey.ID, models.CertificateRequest{
		CommonName:         fmt.Sprintf("%s-tls-admin", params.MspID),
		Organization:       []string{params.Name},
		OrganizationalUnit: []string{"admin"},
		Country:            []string{"US"},
		Locality:           []string{"San Francisco"},
		Province:           []string{"California"},
	})
	if err != nil {
		_ = s.keyManagement.DeleteKey(ctx, signKey.ID)
		_ = s.keyManagement.DeleteKey(ctx, signAdminKey.ID)
		_ = s.keyManagement.DeleteKey(ctx, signClientKey.ID)
		_ = s.keyManagement.DeleteKey(ctx, tlsKey.ID)
		_ = s.keyManagement.DeleteKey(ctx, tlsAdminKey.ID)
		return nil, fmt.Errorf("failed to sign TLS admin certificate: %w", err)
	}

	// Create organization
	org, err := s.queries.CreateFabricOrganization(ctx, &db.CreateFabricOrganizationParams{
		MspID:           params.MspID,
		Description:     sql.NullString{String: params.Description, Valid: params.Description != ""},
		ProviderID:      sql.NullInt64{Int64: params.ProviderID, Valid: true},
		SignKeyID:       sql.NullInt64{Int64: int64(signKey.ID), Valid: true},
		TlsRootKeyID:    sql.NullInt64{Int64: int64(tlsKey.ID), Valid: true},
		AdminTlsKeyID:   sql.NullInt64{Int64: int64(tlsAdminKey.ID), Valid: true},
		AdminSignKeyID:  sql.NullInt64{Int64: int64(signAdminKey.ID), Valid: true},
		ClientSignKeyID: sql.NullInt64{Int64: int64(signClientKey.ID), Valid: true},
	})

	if err != nil {
		_ = s.keyManagement.DeleteKey(ctx, signKey.ID)
		_ = s.keyManagement.DeleteKey(ctx, signAdminKey.ID)
		_ = s.keyManagement.DeleteKey(ctx, signClientKey.ID)
		_ = s.keyManagement.DeleteKey(ctx, tlsKey.ID)
		_ = s.keyManagement.DeleteKey(ctx, tlsAdminKey.ID)
		return nil, fmt.Errorf("failed to create organization: %w", err)
	}

	// After creating the organization, fetch it with the provider name
	createdOrg, err := s.queries.GetFabricOrganizationWithKeys(ctx, org.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch created organization: %w", err)
	}

	return toOrganizationDTO(createdOrg), nil
}

func (s *OrganizationService) GetOrganization(ctx context.Context, id int64) (*OrganizationDTO, error) {
	org, err := s.queries.GetFabricOrganizationWithKeys(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("organization not found")
		}
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	return toOrganizationDTO(org), nil
}

func (s *OrganizationService) GetOrganizationByMspID(ctx context.Context, mspID string) (*OrganizationDTO, error) {
	org, err := s.queries.GetFabricOrganizationByMspID(ctx, mspID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("organization not found")
		}
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}
	return mapDBOrganizationToServiceOrganization(org), nil
}

func (s *OrganizationService) UpdateOrganization(ctx context.Context, id int64, req UpdateOrganizationParams) (*OrganizationDTO, error) {
	// Get existing organization
	org, err := s.queries.GetFabricOrganization(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("organization not found")
		}
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	// Update fields if provided
	if req.Description != nil {
		org.Description = sql.NullString{String: *req.Description, Valid: true}
	}

	// Update organization
	_, err = s.queries.UpdateFabricOrganization(ctx, &db.UpdateFabricOrganizationParams{
		ID:          id,
		Description: org.Description,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update organization: %w", err)
	}

	// Fetch the updated organization with keys
	updatedOrg, err := s.queries.GetFabricOrganizationWithKeys(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch updated organization: %w", err)
	}

	return toOrganizationDTO(updatedOrg), nil
}

func (s *OrganizationService) DeleteOrganization(ctx context.Context, id int64) error {
	// Get the organization first to retrieve the MspID
	org, err := s.queries.GetFabricOrganization(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("organization not found")
		}
		return fmt.Errorf("failed to get organization: %w", err)
	}

	// Delete the organization from the database
	err = s.queries.DeleteFabricOrganization(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("organization not found")
		}
		return fmt.Errorf("failed to delete organization: %w", err)
	}

	// Delete the organization directory
	// Convert MspID to lowercase for the directory name
	mspIDLower := strings.ToLower(org.MspID)

	orgDir := filepath.Join(s.configService.GetDataPath(), "orgs", mspIDLower)
	err = os.RemoveAll(orgDir)
	if err != nil {
		// Log the error but don't fail the operation
		// The database record is already deleted
		fmt.Printf("Warning: failed to delete organization directory %s: %v\n", orgDir, err)
	}

	return nil
}

func (s *OrganizationService) ListOrganizations(ctx context.Context) ([]OrganizationDTO, error) {
	orgs, err := s.queries.ListFabricOrganizationsWithKeys(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list organizations: %w", err)
	}

	dtos := make([]OrganizationDTO, len(orgs))
	for i, org := range orgs {
		dtos[i] = *toOrganizationListDTO(org)
	}
	return dtos, nil
}

// InitializeCRL initializes a new CRL for the organization
func (s *OrganizationService) InitializeCRL(ctx context.Context, orgID int64) error {
	// Get organization details
	org, err := s.queries.GetFabricOrganizationWithKeys(ctx, orgID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("organization not found")
		}
		return fmt.Errorf("failed to get organization: %w", err)
	}

	if !org.SignKeyID.Valid {
		return fmt.Errorf("organization has no admin sign key")
	}

	// Update the CRL timestamps in the organization
	now := time.Now()

	err = s.queries.UpdateOrganizationCRL(ctx, &db.UpdateOrganizationCRLParams{
		ID:            orgID,
		CrlLastUpdate: sql.NullTime{Time: now, Valid: true},
		CrlKeyID:      org.SignKeyID,
	})
	if err != nil {
		return fmt.Errorf("failed to initialize CRL: %w", err)
	}

	return nil
}

// RevokeCertificate adds a certificate to the organization's CRL
func (s *OrganizationService) RevokeCertificate(ctx context.Context, orgID int64, serialNumber *big.Int, reason int) error {
	// Get organization details
	org, err := s.queries.GetFabricOrganizationWithKeys(ctx, orgID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("organization not found")
		}
		return fmt.Errorf("failed to get organization: %w", err)
	}

	if !org.SignKeyID.Valid {
		return fmt.Errorf("organization has no admin sign key")
	}

	// Add the certificate to the database
	err = s.queries.AddRevokedCertificate(ctx, &db.AddRevokedCertificateParams{
		FabricOrganizationID: orgID,
		SerialNumber:         serialNumber.Text(16), // Store as hex string
		RevocationTime:       time.Now(),
		Reason:               int64(reason),
		IssuerCertificateID: sql.NullInt64{
			Int64: org.SignKeyID.Int64,
			Valid: true,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to add revoked certificate: %w", err)
	}

	// Update the CRL timestamps
	now := time.Now()

	err = s.queries.UpdateOrganizationCRL(ctx, &db.UpdateOrganizationCRLParams{
		ID:            orgID,
		CrlLastUpdate: sql.NullTime{Time: now, Valid: true},
		CrlKeyID:      org.AdminSignKeyID,
	})
	if err != nil {
		return fmt.Errorf("failed to update CRL timestamps: %w", err)
	}

	return nil
}

// GetCRL returns the current CRL for the organization in PEM format
func (s *OrganizationService) GetCRL(ctx context.Context, orgID int64) ([]byte, error) {
	// Get organization details
	org, err := s.queries.GetFabricOrganizationWithKeys(ctx, orgID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("organization not found")
		}
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	// Get all revoked certificates for this organization
	revokedCerts, err := s.queries.GetRevokedCertificates(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get revoked certificates: %w", err)
	}

	// Get the admin signing key for signing the CRL
	adminSignKey, err := s.keyManagement.GetKey(ctx, int(org.SignKeyID.Int64))
	if err != nil {
		return nil, fmt.Errorf("failed to get admin sign key: %w", err)
	}

	// Parse the certificate
	cert, err := gwidentity.CertificateFromPEM([]byte(*adminSignKey.Certificate))
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Get private key from key management service
	privateKeyPEM, err := s.keyManagement.GetDecryptedPrivateKey(int(org.SignKeyID.Int64))
	if err != nil {
		return nil, fmt.Errorf("failed to get private key: %w", err)
	}

	// Parse the private key
	priv, err := gwidentity.PrivateKeyFromPEM([]byte(privateKeyPEM))
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	// Cast private key to crypto.Signer
	signer, ok := priv.(crypto.Signer)
	if !ok {
		return nil, fmt.Errorf("private key does not implement crypto.Signer")
	}

	// Create CRL
	now := time.Now()
	crl := &x509.RevocationList{
		Number:     big.NewInt(1),
		ThisUpdate: now,
		NextUpdate: now.AddDate(0, 0, 7), // Valid for 7 days
	}

	// Add all revoked certificates
	for _, rc := range revokedCerts {
		serialNumber, ok := new(big.Int).SetString(rc.SerialNumber, 16)
		if !ok {
			return nil, fmt.Errorf("invalid serial number format: %s", rc.SerialNumber)
		}

		revokedCert := pkix.RevokedCertificate{
			SerialNumber:   serialNumber,
			RevocationTime: rc.RevocationTime,
			Extensions: []pkix.Extension{
				{
					Id:    asn1.ObjectIdentifier{2, 5, 29, 21}, // CRLReason OID
					Value: []byte{byte(rc.Reason)},
				},
			},
		}
		crl.RevokedCertificates = append(crl.RevokedCertificates, revokedCert)
	}

	// Create the CRL
	crlBytes, err := x509.CreateRevocationList(rand.Reader, crl, cert, signer)
	if err != nil {
		return nil, fmt.Errorf("failed to create CRL: %w", err)
	}

	// Encode the CRL in PEM format
	pemBlock := &pem.Block{
		Type:  "X509 CRL",
		Bytes: crlBytes,
	}

	return pem.EncodeToMemory(pemBlock), nil
}
