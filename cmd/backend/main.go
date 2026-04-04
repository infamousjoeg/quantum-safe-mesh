package main

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"quantum-safe-mesh/pkg/models"
	"quantum-safe-mesh/pkg/pqc"
)

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

type BackendService struct {
	dilithiumKeyPair *pqc.DilithiumKeyPair
	kyberKeyPair     *pqc.KyberKeyPair
	serviceID        string
	authServiceURL   string
	publicKeyCache   map[string][]byte
	sharedSecrets    map[string][]byte
	mutex            sync.RWMutex
	requestCounter   int
	seenNonces       map[string]time.Time
	tlsClient        *http.Client
}

func NewBackendService() (*BackendService, error) {
	log.Println("🚀 Starting Backend Service...")

	serviceID := "backend-service"

	dilithiumKeyPair, kyberKeyPair, err := pqc.LoadKeyPair(serviceID)
	if err != nil {
		log.Printf("Keys not found, generating new ones...")
		dilithiumKeyPair, err = pqc.GenerateDilithiumKeyPair()
		if err != nil {
			return nil, fmt.Errorf("failed to generate Dilithium keypair: %w", err)
		}

		kyberKeyPair, err = pqc.GenerateKyberKeyPair()
		if err != nil {
			return nil, fmt.Errorf("failed to generate Kyber keypair: %w", err)
		}

		if err := pqc.SaveKeyPair(serviceID, dilithiumKeyPair, kyberKeyPair); err != nil {
			log.Printf("Warning: failed to save keys: %v", err)
		}
	}

	bs := &BackendService{
		dilithiumKeyPair: dilithiumKeyPair,
		kyberKeyPair:     kyberKeyPair,
		serviceID:        serviceID,
		authServiceURL:   getEnvOrDefault("AUTH_SERVICE_URL", "https://localhost:8080"),
		publicKeyCache:   make(map[string][]byte),
		sharedSecrets:    make(map[string][]byte),
		requestCounter:   0,
		seenNonces:       make(map[string]time.Time),
	}

	// Setup TLS client for auth service calls
	_, _, caCertPEM, err := pqc.GenerateCA("keys")
	if err != nil {
		return nil, fmt.Errorf("failed to load CA: %w", err)
	}
	tlsConfig := pqc.NewTLSClient(caCertPEM)
	bs.tlsClient = &http.Client{
		Transport: &http.Transport{TLSClientConfig: tlsConfig},
	}
	log.Println("✅ TLS client configured")

	if err := bs.registerWithAuthService(); err != nil {
		log.Printf("Warning: failed to register with auth service: %v", err)
	}

	log.Printf("✅ Backend Service initialized with ID: %s", serviceID)
	return bs, nil
}

func (bs *BackendService) registerWithAuthService() error {
	log.Println("📝 Registering with Auth Service...")

	keyPair := models.ServiceKeyPair{
		ServiceID:      bs.serviceID,
		PublicKey:       bs.dilithiumKeyPair.GetPublicKeyBytes(),
		KyberPublicKey: bs.kyberKeyPair.GetPublicKeyBytes(),
	}

	// Sign the registration to prove key ownership
	regPayload, err := json.Marshal(keyPair)
	if err != nil {
		return fmt.Errorf("failed to marshal registration request: %w", err)
	}

	sig, err := bs.dilithiumKeyPair.Sign(regPayload)
	if err != nil {
		return fmt.Errorf("failed to sign registration: %w", err)
	}
	keyPair.Signature = sig

	payload, err := json.Marshal(keyPair)
	if err != nil {
		return fmt.Errorf("failed to marshal signed registration request: %w", err)
	}

	resp, err := bs.tlsClient.Post(bs.authServiceURL+"/register", "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to register with auth service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("auth service registration failed with status: %d", resp.StatusCode)
	}

	log.Println("✅ Successfully registered with Auth Service")
	return nil
}

func (bs *BackendService) handleKeyExchange(w http.ResponseWriter, r *http.Request) {
	log.Println("🤝 Received key exchange request")

	var request struct {
		ServiceID  string `json:"service_id"`
		Ciphertext []byte `json:"ciphertext"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Printf("❌ Invalid key exchange request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	sharedSecret, err := bs.kyberKeyPair.Decapsulate(request.Ciphertext)
	if err != nil {
		log.Printf("❌ Failed to decapsulate: %v", err)
		http.Error(w, "Decapsulation failed", http.StatusInternalServerError)
		return
	}

	bs.mutex.Lock()
	bs.sharedSecrets[request.ServiceID] = sharedSecret
	bs.mutex.Unlock()

	log.Printf("✅ Key exchange completed with %s, shared secret: %d bytes", request.ServiceID, len(sharedSecret))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "success",
		"timestamp": time.Now(),
	})
}

func (bs *BackendService) performKeyExchangeWithAuth() error {
	log.Println("🤝 Performing key exchange with auth service")

	request := models.KeyExchangeRequest{
		ServiceID:      bs.serviceID,
		KyberPublicKey: bs.kyberKeyPair.GetPublicKeyBytes(),
		Timestamp:      time.Now(),
	}

	requestData, _ := json.Marshal(map[string]interface{}{
		"service_id":       request.ServiceID,
		"kyber_public_key": request.KyberPublicKey,
		"timestamp":        request.Timestamp,
	})

	signature, err := bs.dilithiumKeyPair.Sign(requestData)
	if err != nil {
		return fmt.Errorf("failed to sign key exchange request: %w", err)
	}

	request.Signature = signature

	payload, _ := json.Marshal(request)
	resp, err := bs.tlsClient.Post(bs.authServiceURL+"/key-exchange", "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("key exchange request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("key exchange failed with status: %d", resp.StatusCode)
	}

	var response models.KeyExchangeResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("failed to decode key exchange response: %w", err)
	}

	sharedSecret, err := bs.kyberKeyPair.Decapsulate(response.Ciphertext)
	if err != nil {
		return fmt.Errorf("failed to decapsulate shared secret: %w", err)
	}

	bs.mutex.Lock()
	bs.sharedSecrets["auth-service"] = sharedSecret
	bs.mutex.Unlock()

	log.Printf("✅ Key exchange with auth completed, shared secret: %d bytes", len(sharedSecret))
	return nil
}

func (bs *BackendService) getServicePublicKey(serviceID string) ([]byte, error) {
	bs.mutex.RLock()
	if publicKey, exists := bs.publicKeyCache[serviceID]; exists {
		bs.mutex.RUnlock()
		return publicKey, nil
	}
	bs.mutex.RUnlock()

	log.Printf("🔍 Fetching public key for service: %s", serviceID)

	resp, err := bs.tlsClient.Get(fmt.Sprintf("%s/public-key/%s", bs.authServiceURL, serviceID))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch public key: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get public key, status: %d", resp.StatusCode)
	}

	var response models.ServiceResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var keyData map[string]interface{}
	if err := json.Unmarshal(response.Data, &keyData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal key data: %w", err)
	}

	publicKeyBytes, err := pqc.DeserializePublicKeyFromJSON(keyData)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize public key: %w", err)
	}

	bs.mutex.Lock()
	bs.publicKeyCache[serviceID] = publicKeyBytes
	bs.mutex.Unlock()

	return publicKeyBytes, nil
}

func (bs *BackendService) verifyRequest(request models.ServiceRequest) (models.ServiceRequest, error) {
	log.Printf("🔍 Verifying request from service: %s", request.ServiceID)

	publicKey, err := bs.getServicePublicKey(request.ServiceID)
	if err != nil {
		return request, fmt.Errorf("failed to get public key: %w", err)
	}

	requestData := models.ServiceRequest{
		ServiceID: request.ServiceID,
		Timestamp: request.Timestamp,
		Data:      request.Data,
		Headers:   request.Headers,
		Nonce:     request.Nonce,
	}

	payload, err := json.Marshal(requestData)
	if err != nil {
		return request, fmt.Errorf("failed to marshal request for verification: %w", err)
	}

	if err := pqc.VerifyDilithiumSignature(publicKey, payload, request.Signature); err != nil {
		return request, fmt.Errorf("signature verification failed: %w", err)
	}

	// Replay protection: nonce is mandatory
	if request.Nonce == "" {
		return request, fmt.Errorf("missing nonce: replay protection requires a nonce")
	}

	bs.mutex.Lock()
	// Clean up expired nonces (older than 60s)
	for n, t := range bs.seenNonces {
		if time.Since(t) > 60*time.Second {
			delete(bs.seenNonces, n)
		}
	}
	if _, seen := bs.seenNonces[request.Nonce]; seen {
		bs.mutex.Unlock()
		return request, fmt.Errorf("replay detected: duplicate nonce %s", request.Nonce)
	}
	bs.seenNonces[request.Nonce] = time.Now()
	bs.mutex.Unlock()
	log.Printf("✅ Nonce accepted: %s", request.Nonce)

	log.Printf("✅ Request verified successfully from service: %s", request.ServiceID)

	// Validate timestamp freshness (±30s window)
	age := time.Since(request.Timestamp)
	if age < 0 {
		age = -age
	}
	if age > 30*time.Second {
		return request, fmt.Errorf("request timestamp too old or too far in future: %v", age)
	}
	log.Printf("✅ Request timestamp valid (age: %v)", age)

	bs.mutex.RLock()
	sharedSecret, hasSecret := bs.sharedSecrets[request.ServiceID]
	bs.mutex.RUnlock()

	if hasSecret {
		var encoded string
		if err := json.Unmarshal(request.Data, &encoded); err != nil {
			return request, fmt.Errorf("failed to unmarshal encrypted data: %w", err)
		}
		encryptedBytes, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			return request, fmt.Errorf("failed to decode base64 data: %w", err)
		}
		decryptedData, err := pqc.Decrypt(sharedSecret, encryptedBytes)
		if err != nil {
			return request, fmt.Errorf("failed to decrypt request data: %w", err)
		}
		request.Data = decryptedData
		log.Printf("🔓 Decrypted request data (%d bytes)", len(decryptedData))
	}

	return request, nil
}

func (bs *BackendService) encryptResponseData(serviceID string, data []byte) ([]byte, error) {
	bs.mutex.RLock()
	sharedSecret, hasSecret := bs.sharedSecrets[serviceID]
	bs.mutex.RUnlock()

	if !hasSecret {
		return data, nil
	}

	encryptedBytes, err := pqc.Encrypt(sharedSecret, data)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt response data: %w", err)
	}
	// Wrap as base64 JSON string so json.RawMessage can marshal it
	encoded := base64.StdEncoding.EncodeToString(encryptedBytes)
	jsonEncoded, _ := json.Marshal(encoded)
	log.Printf("🔒 Encrypted response data (%d bytes → %d bytes)", len(data), len(encryptedBytes))
	return jsonEncoded, nil
}

func (bs *BackendService) processEcho(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	log.Println("🔄 Processing echo request")

	var request models.ServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Printf("❌ Invalid request format: %v", err)
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	var err error
	request, err = bs.verifyRequest(request)
	if err != nil {
		log.Printf("❌ Request verification failed: %v", err)
		http.Error(w, "Request verification failed", http.StatusUnauthorized)
		return
	}

	bs.mutex.Lock()
	bs.requestCounter++
	currentCount := bs.requestCounter
	bs.mutex.Unlock()

	var requestData interface{}
	if err := json.Unmarshal(request.Data, &requestData); err != nil {
		requestData = string(request.Data)
	}

	responseData := map[string]interface{}{
		"message":         "Echo from Backend Service",
		"original_data":   requestData,
		"service_id":      bs.serviceID,
		"request_count":   currentCount,
		"timestamp":       time.Now(),
		"processing_time": time.Since(start).String(),
		"from_service":    request.ServiceID,
	}

	responsePayload, err := json.Marshal(responseData)
	if err != nil {
		log.Printf("❌ Failed to marshal response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	responsePayload, err = bs.encryptResponseData(request.ServiceID, responsePayload)
	if err != nil {
		log.Printf("❌ Failed to encrypt response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	signature, err := bs.dilithiumKeyPair.Sign(responsePayload)
	if err != nil {
		log.Printf("❌ Failed to sign response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := models.ServiceResponse{
		ServiceID: bs.serviceID,
		Timestamp: time.Now(),
		Data:      responsePayload,
		Signature: signature,
		Success:   true,
	}

	duration := time.Since(start)
	log.Printf("✅ Echo request processed successfully in %v (request #%d)", duration, currentCount)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (bs *BackendService) processData(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	log.Println("🧠 Processing data transformation request")

	var request models.ServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Printf("❌ Invalid request format: %v", err)
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	var err error
	request, err = bs.verifyRequest(request)
	if err != nil {
		log.Printf("❌ Request verification failed: %v", err)
		http.Error(w, "Request verification failed", http.StatusUnauthorized)
		return
	}

	bs.mutex.Lock()
	bs.requestCounter++
	currentCount := bs.requestCounter
	bs.mutex.Unlock()

	var inputData map[string]interface{}
	if err := json.Unmarshal(request.Data, &inputData); err != nil {
		inputData = map[string]interface{}{
			"raw_data": string(request.Data),
		}
	}

	transformedData := map[string]interface{}{
		"service_id":      bs.serviceID,
		"operation":       "data_transformation",
		"input":           inputData,
		"transformed_at":  time.Now(),
		"request_count":   currentCount,
		"processing_time": time.Since(start).String(),
		"metadata": map[string]interface{}{
			"quantum_safe": true,
			"algorithm":    "Dilithium3",
			"from_service": request.ServiceID,
		},
	}

	responsePayload, err := json.Marshal(transformedData)
	if err != nil {
		log.Printf("❌ Failed to marshal response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	responsePayload, err = bs.encryptResponseData(request.ServiceID, responsePayload)
	if err != nil {
		log.Printf("❌ Failed to encrypt response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	signature, err := bs.dilithiumKeyPair.Sign(responsePayload)
	if err != nil {
		log.Printf("❌ Failed to sign response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := models.ServiceResponse{
		ServiceID: bs.serviceID,
		Timestamp: time.Now(),
		Data:      responsePayload,
		Signature: signature,
		Success:   true,
	}

	duration := time.Since(start)
	log.Printf("✅ Data transformation completed in %v (request #%d)", duration, currentCount)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (bs *BackendService) getStatus(w http.ResponseWriter, r *http.Request) {
	log.Println("📊 Status request received")

	bs.mutex.RLock()
	currentCount := bs.requestCounter
	bs.mutex.RUnlock()

	statusData := map[string]interface{}{
		"service_id":       bs.serviceID,
		"status":           "healthy",
		"requests_handled": currentCount,
		"uptime":           time.Since(time.Now().Add(-time.Hour)).String(),
		"quantum_safe":     true,
		"algorithms": map[string]string{
			"signature": "Dilithium3",
			"kem":       "Kyber768",
		},
		"timestamp": time.Now(),
	}

	responsePayload, err := json.Marshal(statusData)
	if err != nil {
		log.Printf("❌ Failed to marshal status response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	responsePayload, err = bs.encryptResponseData("api-gateway", responsePayload)
	if err != nil {
		log.Printf("❌ Failed to encrypt response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	signature, err := bs.dilithiumKeyPair.Sign(responsePayload)
	if err != nil {
		log.Printf("❌ Failed to sign status response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := models.ServiceResponse{
		ServiceID: bs.serviceID,
		Timestamp: time.Now(),
		Data:      responsePayload,
		Signature: signature,
		Success:   true,
	}

	log.Println("✅ Status response sent")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	backendService, err := NewBackendService()
	if err != nil {
		log.Fatalf("Failed to create backend service: %v", err)
	}

	if err := backendService.performKeyExchangeWithAuth(); err != nil {
		log.Printf("Warning: key exchange with auth failed: %v", err)
	}

	r := mux.NewRouter()

	r.HandleFunc("/echo", backendService.processEcho).Methods("POST")
	r.HandleFunc("/key-exchange", backendService.handleKeyExchange).Methods("POST")
	r.HandleFunc("/process", backendService.processData).Methods("POST")
	r.HandleFunc("/status", backendService.getStatus).Methods("GET", "POST")
	r.HandleFunc("/metrics", pqc.MetricsHandler()).Methods("GET")

	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Backend OK"))
	}).Methods("GET")

	// Setup TLS server
	caCert, caKey, _, err := pqc.GenerateCA("keys")
	if err != nil {
		log.Fatalf("Failed to load CA: %v", err)
	}
	tlsCert, err := pqc.GenerateServiceCert("keys", "backend-service", caCert, caKey)
	if err != nil {
		log.Fatalf("Failed to generate TLS cert: %v", err)
	}
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
	}
	server := &http.Server{
		Addr:      ":8082",
		Handler:   r,
		TLSConfig: tlsConfig,
	}
	log.Println("🌟 Backend Service starting on :8082 (TLS)")
	log.Fatal(server.ListenAndServeTLS("", ""))
}
