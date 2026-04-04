package main

import (
	"bytes"
	crand "crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"quantum-safe-mesh/pkg/models"
	"quantum-safe-mesh/pkg/pqc"
)

func generateNonce() string {
	b := make([]byte, 16)
	crand.Read(b)
	return hex.EncodeToString(b)
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

type APIGateway struct {
	dilithiumKeyPair  *pqc.DilithiumKeyPair
	kyberKeyPair      *pqc.KyberKeyPair
	serviceID         string
	authServiceURL    string
	backendServiceURL string
	publicKeyCache    map[string][]byte
	sharedSecrets     map[string][]byte // serviceID -> 32-byte AES key
	mutex             sync.RWMutex
	tlsClient         *http.Client
}

func NewAPIGateway() (*APIGateway, error) {
	log.Println("🚀 Starting API Gateway...")

	serviceID := "api-gateway"

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

	gw := &APIGateway{
		dilithiumKeyPair:  dilithiumKeyPair,
		kyberKeyPair:      kyberKeyPair,
		serviceID:         serviceID,
		authServiceURL:    getEnvOrDefault("AUTH_SERVICE_URL", "https://localhost:8080"),
		backendServiceURL: getEnvOrDefault("BACKEND_SERVICE_URL", "https://localhost:8082"),
		publicKeyCache:    make(map[string][]byte),
		sharedSecrets:     make(map[string][]byte),
	}

	// Setup TLS client
	_, _, caCertPEM, err := pqc.GenerateCA("keys")
	if err != nil {
		return nil, fmt.Errorf("failed to load CA: %w", err)
	}
	tlsConf := pqc.NewTLSClient(caCertPEM)
	gw.tlsClient = &http.Client{
		Transport: &http.Transport{TLSClientConfig: tlsConf},
	}
	log.Println("✅ TLS client configured")

	if err := gw.registerWithAuthService(); err != nil {
		log.Printf("Warning: failed to register with auth service: %v", err)
	}

	log.Printf("✅ API Gateway initialized with ID: %s", serviceID)
	return gw, nil
}

func (gw *APIGateway) registerWithAuthService() error {
	log.Println("📝 Registering with Auth Service...")

	keyPair := models.ServiceKeyPair{
		ServiceID:      gw.serviceID,
		PublicKey:       gw.dilithiumKeyPair.GetPublicKeyBytes(),
		KyberPublicKey: gw.kyberKeyPair.GetPublicKeyBytes(),
	}

	// Sign the registration to prove key ownership
	regPayload, err := json.Marshal(keyPair)
	if err != nil {
		return fmt.Errorf("failed to marshal registration request: %w", err)
	}

	sig, err := gw.dilithiumKeyPair.Sign(regPayload)
	if err != nil {
		return fmt.Errorf("failed to sign registration: %w", err)
	}
	keyPair.Signature = sig

	payload, err := json.Marshal(keyPair)
	if err != nil {
		return fmt.Errorf("failed to marshal signed registration request: %w", err)
	}

	resp, err := gw.tlsClient.Post(gw.authServiceURL+"/register", "application/json", bytes.NewBuffer(payload))
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

func (gw *APIGateway) getServicePublicKey(serviceID string) ([]byte, error) {
	gw.mutex.RLock()
	if publicKey, exists := gw.publicKeyCache[serviceID]; exists {
		gw.mutex.RUnlock()
		return publicKey, nil
	}
	gw.mutex.RUnlock()

	log.Printf("🔍 Fetching public key for service: %s", serviceID)

	resp, err := gw.tlsClient.Get(fmt.Sprintf("%s/public-key/%s", gw.authServiceURL, serviceID))
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

	gw.mutex.Lock()
	gw.publicKeyCache[serviceID] = publicKeyBytes
	gw.mutex.Unlock()

	return publicKeyBytes, nil
}

func (gw *APIGateway) getServiceKyberPublicKey(serviceID string) ([]byte, error) {
	log.Printf("🔍 Fetching Kyber public key for service: %s", serviceID)

	resp, err := gw.tlsClient.Get(fmt.Sprintf("%s/public-kyber-key/%s", gw.authServiceURL, serviceID))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Kyber public key: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get Kyber public key, status: %d", resp.StatusCode)
	}

	var response models.ServiceResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var keyData map[string]interface{}
	if err := json.Unmarshal(response.Data, &keyData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal key data: %w", err)
	}

	kyberKeyRaw, exists := keyData["kyber_public_key"]
	if !exists {
		return nil, fmt.Errorf("kyber_public_key field not found")
	}

	if keyStr, ok := kyberKeyRaw.(string); ok {
		return base64.StdEncoding.DecodeString(keyStr)
	}

	return nil, fmt.Errorf("unexpected kyber_public_key type: %T", kyberKeyRaw)
}

func (gw *APIGateway) performPairwiseKeyExchange(serviceID string) error {
	log.Printf("🤝 Performing pairwise key exchange with %s", serviceID)

	kyberPubKey, err := gw.getServiceKyberPublicKey(serviceID)
	if err != nil {
		return fmt.Errorf("failed to get Kyber public key: %w", err)
	}

	ciphertext, sharedSecret, err := pqc.EncapsulateWithPublicKey(kyberPubKey)
	if err != nil {
		return fmt.Errorf("failed to encapsulate: %w", err)
	}

	exchangeReq := map[string]interface{}{
		"service_id": gw.serviceID,
		"ciphertext": ciphertext,
	}
	payload, _ := json.Marshal(exchangeReq)

	backendURL := gw.backendServiceURL + "/key-exchange"
	resp, err := gw.tlsClient.Post(backendURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("key exchange request to %s failed: %w", serviceID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("key exchange with %s failed with status: %d", serviceID, resp.StatusCode)
	}

	gw.mutex.Lock()
	gw.sharedSecrets[serviceID] = sharedSecret
	gw.mutex.Unlock()

	log.Printf("✅ Pairwise key exchange completed with %s, shared secret: %d bytes", serviceID, len(sharedSecret))
	return nil
}

func (gw *APIGateway) verifyRequest(r *http.Request, serviceID string) error {
	log.Printf("🔍 Verifying request from service: %s", serviceID)

	signature := r.Header.Get("X-Signature")
	if signature == "" {
		return fmt.Errorf("missing signature header")
	}

	publicKey, err := gw.getServicePublicKey(serviceID)
	if err != nil {
		return fmt.Errorf("failed to get public key: %w", err)
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("failed to read request body: %w", err)
	}
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	signatureBytes, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return fmt.Errorf("invalid base64 signature: %w", err)
	}
	if err := pqc.VerifyDilithiumSignature(publicKey, body, signatureBytes); err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}

	log.Printf("✅ Request verified successfully from service: %s", serviceID)
	return nil
}

func (gw *APIGateway) forwardToBackend(w http.ResponseWriter, r *http.Request) {
	log.Println("🔄 Forwarding request to backend service")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("❌ Failed to read request body: %v", err)
		http.Error(w, "Failed to read request", http.StatusInternalServerError)
		return
	}

	// Handle empty request bodies for GET requests
	var requestBody json.RawMessage
	if len(body) == 0 {
		// For empty bodies (like GET requests), use empty JSON object
		requestBody = json.RawMessage("{}")
	} else {
		requestBody = json.RawMessage(body)
	}

	// Encrypt the request data if we have a shared secret
	gw.mutex.RLock()
	sharedSecret, hasSecret := gw.sharedSecrets["backend-service"]
	gw.mutex.RUnlock()

	if hasSecret {
		encryptedBytes, err := pqc.Encrypt(sharedSecret, requestBody)
		if err != nil {
			log.Printf("❌ Failed to encrypt request data: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		// Wrap encrypted bytes as a JSON base64 string so json.RawMessage can marshal it
		encoded := base64.StdEncoding.EncodeToString(encryptedBytes)
		requestBody, _ = json.Marshal(encoded)
		log.Printf("🔒 Encrypted request data (%d bytes)", len(encryptedBytes))
	}

	requestData := models.ServiceRequest{
		ServiceID: gw.serviceID,
		Timestamp: time.Now(),
		Nonce:     generateNonce(),
		Data:      requestBody,
		Headers:   make(map[string]string),
	}

	for key, values := range r.Header {
		if len(values) > 0 {
			requestData.Headers[key] = values[0]
		}
	}

	requestPayload, err := json.Marshal(requestData)
	if err != nil {
		log.Printf("❌ Failed to marshal request: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	signature, err := gw.dilithiumKeyPair.Sign(requestPayload)
	if err != nil {
		log.Printf("❌ Failed to sign request: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	requestData.Signature = signature
	signedPayload, _ := json.Marshal(requestData)

	req, err := http.NewRequest("POST", gw.backendServiceURL+r.URL.Path, bytes.NewBuffer(signedPayload))
	if err != nil {
		log.Printf("❌ Failed to create request: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-ID", gw.serviceID)

	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: gw.tlsClient.Transport,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("❌ Backend request failed: %v", err)
		http.Error(w, "Backend service unavailable", http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	var backendResponse models.ServiceResponse
	if err := json.NewDecoder(resp.Body).Decode(&backendResponse); err != nil {
		log.Printf("❌ Failed to decode backend response: %v", err)
		http.Error(w, "Invalid backend response", http.StatusInternalServerError)
		return
	}

	backendPublicKey, err := gw.getServicePublicKey("backend-service")
	if err != nil {
		log.Printf("❌ Failed to get backend public key: %v", err)
		http.Error(w, "Authentication error", http.StatusInternalServerError)
		return
	}

	if err := pqc.VerifyDilithiumSignature(backendPublicKey, backendResponse.Data, backendResponse.Signature); err != nil {
		log.Printf("❌ Backend response signature verification failed: %v", err)
		http.Error(w, "Invalid backend signature", http.StatusUnauthorized)
		return
	}

	log.Println("✅ Backend response verified successfully")

	// Decrypt response Data if we have a shared secret
	if hasSecret {
		var encoded string
		if err := json.Unmarshal(backendResponse.Data, &encoded); err != nil {
			log.Printf("❌ Failed to unmarshal encrypted response: %v", err)
			http.Error(w, "Decryption error", http.StatusInternalServerError)
			return
		}
		encryptedBytes, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			log.Printf("❌ Failed to decode base64 response: %v", err)
			http.Error(w, "Decryption error", http.StatusInternalServerError)
			return
		}
		decryptedData, err := pqc.Decrypt(sharedSecret, encryptedBytes)
		if err != nil {
			log.Printf("❌ Failed to decrypt backend response: %v", err)
			http.Error(w, "Decryption error", http.StatusInternalServerError)
			return
		}
		backendResponse.Data = decryptedData
		log.Printf("🔓 Decrypted response data (%d bytes)", len(decryptedData))
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Gateway-Service", gw.serviceID)
	w.WriteHeader(resp.StatusCode)

	json.NewEncoder(w).Encode(backendResponse)
}

func (gw *APIGateway) handleRequest(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	log.Printf("🌐 Gateway received %s request to %s", r.Method, r.URL.Path)

	clientID := r.Header.Get("X-Client-ID")
	if clientID == "" {
		clientID = "unknown-client"
	}

	if r.URL.Path == "/health" {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Gateway OK"))
		return
	}

	if err := gw.verifyRequest(r, clientID); err != nil {
		log.Printf("❌ Client request verification failed: %v", err)
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	gw.forwardToBackend(w, r)

	duration := time.Since(start)
	log.Printf("⏱️  Request processed in %v", duration)
}

func (gw *APIGateway) performKeyExchange() error {
	log.Println("🤝 Performing key exchange with backend service")

	request := models.KeyExchangeRequest{
		ServiceID:      gw.serviceID,
		KyberPublicKey: gw.kyberKeyPair.GetPublicKeyBytes(),
		Timestamp:      time.Now(),
	}

	requestData, _ := json.Marshal(map[string]interface{}{
		"service_id":       request.ServiceID,
		"kyber_public_key": request.KyberPublicKey,
		"timestamp":        request.Timestamp,
	})

	signature, err := gw.dilithiumKeyPair.Sign(requestData)
	if err != nil {
		return fmt.Errorf("failed to sign key exchange request: %w", err)
	}

	request.Signature = signature

	payload, _ := json.Marshal(request)
	resp, err := gw.tlsClient.Post(gw.authServiceURL+"/key-exchange", "application/json", bytes.NewBuffer(payload))
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

	sharedSecret, err := gw.kyberKeyPair.Decapsulate(response.Ciphertext)
	if err != nil {
		return fmt.Errorf("failed to decapsulate shared secret: %w", err)
	}

	log.Printf("✅ Key exchange completed, shared secret size: %d bytes", len(sharedSecret))
	return nil
}

func (gw *APIGateway) startKeyRotation(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			log.Println("🔄 Rotating shared secret with backend-service...")
			if err := gw.performPairwiseKeyExchange("backend-service"); err != nil {
				log.Printf("❌ Key rotation failed: %v", err)
			} else {
				log.Println("✅ Key rotation completed successfully")
			}
		}
	}()
}

func main() {
	gateway, err := NewAPIGateway()
	if err != nil {
		log.Fatalf("Failed to create API gateway: %v", err)
	}

	if err := gateway.performKeyExchange(); err != nil {
		log.Printf("Warning: key exchange failed: %v", err)
	}

	if err := gateway.performPairwiseKeyExchange("backend-service"); err != nil {
		log.Printf("Warning: pairwise key exchange with backend failed: %v", err)
	}

	// Start periodic key rotation (every 1 hour)
	gateway.startKeyRotation(1 * time.Hour)
	log.Println("✅ Periodic key rotation enabled (1 hour interval)")

	r := mux.NewRouter()
	r.HandleFunc("/metrics", pqc.MetricsHandler()).Methods("GET")
	r.PathPrefix("/").HandlerFunc(gateway.handleRequest)

	caCert, caKey, _, err := pqc.GenerateCA("keys")
	if err != nil {
		log.Fatalf("Failed to load CA: %v", err)
	}
	tlsCert, err := pqc.GenerateServiceCert("keys", "api-gateway", caCert, caKey)
	if err != nil {
		log.Fatalf("Failed to generate TLS cert: %v", err)
	}
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
	}
	server := &http.Server{
		Addr:      ":8081",
		Handler:   r,
		TLSConfig: tlsConfig,
	}
	log.Println("🌟 API Gateway starting on :8081 (TLS)")
	log.Fatal(server.ListenAndServeTLS("", ""))
}
