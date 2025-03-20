package certutils

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
)

func EncodeX509Certificate(crt *x509.Certificate) []byte {
	pemPk := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: crt.Raw,
	})
	return pemPk
}

func ParseX509Certificate(contents []byte) (*x509.Certificate, error) {
	if len(contents) == 0 {
		return nil, errors.New("certificate pem is empty")
	}
	block, _ := pem.Decode(contents)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}
	crt, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}
	return crt, nil
}
