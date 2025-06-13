package env

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"
)

// GetCertificates returns the SSL certificates to use in the HTTP client
func GetCertificates() (*tls.Config, error) {
	insecure := os.Getenv("ESCAPE_SSL_INSECURE")
	if insecure == "true" {
		return &tls.Config{
			InsecureSkipVerify: true,
		}, nil
	}

	certPath := os.Getenv("ESCAPE_SSL_CERT_PATH")
	if certPath == "" {
		return nil, nil
	}

	cert, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read ESCAPE_SSL_CERT_PATH: %w", err)
	}
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(cert) {
		return nil, errors.New("failed to append certificate to pool")
	}

	return &tls.Config{
		RootCAs: certPool,
	}, nil
}
