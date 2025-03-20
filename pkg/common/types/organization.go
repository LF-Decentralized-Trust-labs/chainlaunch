package types

import (
	"database/sql"
	"time"
)

// OrganizationDTO represents the shared organization data structure
type OrganizationDTO struct {
	ID              int64
	MspID           string
	Description     sql.NullString
	SignKeyID       sql.NullInt64
	TlsRootKeyID    sql.NullInt64
	SignPublicKey   string
	SignCertificate string
	TlsPublicKey    string
	TlsCertificate  string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
