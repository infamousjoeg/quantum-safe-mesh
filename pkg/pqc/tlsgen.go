package pqc

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

// GenerateCA creates a self-signed CA certificate and key pair.
// Returns the CA cert, key, and PEM-encoded cert bytes for distribution.
func GenerateCA(keysDir string) (*x509.Certificate, *ecdsa.PrivateKey, []byte, error) {
	caKeyPath := filepath.Join(keysDir, "ca.key")
	caCertPath := filepath.Join(keysDir, "ca.crt")

	// Check if CA already exists
	if certPEM, err := os.ReadFile(caCertPath); err == nil {
		keyPEM, err := os.ReadFile(caKeyPath)
		if err == nil {
			certBlock, _ := pem.Decode(certPEM)
			keyBlock, _ := pem.Decode(keyPEM)
			if certBlock != nil && keyBlock != nil {
				cert, err := x509.ParseCertificate(certBlock.Bytes)
				if err == nil {
					key, err := x509.ParseECPrivateKey(keyBlock.Bytes)
					if err == nil {
						return cert, key, certPEM, nil
					}
				}
			}
		}
	}

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to generate CA key: %w", err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Quantum-Safe Mesh CA"},
			CommonName:   "quantum-safe-mesh-ca",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create CA certificate: %w", err)
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to parse CA certificate: %w", err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyDER, _ := x509.MarshalECPrivateKey(key)
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})

	if err := os.MkdirAll(keysDir, 0755); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create keys directory: %w", err)
	}
	os.WriteFile(caCertPath, certPEM, 0644)
	os.WriteFile(caKeyPath, keyPEM, 0600)

	return cert, key, certPEM, nil
}

// GenerateServiceCert creates a TLS certificate for a service, signed by the CA.
func GenerateServiceCert(keysDir, serviceName string, caCert *x509.Certificate, caKey *ecdsa.PrivateKey) (tls.Certificate, error) {
	certPath := filepath.Join(keysDir, serviceName+"_tls.crt")
	keyPath := filepath.Join(keysDir, serviceName+"_tls.key")

	// Check if cert already exists
	if tlsCert, err := tls.LoadX509KeyPair(certPath, keyPath); err == nil {
		return tlsCert, nil
	}

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to generate service key: %w", err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject: pkix.Name{
			Organization: []string{"Quantum-Safe Mesh"},
			CommonName:   serviceName,
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:  x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
			x509.ExtKeyUsageClientAuth,
		},
		IPAddresses: []net.IP{net.ParseIP("127.0.0.1")},
		DNSNames:    []string{"localhost", serviceName},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, caCert, &key.PublicKey, caKey)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to create service certificate: %w", err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyDER, _ := x509.MarshalECPrivateKey(key)
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})

	os.WriteFile(certPath, certPEM, 0644)
	os.WriteFile(keyPath, keyPEM, 0600)

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to load keypair: %w", err)
	}

	return tlsCert, nil
}

// NewTLSClient creates an HTTP client that trusts the given CA certificate PEM.
func NewTLSClient(caCertPEM []byte) *tls.Config {
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(caCertPEM)
	return &tls.Config{
		RootCAs: pool,
	}
}
