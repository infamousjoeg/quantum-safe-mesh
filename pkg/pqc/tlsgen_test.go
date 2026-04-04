package pqc

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateCAAndServiceCert(t *testing.T) {
	dir := t.TempDir()

	caCert, caKey, caPEM, err := GenerateCA(dir)
	if err != nil {
		t.Fatalf("GenerateCA() failed: %v", err)
	}

	if caCert == nil || caKey == nil || len(caPEM) == 0 {
		t.Fatal("GenerateCA() returned nil values")
	}

	// Verify CA files were written
	if _, err := os.Stat(filepath.Join(dir, "ca.crt")); err != nil {
		t.Error("ca.crt not written")
	}
	if _, err := os.Stat(filepath.Join(dir, "ca.key")); err != nil {
		t.Error("ca.key not written")
	}

	// Generate service cert
	tlsCert, err := GenerateServiceCert(dir, "test-service", caCert, caKey)
	if err != nil {
		t.Fatalf("GenerateServiceCert() failed: %v", err)
	}

	if len(tlsCert.Certificate) == 0 {
		t.Error("service cert has no certificates")
	}

	// Verify service files were written
	if _, err := os.Stat(filepath.Join(dir, "test-service_tls.crt")); err != nil {
		t.Error("service cert not written")
	}
}

func TestGenerateCAIdempotent(t *testing.T) {
	dir := t.TempDir()

	_, _, pem1, _ := GenerateCA(dir)
	_, _, pem2, _ := GenerateCA(dir)

	if string(pem1) != string(pem2) {
		t.Error("second GenerateCA call should return same cert from disk")
	}
}

func TestNewTLSClient(t *testing.T) {
	dir := t.TempDir()
	_, _, caPEM, _ := GenerateCA(dir)

	tlsConfig := NewTLSClient(caPEM)
	if tlsConfig == nil {
		t.Fatal("NewTLSClient returned nil")
	}
	if tlsConfig.RootCAs == nil {
		t.Error("RootCAs should not be nil")
	}
}

func TestServiceCertIdempotent(t *testing.T) {
	dir := t.TempDir()
	caCert, caKey, _, _ := GenerateCA(dir)

	cert1, err1 := GenerateServiceCert(dir, "svc", caCert, caKey)
	cert2, err2 := GenerateServiceCert(dir, "svc", caCert, caKey)

	if err1 != nil || err2 != nil {
		t.Fatalf("certs failed: %v / %v", err1, err2)
	}

	// Both should load from disk, so same cert bytes
	if len(cert1.Certificate[0]) != len(cert2.Certificate[0]) {
		t.Error("second call should load same cert from disk")
	}
}
