package database

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"os"
	"time"

	"github.com/chainlaunch/chainlaunch/pkg/db"
	"github.com/chainlaunch/chainlaunch/pkg/keymanagement/models"
	"github.com/chainlaunch/chainlaunch/pkg/keymanagement/providers/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sirupsen/logrus"
)

type DatabaseProvider struct {
	queries       *db.Queries
	encryptionKey []byte
}

type encryptedData struct {
	IV      string `json:"iv"`
	Data    string `json:"data"`
	AuthTag string `json:"authTag"`
}

func NewDatabaseProvider(queries *db.Queries) (*DatabaseProvider, error) {
	// Get encryption key from environment
	keyStr := os.Getenv("KEY_ENCRYPTION_KEY")
	if keyStr == "" {
		return nil, fmt.Errorf("KEY_ENCRYPTION_KEY environment variable not set")
	}

	key, err := hex.DecodeString(keyStr)
	if err != nil {
		return nil, fmt.Errorf("invalid encryption key: %w", err)
	}

	return &DatabaseProvider{
		queries:       queries,
		encryptionKey: key,
	}, nil
}
func (p *DatabaseProvider) GetDecryptedPrivateKey(id int) (string, error) {
	key, err := p.queries.GetKey(context.Background(), int64(id))
	if err != nil {
		return "", err
	}
	return p.decrypt(key.PrivateKey)
}

func (p *DatabaseProvider) encrypt(plaintext string) (string, error) {
	// Create new AES cipher
	block, err := aes.NewCipher(p.encryptionKey)
	if err != nil {
		return "", err
	}

	// Create GCM mode
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Create nonce
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// Encrypt
	ciphertext := aesGCM.Seal(nil, nonce, []byte(plaintext), nil)

	// Create encrypted data structure
	data := encryptedData{
		IV:      base64.StdEncoding.EncodeToString(nonce),
		Data:    base64.StdEncoding.EncodeToString(ciphertext[:len(ciphertext)-16]),
		AuthTag: base64.StdEncoding.EncodeToString(ciphertext[len(ciphertext)-16:]),
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

func (p *DatabaseProvider) decrypt(encryptedStr string) (string, error) {
	var data encryptedData
	if err := json.Unmarshal([]byte(encryptedStr), &data); err != nil {
		return "", err
	}

	// Decode components
	nonce, err := base64.StdEncoding.DecodeString(data.IV)
	if err != nil {
		return "", err
	}

	ciphertext, err := base64.StdEncoding.DecodeString(data.Data)
	if err != nil {
		return "", err
	}

	authTag, err := base64.StdEncoding.DecodeString(data.AuthTag)
	if err != nil {
		return "", err
	}

	// Combine ciphertext and auth tag
	fullCiphertext := append(ciphertext, authTag...)

	// Create new AES cipher
	block, err := aes.NewCipher(p.encryptionKey)
	if err != nil {
		return "", err
	}

	// Create GCM mode
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Decrypt
	plaintext, err := aesGCM.Open(nil, nonce, fullCiphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

func (p *DatabaseProvider) GenerateKey(ctx context.Context, req types.GenerateKeyRequest) (*models.KeyResponse, error) {
	// Convert provider request to internal request
	createReq := models.CreateKeyRequest{
		Name:        req.Name,
		Description: req.Description,
		Algorithm:   models.KeyAlgorithm(req.Algorithm),
		KeySize:     req.KeySize,
		ProviderID:  req.ProviderID,
		IsCA:        req.IsCA,
	}
	if req.Curve != nil {
		createReq.Curve = (*models.ECCurve)(req.Curve)
	}

	// Generate key pair
	keyPair, err := p.generateKeyPair(createReq)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key pair: %w", err)
	}

	// Create key in database
	params := db.CreateKeyParams{
		Name:              req.Name,
		Algorithm:         string(req.Algorithm),
		PublicKey:         keyPair.PublicKey,
		Format:            "PEM",
		Status:            req.Status,
		Sha256Fingerprint: keyPair.SHA256Fingerprint,
		Sha1Fingerprint:   keyPair.SHA1Fingerprint,
		ProviderID:        int64(*req.ProviderID),
		UserID:            int64(req.UserID),
		EthereumAddress:   sql.NullString{String: keyPair.EthereumAddress, Valid: keyPair.EthereumAddress != ""},
	}
	if req.IsCA != nil && *req.IsCA == 1 {
		params.IsCa = int64(1)
	}

	if req.Description != nil {
		params.Description = sql.NullString{String: *req.Description, Valid: true}
	}

	// Only set keySize for RSA algorithm
	if req.Algorithm == types.KeyAlgorithmRSA {
		params.KeySize = sql.NullInt64{Int64: int64(*req.KeySize), Valid: req.KeySize != nil}
	}

	// Only set curve for EC algorithm
	if req.Algorithm == types.KeyAlgorithmEC {
		params.Curve = sql.NullString{String: string(*req.Curve), Valid: req.Curve != nil}
	}

	// Encrypt private key before storing
	encryptedPrivateKey, err := p.encrypt(keyPair.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt private key: %w", err)
	}
	params.PrivateKey = encryptedPrivateKey

	// Generate self-signed certificate if requested or if isCA is 1
	if req.Certificate != nil || (req.IsCA != nil && *req.IsCA == 1) {
		var certReq *types.CertificateRequest
		if req.Certificate != nil {
			// Use provided certificate configuration
			certReq = req.Certificate
			// Ensure IsCA is set if this is a CA key
			if req.IsCA != nil && *req.IsCA == 1 {
				certReq.IsCA = true
				certReq.KeyUsage |= x509.KeyUsageCertSign | x509.KeyUsageCRLSign
			}
		} else {
			// Use default CA certificate configuration
			certReq = &types.CertificateRequest{
				CommonName:   req.Name,
				Organization: []string{"ChainDeploy"},
				Country:      []string{"US"},
				ValidFrom:    time.Now(),
				ValidFor:     time.Hour * 24 * 365 * 10, // 10 years
				IsCA:         true,
				KeyUsage:     x509.KeyUsageCertSign | x509.KeyUsageCRLSign | x509.KeyUsageDigitalSignature,
				ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
			}
		}

		cert, err := p.generateSelfSignedCert(keyPair, certReq)
		if err != nil {
			return nil, fmt.Errorf("failed to generate certificate: %w", err)
		}
		logrus.Debugf("Generated certificate: %s", cert)
		params.Certificate = sql.NullString{String: cert, Valid: true}
	}

	key, err := p.queries.CreateKey(ctx, params)
	if err != nil {
		return nil, err
	}

	return mapDBKeyToResponse(key), nil
}

func (p *DatabaseProvider) StoreKey(ctx context.Context, req types.StoreKeyRequest) (*models.KeyResponse, error) {
	// Encrypt private key before storage
	encryptedPrivateKey, err := p.encrypt(req.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt private key: %w", err)
	}

	// Store in database
	key, err := p.queries.CreateKey(ctx, db.CreateKeyParams{
		Name:              req.Name,
		Description:       sql.NullString{String: *req.Description, Valid: req.Description != nil},
		Algorithm:         string(req.Algorithm),
		KeySize:           sql.NullInt64{Int64: int64(*req.KeySize), Valid: req.KeySize != nil},
		Curve:             sql.NullString{String: string(*req.Curve), Valid: req.Curve != nil},
		Format:            req.Format,
		PublicKey:         req.PublicKey,
		PrivateKey:        encryptedPrivateKey,
		Certificate:       sql.NullString{String: *req.Certificate, Valid: req.Certificate != nil},
		Status:            req.Status,
		ExpiresAt:         sql.NullTime{Time: *req.ExpiresAt, Valid: req.ExpiresAt != nil},
		Sha256Fingerprint: req.SHA256Fingerprint,
		Sha1Fingerprint:   req.SHA1Fingerprint,
		ProviderID:        int64(*req.ProviderID),
		UserID:            int64(req.UserID),
		EthereumAddress:   sql.NullString{String: *req.EthereumAddress, Valid: req.EthereumAddress != nil},
	})
	if err != nil {
		return nil, err
	}

	return mapDBKeyToResponse(key), nil
}

func (p *DatabaseProvider) RetrieveKey(ctx context.Context, id int) (*models.KeyResponse, error) {
	key, err := p.queries.GetKey(ctx, int64(id))
	if err != nil {
		return nil, err
	}

	response := &models.KeyResponse{
		ID:                int(key.ID),
		Name:              key.Name,
		Description:       &key.Description.String,
		Algorithm:         models.KeyAlgorithm(key.Algorithm),
		Format:            key.Format,
		PublicKey:         key.PublicKey,
		Certificate:       &key.Certificate.String,
		Status:            key.Status,
		CreatedAt:         key.CreatedAt,
		ExpiresAt:         &key.ExpiresAt.Time,
		LastRotatedAt:     &key.LastRotatedAt.Time,
		SHA256Fingerprint: key.Sha256Fingerprint,
		SHA1Fingerprint:   key.Sha1Fingerprint,
		Provider:          models.KeyProviderInfo{ID: int(key.ProviderID), Name: "Database"},
		EthereumAddress:   key.EthereumAddress.String,
	}

	// Only include keySize for RSA algorithm
	if models.KeyAlgorithm(key.Algorithm) == models.KeyAlgorithmRSA && key.KeySize.Valid {
		keySize := int(key.KeySize.Int64)
		response.KeySize = &keySize
	}

	// Only include curve for EC algorithm
	if models.KeyAlgorithm(key.Algorithm) == models.KeyAlgorithmEC && key.Curve.Valid {
		curve := models.ECCurve(key.Curve.String)
		response.Curve = &curve
	}

	return response, nil
}

func (p *DatabaseProvider) DeleteKey(ctx context.Context, id int) error {
	return p.queries.DeleteKey(ctx, int64(id))
}

// Helper function to map database key to response
func mapDBKeyToResponse(key db.Key) *models.KeyResponse {
	response := &models.KeyResponse{
		ID:                int(key.ID),
		Name:              key.Name,
		Description:       &key.Description.String,
		Algorithm:         models.KeyAlgorithm(key.Algorithm),
		Format:            key.Format,
		PublicKey:         key.PublicKey,
		Certificate:       &key.Certificate.String,
		Status:            key.Status,
		CreatedAt:         key.CreatedAt,
		ExpiresAt:         &key.ExpiresAt.Time,
		LastRotatedAt:     &key.LastRotatedAt.Time,
		SHA256Fingerprint: key.Sha256Fingerprint,
		SHA1Fingerprint:   key.Sha1Fingerprint,
		Provider:          models.KeyProviderInfo{ID: int(key.ProviderID), Name: "Database"},
		EthereumAddress:   key.EthereumAddress.String,
	}

	// Only include keySize for RSA algorithm
	if models.KeyAlgorithm(key.Algorithm) == models.KeyAlgorithmRSA && key.KeySize.Valid {
		keySize := int(key.KeySize.Int64)
		response.KeySize = &keySize
	}

	// Only include curve for EC algorithm
	if models.KeyAlgorithm(key.Algorithm) == models.KeyAlgorithmEC && key.Curve.Valid {
		curve := models.ECCurve(key.Curve.String)
		response.Curve = &curve
	}

	return response
}

// KeyPair represents a public/private key pair
type KeyPair struct {
	PublicKey         string
	PrivateKey        string
	SHA256Fingerprint string
	SHA1Fingerprint   string
	EthereumAddress   string
}

func (s *DatabaseProvider) generateKeyPair(req models.CreateKeyRequest) (*KeyPair, error) {
	var keyPair *KeyPair
	var err error

	switch req.Algorithm {
	case models.KeyAlgorithmRSA:
		keyPair, err = s.generateRSAKeyPair(req)
	case models.KeyAlgorithmEC:
		keyPair, err = s.generateECKeyPair(req)
	case models.KeyAlgorithmED25519:
		keyPair, err = s.generateED25519KeyPair()
	default:
		return nil, fmt.Errorf("unsupported algorithm: %s", req.Algorithm)
	}

	if err != nil {
		return nil, err
	}
	// Calculate fingerprints from public key
	var publicKeyBytes []byte
	if req.Curve != nil && *req.Curve == models.ECCurveSECP256K1 {
		// For secp256k1, public key is already hex encoded
		var err error
		publicKeyBytes, err = hex.DecodeString(keyPair.PublicKey)
		if err != nil {
			return nil, fmt.Errorf("failed to decode hex public key: %w", err)
		}
	} else {
		// For other curves, public key is PEM encoded
		block, _ := pem.Decode([]byte(keyPair.PublicKey))
		if block == nil {
			return nil, fmt.Errorf("failed to decode public key PEM")
		}
		publicKeyBytes = block.Bytes
	}

	sha256Sum := sha256.Sum256(publicKeyBytes)
	sha1Sum := sha1.Sum(publicKeyBytes)

	keyPair.SHA256Fingerprint = hex.EncodeToString(sha256Sum[:])
	keyPair.SHA1Fingerprint = hex.EncodeToString(sha1Sum[:])

	return keyPair, nil
}

func (s *DatabaseProvider) generateRSAKeyPair(req models.CreateKeyRequest) (*KeyPair, error) {
	keySize := 2048
	if req.KeySize != nil {
		keySize = *req.KeySize
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key: %w", err)
	}

	// Encode private key
	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal private key: %w", err)
	}
	privatePEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	// Encode public key
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public key: %w", err)
	}
	publicPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	return &KeyPair{
		PublicKey:  string(publicPEM),
		PrivateKey: string(privatePEM),
	}, nil
}

func (s *DatabaseProvider) generateECKeyPair(req models.CreateKeyRequest) (*KeyPair, error) {
	if req.Curve == nil {
		return nil, fmt.Errorf("curve must be specified for EC keys")
	}

	var privateKeyBytes []byte
	var err error
	var privateKey *ecdsa.PrivateKey
	switch *req.Curve {
	case "P-224":
		privateKey, err = ecdsa.GenerateKey(elliptic.P224(), rand.Reader)
		if err != nil {
			return nil, fmt.Errorf("failed to generate P-224 key: %w", err)
		}
		privateKeyBytes, err = x509.MarshalPKCS8PrivateKey(privateKey)
	case "P-256":
		privateKey, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return nil, fmt.Errorf("failed to generate P-256 key: %w", err)
		}
		privateKeyBytes, err = x509.MarshalPKCS8PrivateKey(privateKey)
	case "P-384":
		privateKey, err = ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
		if err != nil {
			return nil, fmt.Errorf("failed to generate P-384 key: %w", err)
		}
		privateKeyBytes, err = x509.MarshalPKCS8PrivateKey(privateKey)
	case "P-521":
		privateKey, err = ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
		if err != nil {
			return nil, fmt.Errorf("failed to generate P-521 key: %w", err)
		}
		privateKeyBytes, err = x509.MarshalPKCS8PrivateKey(privateKey)
	case "secp256k1":
		privateKey, err = crypto.GenerateKey()
		if err != nil {
			return nil, fmt.Errorf("failed to generate secp256k1 key: %w", err)
		}
		publicKey := privateKey.Public().(*ecdsa.PublicKey)
		// Handle public key separately for secp256k1
		publicKeyBytes := crypto.FromECDSAPub(publicKey)
		publicKeyHex := hex.EncodeToString(publicKeyBytes)

		privateKeyBytes := crypto.FromECDSA(privateKey)
		privateKeyHex := hex.EncodeToString(privateKeyBytes)

		var ethereumAddress string
		if *req.Curve == "secp256k1" {
			// Generate Ethereum address from public key
			publicKeyECDSA, ok := privateKey.Public().(*ecdsa.PublicKey)
			if !ok {
				return nil, fmt.Errorf("error casting public key to ECDSA")
			}

			// Generate Ethereum address
			address := crypto.PubkeyToAddress(*publicKeyECDSA)
			ethereumAddress = address.Hex()
		}

		return &KeyPair{
			PublicKey:       publicKeyHex,
			PrivateKey:      privateKeyHex,
			EthereumAddress: ethereumAddress,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported curve: %s", *req.Curve)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to marshal private key: %w", err)
	}

	// Encode private key
	privatePEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	// Encode public key
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&ecdsa.PublicKey{Curve: elliptic.P256(), X: privateKey.X, Y: privateKey.Y})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public key: %w", err)
	}
	publicPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	var ethereumAddress string
	if *req.Curve == "secp256k1" {
		// Generate Ethereum address from public key
		publicKeyECDSA, ok := privateKey.Public().(*ecdsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("error casting public key to ECDSA")
		}

		// Generate Ethereum address
		address := crypto.PubkeyToAddress(*publicKeyECDSA)
		ethereumAddress = address.Hex()
	}

	return &KeyPair{
		PublicKey:       string(publicPEM),
		PrivateKey:      string(privatePEM),
		EthereumAddress: ethereumAddress,
	}, nil
}

func (s *DatabaseProvider) generateED25519KeyPair() (*KeyPair, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ED25519 key: %w", err)
	}

	// Encode private key
	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal private key: %w", err)
	}
	privatePEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	// Encode public key
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public key: %w", err)
	}
	publicPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	return &KeyPair{
		PublicKey:  string(publicPEM),
		PrivateKey: string(privatePEM),
	}, nil
}

func (p *DatabaseProvider) generateSelfSignedCert(keyPair *KeyPair, req *types.CertificateRequest) (string, error) {
	// Decode private key
	block, _ := pem.Decode([]byte(keyPair.PrivateKey))
	if block == nil {
		return "", fmt.Errorf("failed to decode private key")
	}

	privKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse private key: %w", err)
	}

	// Decode public key
	block, _ = pem.Decode([]byte(keyPair.PublicKey))
	if block == nil {
		return "", fmt.Errorf("failed to decode public key")
	}

	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse public key: %w", err)
	}

	// Generate serial number
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return "", fmt.Errorf("failed to generate serial number: %w", err)
	}

	// Create certificate template
	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:         req.CommonName,
			Organization:       req.Organization,
			OrganizationalUnit: req.OrganizationalUnit,
			Country:            req.Country,
			Province:           req.Province,
			Locality:           req.Locality,
			StreetAddress:      req.StreetAddress,
			PostalCode:         req.PostalCode,
		},
		NotBefore:             req.ValidFrom.Add(-time.Minute * 1),
		NotAfter:              req.ValidFrom.Add(time.Hour * 24 * 365),
		KeyUsage:              req.KeyUsage,
		ExtKeyUsage:           req.ExtKeyUsage,
		BasicConstraintsValid: true,
		IsCA:                  req.IsCA,
		DNSNames:              req.DNSNames,
		EmailAddresses:        req.EmailAddresses,
		IPAddresses:           req.IPAddresses,
		URIs:                  req.URIs,
	}

	// For self-signed certificates, the template is both the template and parent
	certBytes, err := x509.CreateCertificate(rand.Reader, template, template, pubKey, privKey)
	if err != nil {
		return "", fmt.Errorf("failed to create certificate: %w", err)
	}

	// Encode certificate to PEM
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	return string(certPEM), nil
}

func (p *DatabaseProvider) SignCertificate(ctx context.Context, req types.SignCertificateRequest) (*models.KeyResponse, error) {
	// Get the key to sign
	key, err := p.queries.GetKey(ctx, int64(req.KeyID))
	if err != nil {
		return nil, fmt.Errorf("failed to get key: %w", err)
	}

	// Get the CA key
	caKey, err := p.queries.GetKey(ctx, int64(req.CAKeyID))
	if err != nil {
		return nil, fmt.Errorf("failed to get CA key: %w", err)
	}

	// Verify CA key has CA certificate
	if !caKey.Certificate.Valid {
		return nil, fmt.Errorf("CA key %d has no certificate", req.CAKeyID)
	}

	// Decode CA private key
	decryptedCAKey, err := p.decrypt(caKey.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt CA private key: %w", err)
	}

	block, _ := pem.Decode([]byte(decryptedCAKey))
	if block == nil {
		return nil, fmt.Errorf("failed to decode CA private key")
	}

	caPrivKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CA private key: %w", err)
	}

	// Decode CA certificate
	block, _ = pem.Decode([]byte(caKey.Certificate.String))
	if block == nil {
		return nil, fmt.Errorf("failed to decode CA certificate")
	}

	logrus.Debugf("CA certificate: %+v", caKey.Certificate.String)

	caCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CA certificate: %w", err)
	}

	// Verify CA certificate is actually a CA
	if !caCert.IsCA {
		return nil, fmt.Errorf("key %d's certificate is not a CA certificate, value: %t", req.CAKeyID, caCert.IsCA)
	}

	// Decode public key to sign
	block, _ = pem.Decode([]byte(key.PublicKey))
	if block == nil {
		return nil, fmt.Errorf("failed to decode public key")
	}

	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	// Generate serial number
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number: %w", err)
	}

	// Calculate validity period
	// validUntil := req.ValidFrom.Add(time.Duration(req.ValidFor))
	validUntil := req.ValidFrom.Add(time.Hour * 24 * 365)

	// Create certificate template
	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:         req.CommonName,
			Organization:       req.Organization,
			OrganizationalUnit: req.OrganizationalUnit,
			Country:            req.Country,
			Province:           req.Province,
			Locality:           req.Locality,
			StreetAddress:      req.StreetAddress,
			PostalCode:         req.PostalCode,
		},
		NotBefore:             req.ValidFrom,
		NotAfter:              validUntil,
		KeyUsage:              req.KeyUsage,
		ExtKeyUsage:           req.ExtKeyUsage,
		BasicConstraintsValid: true,
		IsCA:                  false,
		DNSNames:              req.DNSNames,
		EmailAddresses:        req.EmailAddresses,
		IPAddresses:           req.IPAddresses,
		URIs:                  req.URIs,
	}

	// Create certificate using CA
	certBytes, err := x509.CreateCertificate(rand.Reader, template, caCert, pubKey, caPrivKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	// Encode certificate to PEM
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	// Update key with new certificate and other fields
	params := db.UpdateKeyParams{
		ID:                int64(req.KeyID),
		Name:              key.Name,
		Description:       key.Description,
		Algorithm:         key.Algorithm,
		KeySize:           key.KeySize,
		Curve:             key.Curve,
		Format:            key.Format,
		PublicKey:         key.PublicKey,
		PrivateKey:        key.PrivateKey,
		Certificate:       sql.NullString{String: string(certPEM), Valid: true},
		Status:            key.Status,
		ExpiresAt:         sql.NullTime{Time: validUntil, Valid: true},
		Sha256Fingerprint: key.Sha256Fingerprint,
		Sha1Fingerprint:   key.Sha1Fingerprint,
		ProviderID:        key.ProviderID,
		UserID:            key.UserID,
		SigningKeyID:      sql.NullInt64{Int64: int64(req.CAKeyID), Valid: true},
	}

	updatedKey, err := p.queries.UpdateKey(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update key with certificate: %w", err)
	}

	logrus.Debugf("Updated key %d with new certificate", req.KeyID)
	return mapDBKeyToResponse(updatedKey), nil
}
