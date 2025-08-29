package main

import (
	"bytes"
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
	mutex            sync.RWMutex
	requestCounter   int
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
		authServiceURL:   getEnvOrDefault("AUTH_SERVICE_URL", "http://localhost:8080"),
		publicKeyCache:   make(map[string][]byte),
		requestCounter:   0,
	}

	if err := bs.registerWithAuthService(); err != nil {
		log.Printf("Warning: failed to register with auth service: %v", err)
	}

	log.Printf("✅ Backend Service initialized with ID: %s", serviceID)
	return bs, nil
}

func (bs *BackendService) registerWithAuthService() error {
	log.Println("📝 Registering with Auth Service...")

	keyPair := models.ServiceKeyPair{
		ServiceID: bs.serviceID,
		PublicKey: bs.dilithiumKeyPair.GetPublicKeyBytes(),
	}

	payload, err := json.Marshal(keyPair)
	if err != nil {
		return fmt.Errorf("failed to marshal registration request: %w", err)
	}

	resp, err := http.Post(bs.authServiceURL+"/register", "application/json", bytes.NewBuffer(payload))
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

func (bs *BackendService) getServicePublicKey(serviceID string) ([]byte, error) {
	bs.mutex.RLock()
	if publicKey, exists := bs.publicKeyCache[serviceID]; exists {
		bs.mutex.RUnlock()
		return publicKey, nil
	}
	bs.mutex.RUnlock()

	log.Printf("🔍 Fetching public key for service: %s", serviceID)

	resp, err := http.Get(fmt.Sprintf("%s/public-key/%s", bs.authServiceURL, serviceID))
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

	publicKeyBytes, ok := keyData["public_key"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid public key format")
	}

	pubKey := make([]byte, len(publicKeyBytes))
	for i, v := range publicKeyBytes {
		if b, ok := v.(float64); ok {
			pubKey[i] = byte(b)
		} else {
			return nil, fmt.Errorf("invalid public key byte format")
		}
	}

	bs.mutex.Lock()
	bs.publicKeyCache[serviceID] = pubKey
	bs.mutex.Unlock()

	return pubKey, nil
}

func (bs *BackendService) verifyRequest(request models.ServiceRequest) error {
	log.Printf("🔍 Verifying request from service: %s", request.ServiceID)

	publicKey, err := bs.getServicePublicKey(request.ServiceID)
	if err != nil {
		return fmt.Errorf("failed to get public key: %w", err)
	}

	requestData := models.ServiceRequest{
		ServiceID: request.ServiceID,
		Timestamp: request.Timestamp,
		Data:      request.Data,
		Headers:   request.Headers,
	}

	payload, err := json.Marshal(requestData)
	if err != nil {
		return fmt.Errorf("failed to marshal request for verification: %w", err)
	}

	if err := pqc.VerifyDilithiumSignature(publicKey, payload, request.Signature); err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}

	log.Printf("✅ Request verified successfully from service: %s", request.ServiceID)
	return nil
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

	if err := bs.verifyRequest(request); err != nil {
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

	if err := bs.verifyRequest(request); err != nil {
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

	r := mux.NewRouter()

	r.HandleFunc("/echo", backendService.processEcho).Methods("POST")
	r.HandleFunc("/process", backendService.processData).Methods("POST")
	r.HandleFunc("/status", backendService.getStatus).Methods("GET", "POST")

	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Backend OK"))
	}).Methods("GET")

	log.Println("🌟 Backend Service starting on :8082")
	log.Fatal(http.ListenAndServe(":8082", r))
}
