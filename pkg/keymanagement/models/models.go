package models

import (
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"time"

	"crypto/x509"
)

type KeyProviderType string
type KeyAlgorithm string
type ECCurve string

const (
	KeyProviderTypeDatabase KeyProviderType = "DATABASE"
	KeyProviderTypeVault    KeyProviderType = "VAULT"
	KeyProviderTypeHSM      KeyProviderType = "HSM"

	KeyAlgorithmRSA     KeyAlgorithm = "RSA"
	KeyAlgorithmEC      KeyAlgorithm = "EC"
	KeyAlgorithmED25519 KeyAlgorithm = "ED25519"

	ECCurveP256      ECCurve = "P-256"
	ECCurveP384      ECCurve = "P-384"
	ECCurveP521      ECCurve = "P-521"
	ECCurveSECP256K1 ECCurve = "secp256k1"
)

// KeyAlgorithm represents the supported key algorithms
// @Description Supported key algorithms
type CreateKeyRequest struct {
	// Name of the key
	// @Required
	Name string `json:"name" validate:"required" example:"my-key"`

	// Optional description
	Description *string `json:"description,omitempty" example:"Key for signing certificates"`

	// Key algorithm (RSA, EC, ED25519)
	// @Required
	Algorithm KeyAlgorithm `json:"algorithm" validate:"required,oneof=RSA EC ED25519" example:"RSA"`

	// Key size in bits (for RSA)
	KeySize *int `json:"keySize,omitempty" validate:"omitempty,min=2048,max=8192" example:"2048"`

	// Elliptic curve name (for EC keys)
	Curve *ECCurve `json:"curve,omitempty" example:"P-256"`

	// Optional provider ID
	ProviderID *int `json:"providerId,omitempty" example:"1"`

	// Whether this key is a CA
	IsCA *int `json:"isCA,omitempty" example:"0"`

	// Optional: configure CA certificate properties
	Certificate *CertificateRequest `json:"certificate,omitempty"`
}

func (r *CreateKeyRequest) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("name is required")
	}

	// Validate algorithm
	validAlgorithms := map[KeyAlgorithm]bool{
		KeyAlgorithmRSA:     true,
		KeyAlgorithmEC:      true,
		KeyAlgorithmED25519: true,
	}
	if !validAlgorithms[r.Algorithm] {
		return fmt.Errorf("invalid algorithm: must be one of RSA, EC, ED25519")
	}

	// Validate RSA key size
	if r.Algorithm == KeyAlgorithmRSA {
		if r.KeySize == nil {
			return fmt.Errorf("key size is required for RSA keys")
		}
		if *r.KeySize < 2048 || *r.KeySize > 8192 {
			return fmt.Errorf("RSA key size must be between 2048 and 8192 bits")
		}
	}

	// Validate EC curve
	if r.Algorithm == KeyAlgorithmEC {
		if r.Curve == nil {
			return fmt.Errorf("curve is required for EC keys")
		}
		validCurves := map[ECCurve]bool{
			ECCurveP256:      true,
			ECCurveP384:      true,
			ECCurveP521:      true,
			ECCurveSECP256K1: true,
		}
		if !validCurves[*r.Curve] {
			return fmt.Errorf("invalid curve: must be one of P-256, P-384, P-521, secp256k1")
		}
	}

	return nil
}

type KeyResponse struct {
	ID                int             `json:"id"`
	Name              string          `json:"name"`
	Description       *string         `json:"description,omitempty"`
	Algorithm         KeyAlgorithm    `json:"algorithm"`
	KeySize           *int            `json:"keySize,omitempty"`
	Curve             *ECCurve        `json:"curve,omitempty"`
	Format            string          `json:"format"`
	PublicKey         string          `json:"publicKey"`
	Certificate       *string         `json:"certificate,omitempty"`
	Status            string          `json:"status"`
	CreatedAt         time.Time       `json:"createdAt"`
	ExpiresAt         *time.Time      `json:"expiresAt,omitempty"`
	LastRotatedAt     *time.Time      `json:"lastRotatedAt,omitempty"`
	SHA256Fingerprint string          `json:"sha256Fingerprint"`
	SHA1Fingerprint   string          `json:"sha1Fingerprint"`
	Provider          KeyProviderInfo `json:"provider"`
	EthereumAddress   string          `json:"ethereumAddress"`
	SigningKeyID      *int            `json:"signingKeyID,omitempty"`
}

type KeyProviderInfo struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type CreateProviderRequest struct {
	Name      string          `json:"name" validate:"required"`
	Type      KeyProviderType `json:"type" validate:"required,oneof=DATABASE VAULT HSM"`
	Config    json.RawMessage `json:"config,omitempty"`
	IsDefault int             `json:"isDefault" validate:"required,oneof=0 1"`
}

func (r *CreateProviderRequest) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("name is required")
	}

	validTypes := map[KeyProviderType]bool{
		KeyProviderTypeDatabase: true,
		KeyProviderTypeVault:    true,
		KeyProviderTypeHSM:      true,
	}

	if !validTypes[r.Type] {
		return fmt.Errorf("invalid provider type: must be one of DATABASE, VAULT, HSM")
	}

	return nil
}

type ProviderResponse struct {
	ID        int             `json:"id"`
	Name      string          `json:"name"`
	Type      KeyProviderType `json:"type"`
	IsDefault int             `json:"isDefault"`
	Config    json.RawMessage `json:"config,omitempty"`
	CreatedAt time.Time       `json:"createdAt"`
}

type PaginatedResponse struct {
	Items      []KeyResponse `json:"items"`
	TotalItems int64         `json:"totalItems"`
	Page       int           `json:"page"`
	PageSize   int           `json:"pageSize"`
}

// Add CertificateRequest model
type CertificateRequest struct {
	CommonName         string             `json:"commonName" validate:"required"`
	Organization       []string           `json:"organization,omitempty"`
	OrganizationalUnit []string           `json:"organizationalUnit,omitempty"`
	Country            []string           `json:"country,omitempty"`
	Province           []string           `json:"province,omitempty"`
	Locality           []string           `json:"locality,omitempty"`
	StreetAddress      []string           `json:"streetAddress,omitempty"`
	PostalCode         []string           `json:"postalCode,omitempty"`
	DNSNames           []string           `json:"dnsNames,omitempty"`
	EmailAddresses     []string           `json:"emailAddresses,omitempty"`
	IPAddresses        []net.IP           `json:"ipAddresses,omitempty"`
	URIs               []*url.URL         `json:"uris,omitempty"`
	ValidFor           Duration           `json:"validFor" validate:"required"`
	IsCA               bool               `json:"isCA"`
	KeyUsage           x509.KeyUsage      `json:"keyUsage"`
	ExtKeyUsage        []x509.ExtKeyUsage `json:"extKeyUsage,omitempty"`
}

// Add Duration type for JSON marshaling
type Duration time.Duration

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case float64:
		*d = Duration(time.Duration(value))
		return nil
	case string:
		tmp, err := time.ParseDuration(value)
		if err != nil {
			return err
		}
		*d = Duration(tmp)
		return nil
	default:
		return fmt.Errorf("invalid duration")
	}
}
