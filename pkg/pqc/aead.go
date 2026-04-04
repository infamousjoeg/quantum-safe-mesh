package pqc

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

const nonceSize = 12

func Encrypt(key, plaintext []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("invalid key size: expected 32, got %d", len(key))
	}
	GlobalMetrics.EncryptCount.Add(1)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}
	nonce := make([]byte, nonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

func Decrypt(key, ciphertext []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("invalid key size: expected 32, got %d", len(key))
	}
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short: %d bytes", len(ciphertext))
	}
	GlobalMetrics.DecryptCount.Add(1)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}
	nonce := ciphertext[:nonceSize]
	encrypted := ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, encrypted, nil)
	if err != nil {
		return nil, fmt.Errorf("GCM decryption failed: %w", err)
	}
	return plaintext, nil
}
