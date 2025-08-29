package main

import (
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

	as.mutex.Lock()
	as.serviceRegistry[keyPair.ServiceID] = keyPair.PublicKey
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

	ciphertext, _, err := pqc.EncapsulateWithPublicKey(request.KyberPublicKey)
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

	log.Printf("✅ Key exchange completed with service: %s", request.ServiceID)

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
	pqc.BenchmarkRSAvsDialithium()

	authService, err := NewAuthService()
	if err != nil {
		log.Fatalf("Failed to create auth service: %v", err)
	}

	r := mux.NewRouter()

	r.HandleFunc("/register", authService.registerService).Methods("POST")
	r.HandleFunc("/public-key/{serviceID}", authService.getPublicKey).Methods("GET")
	r.HandleFunc("/key-exchange", authService.keyExchange).Methods("POST")
	r.HandleFunc("/services", authService.listServices).Methods("GET")

	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	log.Println("🌟 Auth Service starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
