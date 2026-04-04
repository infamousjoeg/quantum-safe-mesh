package pqc

import (
	"testing"

	"github.com/cloudflare/circl/sign/dilithium/mode3"
)

func TestGenerateDilithiumKeyPair(t *testing.T) {
	kp, err := GenerateDilithiumKeyPair()
	if err != nil {
		t.Fatalf("GenerateDilithiumKeyPair() failed: %v", err)
	}

	pubBytes := kp.GetPublicKeyBytes()
	privBytes := kp.GetPrivateKeyBytes()

	if len(pubBytes) != mode3.PublicKeySize {
		t.Errorf("public key size = %d, want %d", len(pubBytes), mode3.PublicKeySize)
	}
	if len(privBytes) != mode3.PrivateKeySize {
		t.Errorf("private key size = %d, want %d", len(privBytes), mode3.PrivateKeySize)
	}
}

func TestDilithiumSignAndVerify(t *testing.T) {
	kp, err := GenerateDilithiumKeyPair()
	if err != nil {
		t.Fatalf("keygen failed: %v", err)
	}

	data := []byte("test message for quantum-safe signing")
	sig, err := kp.Sign(data)
	if err != nil {
		t.Fatalf("Sign() failed: %v", err)
	}

	if len(sig) != mode3.SignatureSize {
		t.Errorf("signature size = %d, want %d", len(sig), mode3.SignatureSize)
	}

	err = VerifyDilithiumSignature(kp.GetPublicKeyBytes(), data, sig)
	if err != nil {
		t.Errorf("VerifyDilithiumSignature() failed on valid signature: %v", err)
	}
}

func TestDilithiumVerifyWrongData(t *testing.T) {
	kp, err := GenerateDilithiumKeyPair()
	if err != nil {
		t.Fatalf("keygen failed: %v", err)
	}

	data := []byte("original message")
	sig, err := kp.Sign(data)
	if err != nil {
		t.Fatalf("Sign() failed: %v", err)
	}

	tampered := []byte("tampered message")
	err = VerifyDilithiumSignature(kp.GetPublicKeyBytes(), tampered, sig)
	if err == nil {
		t.Error("VerifyDilithiumSignature() should fail with tampered data")
	}
}

func TestDilithiumVerifyWrongKey(t *testing.T) {
	kp1, err := GenerateDilithiumKeyPair()
	if err != nil {
		t.Fatalf("keygen 1 failed: %v", err)
	}
	kp2, err := GenerateDilithiumKeyPair()
	if err != nil {
		t.Fatalf("keygen 2 failed: %v", err)
	}

	data := []byte("test message")
	sig, err := kp1.Sign(data)
	if err != nil {
		t.Fatalf("Sign() failed: %v", err)
	}

	err = VerifyDilithiumSignature(kp2.GetPublicKeyBytes(), data, sig)
	if err == nil {
		t.Error("VerifyDilithiumSignature() should fail with wrong public key")
	}
}

func TestDilithiumKeyPairSerialization(t *testing.T) {
	original, err := GenerateDilithiumKeyPair()
	if err != nil {
		t.Fatalf("keygen failed: %v", err)
	}

	pubBytes := original.GetPublicKeyBytes()
	privBytes := original.GetPrivateKeyBytes()

	loaded, err := LoadDilithiumKeyPair(pubBytes, privBytes)
	if err != nil {
		t.Fatalf("LoadDilithiumKeyPair() failed: %v", err)
	}

	// Sign with original, verify with loaded
	data := []byte("round-trip test")
	sig, err := original.Sign(data)
	if err != nil {
		t.Fatalf("Sign() failed: %v", err)
	}

	err = VerifyDilithiumSignature(loaded.GetPublicKeyBytes(), data, sig)
	if err != nil {
		t.Errorf("signature from original should verify with loaded key: %v", err)
	}

	// Sign with loaded, verify with original
	sig2, err := loaded.Sign(data)
	if err != nil {
		t.Fatalf("Sign() with loaded key failed: %v", err)
	}

	err = VerifyDilithiumSignature(original.GetPublicKeyBytes(), data, sig2)
	if err != nil {
		t.Errorf("signature from loaded should verify with original key: %v", err)
	}
}

func TestLoadDilithiumKeyPairInvalidSizes(t *testing.T) {
	_, err := LoadDilithiumKeyPair([]byte("short"), []byte("short"))
	if err == nil {
		t.Error("LoadDilithiumKeyPair() should fail with invalid key sizes")
	}
}
