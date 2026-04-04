package pqc

import (
	"bytes"
	"crypto/rand"
	"testing"
)

func TestGenerateKyberKeyPair(t *testing.T) {
	kp, err := GenerateKyberKeyPair()
	if err != nil {
		t.Fatalf("GenerateKyberKeyPair() failed: %v", err)
	}

	pubBytes := kp.GetPublicKeyBytes()
	privBytes := kp.GetPrivateKeyBytes()

	// Kyber768 key sizes
	if len(pubBytes) != 1184 {
		t.Errorf("public key size = %d, want 1184", len(pubBytes))
	}
	if len(privBytes) != 2400 {
		t.Errorf("private key size = %d, want 2400", len(privBytes))
	}
}

func TestKyberEncapsulateDecapsulate(t *testing.T) {
	kp, err := GenerateKyberKeyPair()
	if err != nil {
		t.Fatalf("keygen failed: %v", err)
	}

	ciphertext, sharedSecret1, err := kp.Encapsulate()
	if err != nil {
		t.Fatalf("Encapsulate() failed: %v", err)
	}

	if len(ciphertext) != 1088 {
		t.Errorf("ciphertext size = %d, want 1088", len(ciphertext))
	}
	if len(sharedSecret1) != 32 {
		t.Errorf("shared secret size = %d, want 32", len(sharedSecret1))
	}

	sharedSecret2, err := kp.Decapsulate(ciphertext)
	if err != nil {
		t.Fatalf("Decapsulate() failed: %v", err)
	}

	if !bytes.Equal(sharedSecret1, sharedSecret2) {
		t.Error("shared secrets from encapsulate and decapsulate do not match")
	}
}

func TestKyberEncapsulateWithPublicKey(t *testing.T) {
	kp, err := GenerateKyberKeyPair()
	if err != nil {
		t.Fatalf("keygen failed: %v", err)
	}

	pubBytes := kp.GetPublicKeyBytes()
	ciphertext, sharedSecret1, err := EncapsulateWithPublicKey(pubBytes)
	if err != nil {
		t.Fatalf("EncapsulateWithPublicKey() failed: %v", err)
	}

	sharedSecret2, err := kp.Decapsulate(ciphertext)
	if err != nil {
		t.Fatalf("Decapsulate() failed: %v", err)
	}

	if !bytes.Equal(sharedSecret1, sharedSecret2) {
		t.Error("shared secrets should match when encapsulating with serialized public key")
	}
}

func TestKyberDecapsulateWrongCiphertext(t *testing.T) {
	kp, err := GenerateKyberKeyPair()
	if err != nil {
		t.Fatalf("keygen failed: %v", err)
	}

	_, sharedSecret1, err := kp.Encapsulate()
	if err != nil {
		t.Fatalf("Encapsulate() failed: %v", err)
	}

	// Create random ciphertext (wrong)
	wrongCiphertext := make([]byte, 1088)
	rand.Read(wrongCiphertext)

	// Kyber768 is IND-CCA2: decapsulation doesn't error on wrong ciphertext,
	// it returns a pseudorandom value that differs from the real shared secret
	sharedSecret2, err := kp.Decapsulate(wrongCiphertext)
	if err != nil {
		t.Fatalf("Decapsulate() should not error even with wrong ciphertext: %v", err)
	}

	if bytes.Equal(sharedSecret1, sharedSecret2) {
		t.Error("decapsulating wrong ciphertext should produce different shared secret")
	}
}

func TestKyberKeyPairSerialization(t *testing.T) {
	original, err := GenerateKyberKeyPair()
	if err != nil {
		t.Fatalf("keygen failed: %v", err)
	}

	pubBytes := original.GetPublicKeyBytes()
	privBytes := original.GetPrivateKeyBytes()

	loaded, err := LoadKyberKeyPair(pubBytes, privBytes)
	if err != nil {
		t.Fatalf("LoadKyberKeyPair() failed: %v", err)
	}

	// Encapsulate with original, decapsulate with loaded
	ciphertext, sharedSecret1, err := original.Encapsulate()
	if err != nil {
		t.Fatalf("Encapsulate() failed: %v", err)
	}

	sharedSecret2, err := loaded.Decapsulate(ciphertext)
	if err != nil {
		t.Fatalf("Decapsulate() with loaded key failed: %v", err)
	}

	if !bytes.Equal(sharedSecret1, sharedSecret2) {
		t.Error("shared secrets should match after key serialization round-trip")
	}
}

func TestEncapsulateWithPublicKeyInvalidSize(t *testing.T) {
	_, _, err := EncapsulateWithPublicKey([]byte("too-short"))
	if err == nil {
		t.Error("EncapsulateWithPublicKey() should fail with invalid key size")
	}
}

func TestLoadKyberKeyPairInvalidSizes(t *testing.T) {
	_, err := LoadKyberKeyPair([]byte("short"), []byte("short"))
	if err == nil {
		t.Error("LoadKyberKeyPair() should fail with invalid key sizes")
	}
}
