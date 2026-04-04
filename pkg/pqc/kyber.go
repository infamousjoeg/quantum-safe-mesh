package pqc

import (
	"crypto/rand"
	"fmt"

	"github.com/cloudflare/circl/kem/kyber/kyber768"
)

type KyberKeyPair struct {
	PublicKey  kyber768.PublicKey
	PrivateKey kyber768.PrivateKey
}

func GenerateKyberKeyPair() (*KyberKeyPair, error) {
	publicKey, privateKey, err := kyber768.GenerateKeyPair(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate Kyber keypair: %w", err)
	}

	return &KyberKeyPair{
		PublicKey:  *publicKey,
		PrivateKey: *privateKey,
	}, nil
}

func (k *KyberKeyPair) GetPublicKeyBytes() []byte {
	bytes, _ := k.PublicKey.MarshalBinary()
	return bytes
}

func (k *KyberKeyPair) GetPrivateKeyBytes() []byte {
	bytes, _ := k.PrivateKey.MarshalBinary()
	return bytes
}

func (k *KyberKeyPair) Encapsulate() ([]byte, []byte, error) {
	ciphertext := make([]byte, 1088)
	sharedSecret := make([]byte, 32)
	seed := make([]byte, 32)
	_, err := rand.Read(seed)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate random seed: %w", err)
	}
	k.PublicKey.EncapsulateTo(ciphertext, sharedSecret, seed)

	return ciphertext, sharedSecret, nil
}

func (k *KyberKeyPair) Decapsulate(ciphertext []byte) ([]byte, error) {
	sharedSecret := make([]byte, 32)
	k.PrivateKey.DecapsulateTo(sharedSecret, ciphertext)

	return sharedSecret, nil
}

func EncapsulateWithPublicKey(publicKeyBytes []byte) ([]byte, []byte, error) {
	if len(publicKeyBytes) != 1184 {
		return nil, nil, fmt.Errorf("invalid public key size: expected %d, got %d", 1184, len(publicKeyBytes))
	}

	var publicKey kyber768.PublicKey
	publicKey.Unpack(publicKeyBytes)

	ciphertext := make([]byte, 1088)
	sharedSecret := make([]byte, 32)
	seed := make([]byte, 32)
	_, err := rand.Read(seed)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate random seed: %w", err)
	}
	publicKey.EncapsulateTo(ciphertext, sharedSecret, seed)

	return ciphertext, sharedSecret, nil
}

func LoadKyberKeyPair(publicKeyBytes, privateKeyBytes []byte) (*KyberKeyPair, error) {
	if len(publicKeyBytes) != 1184 {
		return nil, fmt.Errorf("invalid public key size: expected %d, got %d", 1184, len(publicKeyBytes))
	}

	if len(privateKeyBytes) != 2400 {
		return nil, fmt.Errorf("invalid private key size: expected %d, got %d", 2400, len(privateKeyBytes))
	}

	var publicKey kyber768.PublicKey
	var privateKey kyber768.PrivateKey

	publicKey.Unpack(publicKeyBytes)
	privateKey.Unpack(privateKeyBytes)

	return &KyberKeyPair{
		PublicKey:  publicKey,
		PrivateKey: privateKey,
	}, nil
}
