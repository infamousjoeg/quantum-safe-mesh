package pqc

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoadKeyPair(t *testing.T) {
	// Use a temp directory so we don't pollute the real keys dir
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	dkp, err := GenerateDilithiumKeyPair()
	if err != nil {
		t.Fatalf("Dilithium keygen failed: %v", err)
	}
	kkp, err := GenerateKyberKeyPair()
	if err != nil {
		t.Fatalf("Kyber keygen failed: %v", err)
	}

	err = SaveKeyPair("test-service", dkp, kkp)
	if err != nil {
		t.Fatalf("SaveKeyPair() failed: %v", err)
	}

	// Verify files exist
	for _, name := range []string{
		"test-service_dilithium.pub",
		"test-service_dilithium.key",
		"test-service_kyber.pub",
		"test-service_kyber.key",
	} {
		path := filepath.Join("keys", name)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file %s to exist", path)
		}
	}

	// Verify private keys have restricted permissions
	for _, name := range []string{"test-service_dilithium.key", "test-service_kyber.key"} {
		info, err := os.Stat(filepath.Join("keys", name))
		if err != nil {
			t.Fatalf("stat %s: %v", name, err)
		}
		perm := info.Mode().Perm()
		if perm != 0600 {
			t.Errorf("private key %s has permissions %o, want 0600", name, perm)
		}
	}

	// Load and verify round-trip
	loadedDkp, loadedKkp, err := LoadKeyPair("test-service")
	if err != nil {
		t.Fatalf("LoadKeyPair() failed: %v", err)
	}

	// Verify Dilithium round-trip: sign with original, verify with loaded
	data := []byte("persistence test")
	sig, err := dkp.Sign(data)
	if err != nil {
		t.Fatalf("Sign() failed: %v", err)
	}
	err = VerifyDilithiumSignature(loadedDkp.GetPublicKeyBytes(), data, sig)
	if err != nil {
		t.Errorf("Dilithium key round-trip failed: %v", err)
	}

	// Verify Kyber round-trip: encapsulate with original, decapsulate with loaded
	ct, ss1, err := kkp.Encapsulate()
	if err != nil {
		t.Fatalf("Encapsulate() failed: %v", err)
	}
	ss2, err := loadedKkp.Decapsulate(ct)
	if err != nil {
		t.Fatalf("Decapsulate() failed: %v", err)
	}
	if string(ss1) != string(ss2) {
		t.Error("Kyber shared secrets don't match after key persistence round-trip")
	}
}

func TestDeserializePublicKeyFromJSON_Base64(t *testing.T) {
	// Generate a real key to get valid bytes
	kp, err := GenerateDilithiumKeyPair()
	if err != nil {
		t.Fatalf("keygen failed: %v", err)
	}
	pubBytes := kp.GetPublicKeyBytes()
	encoded := base64.StdEncoding.EncodeToString(pubBytes)

	keyData := map[string]interface{}{
		"public_key": encoded,
	}

	result, err := DeserializePublicKeyFromJSON(keyData)
	if err != nil {
		t.Fatalf("DeserializePublicKeyFromJSON(base64) failed: %v", err)
	}

	if len(result) != len(pubBytes) {
		t.Errorf("deserialized key size = %d, want %d", len(result), len(pubBytes))
	}
}

func TestDeserializePublicKeyFromJSON_Array(t *testing.T) {
	original := []byte{0x01, 0x02, 0x03, 0xFF}
	asInterface := make([]interface{}, len(original))
	for i, b := range original {
		asInterface[i] = float64(b)
	}

	keyData := map[string]interface{}{
		"public_key": asInterface,
	}

	result, err := DeserializePublicKeyFromJSON(keyData)
	if err != nil {
		t.Fatalf("DeserializePublicKeyFromJSON(array) failed: %v", err)
	}

	if string(result) != string(original) {
		t.Errorf("deserialized key = %v, want %v", result, original)
	}
}

func TestDeserializePublicKeyFromJSON_Missing(t *testing.T) {
	keyData := map[string]interface{}{
		"not_a_key": "value",
	}

	_, err := DeserializePublicKeyFromJSON(keyData)
	if err == nil {
		t.Error("DeserializePublicKeyFromJSON() should fail when public_key field is missing")
	}
}
