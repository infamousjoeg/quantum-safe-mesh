package pqc

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cloudflare/circl/sign/dilithium/mode3"
)

type DilithiumKeyPair struct {
	PublicKey  mode3.PublicKey
	PrivateKey mode3.PrivateKey
}

func GenerateDilithiumKeyPair() (*DilithiumKeyPair, error) {
	publicKey, privateKey, err := mode3.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate Dilithium keypair: %w", err)
	}

	return &DilithiumKeyPair{
		PublicKey:  *publicKey,
		PrivateKey: *privateKey,
	}, nil
}

func (d *DilithiumKeyPair) Sign(data []byte) ([]byte, error) {
	GlobalMetrics.SignCount.Add(1)
	signature := make([]byte, mode3.SignatureSize)
	mode3.SignTo(&d.PrivateKey, data, signature)

	return signature, nil
}

func (d *DilithiumKeyPair) GetPublicKeyBytes() []byte {
	return d.PublicKey.Bytes()
}

func (d *DilithiumKeyPair) GetPrivateKeyBytes() []byte {
	return d.PrivateKey.Bytes()
}

func VerifyDilithiumSignature(publicKeyBytes, data, signature []byte) error {
	GlobalMetrics.VerifyCount.Add(1)
	if len(publicKeyBytes) != mode3.PublicKeySize {
		return fmt.Errorf("invalid public key size: expected %d, got %d", mode3.PublicKeySize, len(publicKeyBytes))
	}

	var publicKey mode3.PublicKey
	err := publicKey.UnmarshalBinary(publicKeyBytes)
	if err != nil {
		return fmt.Errorf("failed to unmarshal public key: %w", err)
	}

	if !mode3.Verify(&publicKey, data, signature) {
		GlobalMetrics.VerifyFailCount.Add(1)
		return fmt.Errorf("invalid signature")
	}

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
