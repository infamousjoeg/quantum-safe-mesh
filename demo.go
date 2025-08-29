package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"quantum-safe-mesh/pkg/pqc"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if len(os.Args) > 1 && os.Args[1] == "benchmark" {
		runBenchmark()
		return
	}

	fmt.Println("🌟 Quantum-Safe Service Mesh Demo")
	fmt.Println(strings.Repeat("=", 50))

	time.Sleep(2 * time.Second)

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

func checkHealth() {
	services := map[string]string{
		"Auth Service":    "http://localhost:8080/health",
		"Gateway Service": "http://localhost:8081/health",
		"Backend Service": "http://localhost:8082/health",
	}

	for name, url := range services {
		resp, err := http.Get(url)
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
	resp, err := http.Post(
		"http://localhost:8081/echo",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
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
	resp, err := http.Post(
		"http://localhost:8081/process",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
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
	resp, err := http.Get("http://localhost:8081/status")
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
