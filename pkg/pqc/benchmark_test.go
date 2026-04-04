package pqc

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"testing"
)

var testPayload = []byte("This is a test message for signature performance comparison")

// RSA-2048 benchmarks

func BenchmarkRSA2048Sign(b *testing.B) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		b.Fatal(err)
	}
	hash := sha256.Sum256(testPayload)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rsa.SignPKCS1v15(rand.Reader, key, 0, hash[:])
	}
}

func BenchmarkRSA2048Verify(b *testing.B) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		b.Fatal(err)
	}
	hash := sha256.Sum256(testPayload)
	sig, _ := rsa.SignPKCS1v15(rand.Reader, key, 0, hash[:])
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rsa.VerifyPKCS1v15(&key.PublicKey, 0, hash[:], sig)
	}
}

// Dilithium3 benchmarks

func BenchmarkDilithium3KeyGen(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateDilithiumKeyPair()
	}
}

func BenchmarkDilithium3Sign(b *testing.B) {
	kp, err := GenerateDilithiumKeyPair()
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		kp.Sign(testPayload)
	}
}

func BenchmarkDilithium3Verify(b *testing.B) {
	kp, err := GenerateDilithiumKeyPair()
	if err != nil {
		b.Fatal(err)
	}
	sig, _ := kp.Sign(testPayload)
	pubBytes := kp.GetPublicKeyBytes()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		VerifyDilithiumSignature(pubBytes, testPayload, sig)
	}
}

// Kyber768 benchmarks

func BenchmarkKyber768KeyGen(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateKyberKeyPair()
	}
}

func BenchmarkKyber768Encapsulate(b *testing.B) {
	kp, err := GenerateKyberKeyPair()
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		kp.Encapsulate()
	}
}

func BenchmarkKyber768Decapsulate(b *testing.B) {
	kp, err := GenerateKyberKeyPair()
	if err != nil {
		b.Fatal(err)
	}
	ct, _, _ := kp.Encapsulate()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		kp.Decapsulate(ct)
	}
}

// Key size comparison test (not a benchmark, but provides the data)
func TestKeySizeComparison(t *testing.T) {
	// RSA-2048
	rsaKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	rsaPub, _ := x509.MarshalPKIXPublicKey(&rsaKey.PublicKey)
	rsaPriv := x509.MarshalPKCS1PrivateKey(rsaKey)
	hash := sha256.Sum256(testPayload)
	rsaSig, _ := rsa.SignPKCS1v15(rand.Reader, rsaKey, 0, hash[:])

	// Dilithium3
	dkp, _ := GenerateDilithiumKeyPair()
	dSig, _ := dkp.Sign(testPayload)

	// Kyber768
	kkp, _ := GenerateKyberKeyPair()

	t.Logf("=== Key Size Comparison ===")
	t.Logf("RSA-2048:    pub=%d priv=%d sig=%d", len(rsaPub), len(rsaPriv), len(rsaSig))
	t.Logf("Dilithium3:  pub=%d priv=%d sig=%d", len(dkp.GetPublicKeyBytes()), len(dkp.GetPrivateKeyBytes()), len(dSig))
	t.Logf("Kyber768:    pub=%d priv=%d ct=1088 ss=32", len(kkp.GetPublicKeyBytes()), len(kkp.GetPrivateKeyBytes()))
	t.Logf("Size ratios (Dilithium/RSA): pub=%.1fx priv=%.1fx sig=%.1fx",
		float64(len(dkp.GetPublicKeyBytes()))/float64(len(rsaPub)),
		float64(len(dkp.GetPrivateKeyBytes()))/float64(len(rsaPriv)),
		float64(len(dSig))/float64(len(rsaSig)))
}
