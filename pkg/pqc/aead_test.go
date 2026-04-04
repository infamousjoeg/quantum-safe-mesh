package pqc

import (
	"bytes"
	"crypto/rand"
	"testing"
)

func TestEncryptDecryptRoundTrip(t *testing.T) {
	key := make([]byte, 32)
	rand.Read(key)
	plaintext := []byte(`{"message":"quantum-safe payload","secret":true}`)
	ciphertext, err := Encrypt(key, plaintext)
	if err != nil {
		t.Fatalf("Encrypt() failed: %v", err)
	}
	if bytes.Equal(ciphertext, plaintext) {
		t.Error("ciphertext should differ from plaintext")
	}
	decrypted, err := Decrypt(key, ciphertext)
	if err != nil {
		t.Fatalf("Decrypt() failed: %v", err)
	}
	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("decrypted = %s, want %s", decrypted, plaintext)
	}
}

func TestEncryptDifferentNonces(t *testing.T) {
	key := make([]byte, 32)
	rand.Read(key)
	plaintext := []byte("same input twice")
	ct1, _ := Encrypt(key, plaintext)
	ct2, _ := Encrypt(key, plaintext)
	if bytes.Equal(ct1, ct2) {
		t.Error("two encryptions of same plaintext should produce different ciphertexts")
	}
}

func TestDecryptWrongKey(t *testing.T) {
	key1 := make([]byte, 32)
	key2 := make([]byte, 32)
	rand.Read(key1)
	rand.Read(key2)
	ciphertext, _ := Encrypt(key1, []byte("secret data"))
	_, err := Decrypt(key2, ciphertext)
	if err == nil {
		t.Error("Decrypt() with wrong key should fail")
	}
}

func TestDecryptTamperedCiphertext(t *testing.T) {
	key := make([]byte, 32)
	rand.Read(key)
	ciphertext, _ := Encrypt(key, []byte("secret data"))
	ciphertext[15] ^= 0xFF
	_, err := Decrypt(key, ciphertext)
	if err == nil {
		t.Error("Decrypt() with tampered ciphertext should fail")
	}
}

func TestDecryptTooShort(t *testing.T) {
	key := make([]byte, 32)
	rand.Read(key)
	_, err := Decrypt(key, []byte("short"))
	if err == nil {
		t.Error("Decrypt() with ciphertext shorter than nonce should fail")
	}
}

func TestEncryptInvalidKeySize(t *testing.T) {
	_, err := Encrypt([]byte("too-short"), []byte("data"))
	if err == nil {
		t.Error("Encrypt() with non-32-byte key should fail")
	}
}
