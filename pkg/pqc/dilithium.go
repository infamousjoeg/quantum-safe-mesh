package pqc

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/cloudflare/circl/sign/dilithium/mode3"
)

type DilithiumKeyPair struct {
	PublicKey  mode3.PublicKey
	PrivateKey mode3.PrivateKey
}

func GenerateDilithiumKeyPair() (*DilithiumKeyPair, error) {
	log.Println("🔑 Generating Dilithium3 keypair...")
	start := time.Now()

	publicKey, privateKey, err := mode3.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate Dilithium keypair: %w", err)
	}

	duration := time.Since(start)
	log.Printf("✅ Dilithium3 keypair generated in %v", duration)
	log.Printf("📏 Public key size: %d bytes, Private key size: %d bytes",
		mode3.PublicKeySize, mode3.PrivateKeySize)

	return &DilithiumKeyPair{
		PublicKey:  *publicKey,
		PrivateKey: *privateKey,
	}, nil
}

func (d *DilithiumKeyPair) Sign(data []byte) ([]byte, error) {
	log.Printf("🖊️  Signing data with Dilithium3 (data size: %d bytes)", len(data))
	start := time.Now()

	signature := make([]byte, mode3.SignatureSize)
	mode3.SignTo(&d.PrivateKey, data, signature)

	duration := time.Since(start)
	log.Printf("✅ Data signed in %v (signature size: %d bytes)", duration, len(signature))

	return signature, nil
}

func (d *DilithiumKeyPair) GetPublicKeyBytes() []byte {
	return d.PublicKey.Bytes()
}

func (d *DilithiumKeyPair) GetPrivateKeyBytes() []byte {
	return d.PrivateKey.Bytes()
}

func VerifyDilithiumSignature(publicKeyBytes, data, signature []byte) error {
	log.Printf("🔍 Verifying Dilithium3 signature (data: %d bytes, sig: %d bytes)",
		len(data), len(signature))
	start := time.Now()

	if len(publicKeyBytes) != mode3.PublicKeySize {
		return fmt.Errorf("invalid public key size: expected %d, got %d", mode3.PublicKeySize, len(publicKeyBytes))
	}

	var publicKey mode3.PublicKey
	err := publicKey.UnmarshalBinary(publicKeyBytes)
	if err != nil {
		return fmt.Errorf("failed to unmarshal public key: %w", err)
	}

	if !mode3.Verify(&publicKey, data, signature) {
		duration := time.Since(start)
		log.Printf("❌ Signature verification failed in %v", duration)
		return fmt.Errorf("invalid signature")
	}

	duration := time.Since(start)
	log.Printf("✅ Signature verified successfully in %v", duration)

	return nil
}

func (d *DilithiumKeyPair) SignRequest(serviceID string, data interface{}) ([]byte, error) {
	requestData := map[string]interface{}{
		"service_id": serviceID,
		"timestamp":  time.Now(),
		"data":       data,
	}

	payload, err := json.Marshal(requestData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request data: %w", err)
	}

	return d.Sign(payload)
}

func LoadDilithiumKeyPair(publicKeyBytes, privateKeyBytes []byte) (*DilithiumKeyPair, error) {
	if len(publicKeyBytes) != mode3.PublicKeySize {
		return nil, fmt.Errorf("invalid public key size: expected %d, got %d", mode3.PublicKeySize, len(publicKeyBytes))
	}

	if len(privateKeyBytes) != mode3.PrivateKeySize {
		return nil, fmt.Errorf("invalid private key size: expected %d, got %d", mode3.PrivateKeySize, len(privateKeyBytes))
	}

	var publicKey mode3.PublicKey
	var privateKey mode3.PrivateKey

	err := publicKey.UnmarshalBinary(publicKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal public key: %w", err)
	}

	err = privateKey.UnmarshalBinary(privateKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal private key: %w", err)
	}

	return &DilithiumKeyPair{
		PublicKey:  publicKey,
		PrivateKey: privateKey,
	}, nil
}
