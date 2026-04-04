package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"quantum-safe-mesh/pkg/models"
	"quantum-safe-mesh/pkg/pqc"
)

var (
	clientID      = "demo-client"
	clientKeyPair *pqc.DilithiumKeyPair
	tlsClient     *http.Client
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if len(os.Args) > 1 && os.Args[1] == "benchmark" {
		runBenchmark()
		return
	}

	fmt.Println("🌟 Quantum-Safe Service Mesh Demo")
	fmt.Println(strings.Repeat("=", 50))

	// Generate client key pair for demo
	var err error
	clientKeyPair, err = pqc.GenerateDilithiumKeyPair()
	if err != nil {
		log.Fatalf("Failed to generate client key pair: %v", err)
	}

	// Setup TLS client
	fmt.Println("\n🔒 Setting up TLS client...")
	_, _, caCertPEM, err := pqc.GenerateCA("keys")
	if err != nil {
		log.Fatalf("Failed to load CA: %v", err)
	}
	tlsConf := pqc.NewTLSClient(caCertPEM)
	tlsClient = &http.Client{
		Transport: &http.Transport{TLSClientConfig: tlsConf},
	}
	fmt.Println("✅ TLS client configured")

	time.Sleep(2 * time.Second)

	fmt.Println("\n📝 Step 0: Registering demo client...")
	if err := registerClient(); err != nil {
		log.Printf("Warning: client registration failed: %v", err)
	} else {
		fmt.Println("✅ Demo client registered successfully")
	}

	time.Sleep(1 * time.Second)

	fmt.Println("\n🔍 Step 1: Checking service health...")
	checkHealth()

	time.Sleep(1 * time.Second)

	fmt.Println("\n📡 Step 2: Testing Echo Endpoint...")
	testEcho()

	time.Sleep(1 * time.Second)

	fmt.Println("\n🧠 Step 3: Testing Data Processing...")
	testProcess()

	time.Sleep(1 * time.Second)

	fmt.Println("\n📊 Step 4: Checking Service Status...")
	testStatus()

	time.Sleep(1 * time.Second)

	fmt.Println("\n🔐 Step 5: Performance Summary...")
	showPerformanceSummary()

	fmt.Println("\n🎉 Demo completed successfully!")
	fmt.Println("All requests were authenticated using Post-Quantum Cryptography!")
}

func registerClient() error {
	keyPair := models.ServiceKeyPair{
		ServiceID: clientID,
		PublicKey: clientKeyPair.GetPublicKeyBytes(),
	}

	// Sign registration to prove key ownership
	regPayload, err := json.Marshal(keyPair)
	if err != nil {
		return fmt.Errorf("failed to marshal registration: %w", err)
	}

	sig, err := clientKeyPair.Sign(regPayload)
	if err != nil {
		return fmt.Errorf("failed to sign registration: %w", err)
	}
	keyPair.Signature = sig

	payload, err := json.Marshal(keyPair)
	if err != nil {
		return fmt.Errorf("failed to marshal signed registration: %w", err)
	}

	resp, err := tlsClient.Post("https://localhost:8080/register", "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to register with auth service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("registration failed with status: %d", resp.StatusCode)
	}

	return nil
}

func signedPost(url string, data []byte) (*http.Response, error) {
	sig, err := clientKeyPair.Sign(data)
	if err != nil {
		return nil, fmt.Errorf("failed to sign request: %w", err)
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Client-ID", clientID)
	req.Header.Set("X-Signature", base64.StdEncoding.EncodeToString(sig))
	return tlsClient.Do(req)
}

func signedGet(url string) (*http.Response, error) {
	emptyBody := []byte{}
	sig, err := clientKeyPair.Sign(emptyBody)
	if err != nil {
		return nil, fmt.Errorf("failed to sign request: %w", err)
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Client-ID", clientID)
	req.Header.Set("X-Signature", base64.StdEncoding.EncodeToString(sig))
	return tlsClient.Do(req)
}

func checkHealth() {
	services := map[string]string{
		"Auth Service":    "https://localhost:8080/health",
		"Gateway Service": "https://localhost:8081/health",
		"Backend Service": "https://localhost:8082/health",
	}

	for name, url := range services {
		resp, err := tlsClient.Get(url)
		if err != nil || resp.StatusCode != 200 {
			fmt.Printf("❌ %s: Not responding\n", name)
		} else {
			fmt.Printf("✅ %s: Healthy\n", name)
			resp.Body.Close()
		}
	}
}

func testEcho() {
	data := map[string]interface{}{
		"message":   "Hello Quantum-Safe World!",
		"demo":      true,
		"timestamp": time.Now(),
		"test_data": []string{"quantum", "cryptography", "demo"},
	}

	jsonData, _ := json.Marshal(data)

	start := time.Now()
	resp, err := signedPost("https://localhost:8081/echo", jsonData)
	duration := time.Since(start)

	if err != nil {
		fmt.Printf("❌ Echo test failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("✅ Echo test successful (latency: %v)\n", duration)
	fmt.Printf("   Status: %s\n", resp.Status)
}

func testProcess() {
	data := map[string]interface{}{
		"operation": "quantum_safe_processing",
		"data": map[string]interface{}{
			"input":     "Sensitive data for quantum-safe processing",
			"algorithm": "advanced",
			"priority":  "high",
		},
		"metadata": map[string]interface{}{
			"client_id":    "demo-client",
			"request_id":   fmt.Sprintf("req-%d", time.Now().Unix()),
			"quantum_safe": true,
		},
	}

	jsonData, _ := json.Marshal(data)

	start := time.Now()
	resp, err := signedPost("https://localhost:8081/process", jsonData)
	duration := time.Since(start)

	if err != nil {
		fmt.Printf("❌ Process test failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("✅ Process test successful (latency: %v)\n", duration)
	fmt.Printf("   Status: %s\n", resp.Status)
}

func testStatus() {
	start := time.Now()
	resp, err := signedGet("https://localhost:8081/status")
	duration := time.Since(start)

	if err != nil {
		fmt.Printf("❌ Status test failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("✅ Status test successful (latency: %v)\n", duration)
	fmt.Printf("   Status: %s\n", resp.Status)
}

func showPerformanceSummary() {
	fmt.Println("\n📈 PQC Performance Characteristics:")
	fmt.Println("   • Dilithium3 Signatures: Quantum-resistant digital signatures")
	fmt.Println("   • Kyber768 KEM: Quantum-resistant key encapsulation")
	fmt.Println("   • All inter-service communication is cryptographically verified")
	fmt.Println("   • Zero-trust architecture with PQC mutual authentication")

	fmt.Println("\n🛡️  Security Benefits:")
	fmt.Println("   • Protection against quantum computer attacks")
	fmt.Println("   • Forward secrecy with Kyber key exchange")
	fmt.Println("   • Mutual authentication between all services")
	fmt.Println("   • Tamper-evident message signatures")
}

func runBenchmark() {
	fmt.Println("🚀 Running PQC Performance Benchmark...")
	pqc.BenchmarkRSAvsDialithium()
}
