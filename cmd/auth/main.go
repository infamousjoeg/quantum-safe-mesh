package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"quantum-safe-mesh/pkg/models"
	"quantum-safe-mesh/pkg/pqc"
)

type AuthService struct {
	dilithiumKeyPair *pqc.DilithiumKeyPair
	kyberKeyPair     *pqc.KyberKeyPair
	serviceRegistry  map[string][]byte // serviceID -> dilithium public key
	kyberRegistry    map[string][]byte // serviceID -> kyber public key
	sharedSecrets    map[string][]byte // serviceID -> 32-byte AES key from KEM
	mutex            sync.RWMutex
	serviceID        string
}

func NewAuthService() (*AuthService, error) {
	log.Println("🚀 Starting Auth Service...")

	serviceID := "auth-service"

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

	as := &AuthService{
		dilithiumKeyPair: dilithiumKeyPair,
		kyberKeyPair:     kyberKeyPair,
		serviceRegistry:  make(map[string][]byte),
		kyberRegistry:    make(map[string][]byte),
		sharedSecrets:    make(map[string][]byte),
		serviceID:        serviceID,
	}

	as.serviceRegistry[serviceID] = dilithiumKeyPair.GetPublicKeyBytes()

	log.Printf("✅ Auth Service initialized with ID: %s", serviceID)
	return as, nil
}

func (as *AuthService) registerService(w http.ResponseWriter, r *http.Request) {
	log.Println("📝 Received service registration request")

	var keyPair models.ServiceKeyPair
	if err := json.NewDecoder(r.Body).Decode(&keyPair); err != nil {
		log.Printf("❌ Invalid registration request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Verify the registration is signed by the owner of the public key
	if len(keyPair.Signature) == 0 {
		log.Printf("❌ Registration rejected: missing signature from %s", keyPair.ServiceID)
		http.Error(w, "Registration must be signed", http.StatusUnauthorized)
		return
	}

	// Reconstruct payload without signature for verification
	verifyData := models.ServiceKeyPair{
		ServiceID:      keyPair.ServiceID,
		PublicKey:       keyPair.PublicKey,
		KyberPublicKey: keyPair.KyberPublicKey,
	}
	verifyPayload, _ := json.Marshal(verifyData)

	if err := pqc.VerifyDilithiumSignature(keyPair.PublicKey, verifyPayload, keyPair.Signature); err != nil {
		log.Printf("❌ Registration signature invalid from %s: %v", keyPair.ServiceID, err)
		http.Error(w, "Invalid registration signature", http.StatusUnauthorized)
		return
	}

	log.Printf("✅ Registration signature verified for %s", keyPair.ServiceID)

	as.mutex.Lock()
	as.serviceRegistry[keyPair.ServiceID] = keyPair.PublicKey
	if len(keyPair.KyberPublicKey) > 0 {
		as.kyberRegistry[keyPair.ServiceID] = keyPair.KyberPublicKey
	}
	as.mutex.Unlock()

	log.Printf("✅ Service registered: %s (public key size: %d bytes)",
		keyPair.ServiceID, len(keyPair.PublicKey))

	response := map[string]interface{}{
		"status":     "success",
		"service_id": keyPair.ServiceID,
		"message":    "Service registered successfully",
		"timestamp":  time.Now(),
	}

	responseData, _ := json.Marshal(response)
	signature, err := as.dilithiumKeyPair.Sign(responseData)
	if err != nil {
		log.Printf("❌ Failed to sign response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	signedResponse := models.ServiceResponse{
		ServiceID: as.serviceID,
		Timestamp: time.Now(),
		Data:      responseData,
		Signature: signature,
		Success:   true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(signedResponse)
}

func (as *AuthService) getPublicKey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceID := vars["serviceID"]

	log.Printf("🔍 Public key request for service: %s", serviceID)

	as.mutex.RLock()
	publicKey, exists := as.serviceRegistry[serviceID]
	as.mutex.RUnlock()

	if !exists {
		log.Printf("❌ Service not found: %s", serviceID)
		http.Error(w, "Service not found", http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"service_id": serviceID,
		"public_key": publicKey,
		"timestamp":  time.Now(),
	}

	responseData, _ := json.Marshal(response)
	signature, err := as.dilithiumKeyPair.Sign(responseData)
	if err != nil {
		log.Printf("❌ Failed to sign response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	signedResponse := models.ServiceResponse{
		ServiceID: as.serviceID,
		Timestamp: time.Now(),
		Data:      responseData,
		Signature: signature,
		Success:   true,
	}

	log.Printf("✅ Public key provided for service: %s", serviceID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(signedResponse)
}

func (as *AuthService) getKyberPublicKey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceID := vars["serviceID"]

	log.Printf("🔍 Kyber public key request for service: %s", serviceID)

	as.mutex.RLock()
	kyberPublicKey, exists := as.kyberRegistry[serviceID]
	as.mutex.RUnlock()

	if !exists {
		log.Printf("❌ Kyber public key not found for service: %s", serviceID)
		http.Error(w, "Kyber public key not found", http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"service_id":       serviceID,
		"kyber_public_key": kyberPublicKey,
		"timestamp":        time.Now(),
	}

	responseData, _ := json.Marshal(response)
	signature, err := as.dilithiumKeyPair.Sign(responseData)
	if err != nil {
		log.Printf("❌ Failed to sign response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	signedResponse := models.ServiceResponse{
		ServiceID: as.serviceID,
		Timestamp: time.Now(),
		Data:      responseData,
		Signature: signature,
		Success:   true,
	}

	log.Printf("✅ Kyber public key provided for service: %s", serviceID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(signedResponse)
}

func (as *AuthService) keyExchange(w http.ResponseWriter, r *http.Request) {
	log.Println("🔐 Received key exchange request")

	var request models.KeyExchangeRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Printf("❌ Invalid key exchange request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	as.mutex.RLock()
	servicePublicKey, exists := as.serviceRegistry[request.ServiceID]
	as.mutex.RUnlock()

	if !exists {
		log.Printf("❌ Service not registered: %s", request.ServiceID)
		http.Error(w, "Service not registered", http.StatusUnauthorized)
		return
	}

	requestData, _ := json.Marshal(map[string]interface{}{
		"service_id":       request.ServiceID,
		"kyber_public_key": request.KyberPublicKey,
		"timestamp":        request.Timestamp,
	})

	if err := pqc.VerifyDilithiumSignature(servicePublicKey, requestData, request.Signature); err != nil {
		log.Printf("❌ Invalid signature from service %s: %v", request.ServiceID, err)
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}

	// Validate timestamp freshness (±30s window)
	age := time.Since(request.Timestamp)
	if age < 0 {
		age = -age
	}
	if age > 30*time.Second {
		log.Printf("❌ Key exchange request timestamp too old: %v", age)
		http.Error(w, "Request timestamp expired", http.StatusUnauthorized)
		return
	}
	log.Printf("✅ Key exchange timestamp valid (age: %v)", age)

	ciphertext, sharedSecret, err := pqc.EncapsulateWithPublicKey(request.KyberPublicKey)
	if err != nil {
		log.Printf("❌ Failed to encapsulate: %v", err)
		http.Error(w, "Encapsulation failed", http.StatusInternalServerError)
		return
	}

	response := models.KeyExchangeResponse{
		Ciphertext: ciphertext,
		Timestamp:  time.Now(),
	}

	responseData, _ := json.Marshal(response)
	signature, err := as.dilithiumKeyPair.Sign(responseData)
	if err != nil {
		log.Printf("❌ Failed to sign key exchange response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response.Signature = signature

	// Store the shared secret for potential future encrypted communication
	as.mutex.Lock()
	as.sharedSecrets[request.ServiceID] = sharedSecret
	as.mutex.Unlock()

	log.Printf("✅ Key exchange completed with service: %s (shared secret: %d bytes)", request.ServiceID, len(sharedSecret))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (as *AuthService) listServices(w http.ResponseWriter, r *http.Request) {
	log.Println("📋 Listing registered services")

	as.mutex.RLock()
	services := make([]string, 0, len(as.serviceRegistry))
	for serviceID := range as.serviceRegistry {
		services = append(services, serviceID)
	}
	as.mutex.RUnlock()

	response := map[string]interface{}{
		"services":  services,
		"count":     len(services),
		"timestamp": time.Now(),
	}

	responseData, _ := json.Marshal(response)
	signature, err := as.dilithiumKeyPair.Sign(responseData)
	if err != nil {
		log.Printf("❌ Failed to sign response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	signedResponse := models.ServiceResponse{
		ServiceID: as.serviceID,
		Timestamp: time.Now(),
		Data:      responseData,
		Signature: signature,
		Success:   true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(signedResponse)
}

func main() {
	authService, err := NewAuthService()
	if err != nil {
		log.Fatalf("Failed to create auth service: %v", err)
	}

	// Generate TLS certificates
	caCert, caKey, _, err := pqc.GenerateCA("keys")
	if err != nil {
		log.Fatalf("Failed to generate CA: %v", err)
	}
	tlsCert, err := pqc.GenerateServiceCert("keys", "auth-service", caCert, caKey)
	if err != nil {
		log.Fatalf("Failed to generate TLS cert: %v", err)
	}
	log.Println("✅ TLS certificates ready")

	r := mux.NewRouter()

	r.HandleFunc("/register", authService.registerService).Methods("POST")
	r.HandleFunc("/public-key/{serviceID}", authService.getPublicKey).Methods("GET")
	r.HandleFunc("/public-kyber-key/{serviceID}", authService.getKyberPublicKey).Methods("GET")
	r.HandleFunc("/key-exchange", authService.keyExchange).Methods("POST")
	r.HandleFunc("/services", authService.listServices).Methods("GET")
	r.HandleFunc("/metrics", pqc.MetricsHandler()).Methods("GET")

	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
	}
	server := &http.Server{
		Addr:      ":8080",
		Handler:   r,
		TLSConfig: tlsConfig,
	}
	log.Println("🌟 Auth Service starting on :8080 (TLS)")
	log.Fatal(server.ListenAndServeTLS("", ""))
}
