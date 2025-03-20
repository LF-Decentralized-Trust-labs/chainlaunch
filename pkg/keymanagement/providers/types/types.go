package types

import (
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

// GenerateKeyRequest represents the parameters for key generation
type GenerateKeyRequest struct {
	Name        string
	Description *string
	Algorithm   KeyAlgorithm
	KeySize     *int
	Curve       *ECCurve
	Format      string
	Status      string
	ExpiresAt   *time.Time
	ProviderID  *int
	UserID      int
	Metadata    string
	IsCA        *int
	Certificate *CertificateRequest // Optional: generate self-signed certificate
}

// StoreKeyRequest represents the parameters for storing a key
type StoreKeyRequest struct {
	Name              string
	Description       *string
	Algorithm         KeyAlgorithm
	KeySize           *int
	Curve             *string
	Format            string
	PublicKey         string
	PrivateKey        string
	Certificate       *string
	Status            string
	ExpiresAt         *time.Time
	SHA256Fingerprint string
	SHA1Fingerprint   string
	ProviderID        *int
	UserID            int
	Metadata          string
	EthereumAddress   *string
}

// RotateKeyRequest represents the parameters for key rotation
type RotateKeyRequest struct {
	ID          int
	NewKeySize  *int    // Optional: change key size during rotation
	NewCurve    *string // Optional: change curve during rotation
	Description *string // Optional: update description
}

// CertificateRequest represents the parameters for generating a certificate
type CertificateRequest struct {
	CommonName         string
	Organization       []string
	OrganizationalUnit []string
	Country            []string
	Province           []string
	Locality           []string
	StreetAddress      []string
	PostalCode         []string
	DNSNames           []string
	EmailAddresses     []string
	IPAddresses        []net.IP
	URIs               []*url.URL
	ValidFrom          time.Time
	ValidFor           time.Duration
	IsCA               bool
	KeyUsage           x509.KeyUsage
	ExtKeyUsage        []x509.ExtKeyUsage
}

// SignCertificateRequest represents the parameters for signing a certificate with an existing CA
type SignCertificateRequest struct {
	KeyID   int // ID of the key to sign
	CAKeyID int // ID of the CA key to sign with
	CertificateRequest
}
