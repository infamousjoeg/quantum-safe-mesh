package pqc

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func SaveKeyPair(serviceID string, dilithiumKeyPair *DilithiumKeyPair, kyberKeyPair *KyberKeyPair) error {
	keysDir := "keys"
	if err := os.MkdirAll(keysDir, 0755); err != nil {
		return fmt.Errorf("failed to create keys directory: %w", err)
	}

	dilithiumPubPath := filepath.Join(keysDir, fmt.Sprintf("%s_dilithium.pub", serviceID))
	dilithiumPrivPath := filepath.Join(keysDir, fmt.Sprintf("%s_dilithium.key", serviceID))
	kyberPubPath := filepath.Join(keysDir, fmt.Sprintf("%s_kyber.pub", serviceID))
	kyberPrivPath := filepath.Join(keysDir, fmt.Sprintf("%s_kyber.key", serviceID))

	dilithiumPubBytes := dilithiumKeyPair.GetPublicKeyBytes()
	dilithiumPrivBytes := dilithiumKeyPair.GetPrivateKeyBytes()
	kyberPubBytes := kyberKeyPair.GetPublicKeyBytes()
	kyberPrivBytes := kyberKeyPair.GetPrivateKeyBytes()

	if err := os.WriteFile(dilithiumPubPath, dilithiumPubBytes, 0644); err != nil {
		return fmt.Errorf("failed to save Dilithium public key: %w", err)
	}

	if err := os.WriteFile(dilithiumPrivPath, dilithiumPrivBytes, 0600); err != nil {
		return fmt.Errorf("failed to save Dilithium private key: %w", err)
	}

	if err := os.WriteFile(kyberPubPath, kyberPubBytes, 0644); err != nil {
		return fmt.Errorf("failed to save Kyber public key: %w", err)
	}

	if err := os.WriteFile(kyberPrivPath, kyberPrivBytes, 0600); err != nil {
		return fmt.Errorf("failed to save Kyber private key: %w", err)
	}

	log.Printf("✅ Keys saved for service %s in %s directory", serviceID, keysDir)
	return nil
}

func LoadKeyPair(serviceID string) (*DilithiumKeyPair, *KyberKeyPair, error) {
	keysDir := "keys"

	dilithiumPubPath := filepath.Join(keysDir, fmt.Sprintf("%s_dilithium.pub", serviceID))
	dilithiumPrivPath := filepath.Join(keysDir, fmt.Sprintf("%s_dilithium.key", serviceID))
	kyberPubPath := filepath.Join(keysDir, fmt.Sprintf("%s_kyber.pub", serviceID))
	kyberPrivPath := filepath.Join(keysDir, fmt.Sprintf("%s_kyber.key", serviceID))

	dilithiumPubBytes, err := os.ReadFile(dilithiumPubPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read Dilithium public key: %w", err)
	}

	dilithiumPrivBytes, err := os.ReadFile(dilithiumPrivPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read Dilithium private key: %w", err)
	}

	kyberPubBytes, err := os.ReadFile(kyberPubPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read Kyber public key: %w", err)
	}

	kyberPrivBytes, err := os.ReadFile(kyberPrivPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read Kyber private key: %w", err)
	}

	dilithiumKeyPair, err := LoadDilithiumKeyPair(dilithiumPubBytes, dilithiumPrivBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load Dilithium keypair: %w", err)
	}

	kyberKeyPair, err := LoadKyberKeyPair(kyberPubBytes, kyberPrivBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load Kyber keypair: %w", err)
	}

	log.Printf("✅ Keys loaded for service %s", serviceID)
	return dilithiumKeyPair, kyberKeyPair, nil
}

func BenchmarkRSAvsDialithium() {
	log.Println("\n🚀 Performance Comparison: RSA vs Dilithium3")
	log.Println(strings.Repeat("=", 50))

	testData := []byte("This is a test message for signature performance comparison")

	log.Println("\n📊 RSA-2048 Performance:")
	rsaPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Printf("Failed to generate RSA key: %v", err)
		return
	}

	start := time.Now()
	hash := sha256.Sum256(testData)
	rsaSignature, err := rsa.SignPKCS1v15(rand.Reader, rsaPrivateKey, 0, hash[:])
	if err != nil {
		log.Printf("Failed to sign with RSA: %v", err)
		return
	}
	rsaSignTime := time.Since(start)

	start = time.Now()
	err = rsa.VerifyPKCS1v15(&rsaPrivateKey.PublicKey, 0, hash[:], rsaSignature)
	rsaVerifyTime := time.Since(start)

	if err != nil {
		log.Printf("RSA verification failed: %v", err)
		return
	}

	rsaPubKeyBytes, _ := x509.MarshalPKIXPublicKey(&rsaPrivateKey.PublicKey)
	rsaPrivKeyBytes := x509.MarshalPKCS1PrivateKey(rsaPrivateKey)

	log.Printf("  🖊️  Sign time: %v", rsaSignTime)
	log.Printf("  🔍 Verify time: %v", rsaVerifyTime)
	log.Printf("  📏 Public key size: %d bytes", len(rsaPubKeyBytes))
	log.Printf("  📏 Private key size: %d bytes", len(rsaPrivKeyBytes))
	log.Printf("  📏 Signature size: %d bytes", len(rsaSignature))

	log.Println("\n📊 Dilithium3 Performance:")
	dilithiumKeyPair, err := GenerateDilithiumKeyPair()
	if err != nil {
		log.Printf("Failed to generate Dilithium keypair: %v", err)
		return
	}

	start = time.Now()
	dilithiumSignature, err := dilithiumKeyPair.Sign(testData)
	if err != nil {
		log.Printf("Failed to sign with Dilithium: %v", err)
		return
	}
	dilithiumSignTime := time.Since(start)

	start = time.Now()
	err = VerifyDilithiumSignature(dilithiumKeyPair.PublicKey.Bytes(), testData, dilithiumSignature)
	dilithiumVerifyTime := time.Since(start)

	if err != nil {
		log.Printf("Dilithium verification failed: %v", err)
		return
	}

	log.Printf("  🖊️  Sign time: %v", dilithiumSignTime)
	log.Printf("  🔍 Verify time: %v", dilithiumVerifyTime)
	log.Printf("  📏 Public key size: %d bytes", len(dilithiumKeyPair.PublicKey.Bytes()))
	log.Printf("  📏 Private key size: %d bytes", len(dilithiumKeyPair.PrivateKey.Bytes()))
	log.Printf("  📏 Signature size: %d bytes", len(dilithiumSignature))

	log.Println("\n📈 Performance Summary:")
	log.Printf("  Sign time ratio (Dilithium/RSA): %.2fx",
		float64(dilithiumSignTime.Nanoseconds())/float64(rsaSignTime.Nanoseconds()))
	log.Printf("  Verify time ratio (Dilithium/RSA): %.2fx",
		float64(dilithiumVerifyTime.Nanoseconds())/float64(rsaVerifyTime.Nanoseconds()))
	log.Printf("  Public key size ratio (Dilithium/RSA): %.2fx",
		float64(len(dilithiumKeyPair.GetPublicKeyBytes()))/float64(len(rsaPubKeyBytes)))
	log.Printf("  Signature size ratio (Dilithium/RSA): %.2fx",
		float64(len(dilithiumSignature))/float64(len(rsaSignature)))
	log.Println(strings.Repeat("=", 50))
}
