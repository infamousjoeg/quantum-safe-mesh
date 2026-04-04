package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"quantum-safe-mesh/pkg/models"
	"quantum-safe-mesh/pkg/pqc"
)

func newTestAuthService() *AuthService {
	dkp, _ := pqc.GenerateDilithiumKeyPair()
	kkp, _ := pqc.GenerateKyberKeyPair()
	return &AuthService{
		dilithiumKeyPair: dkp,
		kyberKeyPair:     kkp,
		serviceRegistry:  make(map[string][]byte),
		kyberRegistry:    make(map[string][]byte),
		serviceID:        "auth-service",
	}
}

func TestRegisterServiceWithValidSignature(t *testing.T) {
	as := newTestAuthService()

	clientKP, _ := pqc.GenerateDilithiumKeyPair()
	keyPair := models.ServiceKeyPair{
		ServiceID: "test-service",
		PublicKey: clientKP.GetPublicKeyBytes(),
	}

	// Sign the registration
	regPayload, _ := json.Marshal(keyPair)
	sig, _ := clientKP.Sign(regPayload)
	keyPair.Signature = sig

	payload, _ := json.Marshal(keyPair)
	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	as.registerService(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("signed registration returned %d, want 200. Body: %s", rr.Code, rr.Body.String())
	}

	// Verify key was stored
	as.mutex.RLock()
	_, exists := as.serviceRegistry["test-service"]
	as.mutex.RUnlock()
	if !exists {
		t.Error("service should be in registry after signed registration")
	}
}

func TestRegisterServiceRejectsUnsigned(t *testing.T) {
	as := newTestAuthService()

	keyPair := models.ServiceKeyPair{
		ServiceID: "attacker",
		PublicKey: []byte("fake-key"),
	}

	payload, _ := json.Marshal(keyPair)
	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	as.registerService(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("unsigned registration returned %d, want 401", rr.Code)
	}

	// Verify key was NOT stored
	as.mutex.RLock()
	_, exists := as.serviceRegistry["attacker"]
	as.mutex.RUnlock()
	if exists {
		t.Error("unsigned service should NOT be in registry")
	}
}

func TestRegisterServiceRejectsWrongSignature(t *testing.T) {
	as := newTestAuthService()

	realKP, _ := pqc.GenerateDilithiumKeyPair()
	attackerKP, _ := pqc.GenerateDilithiumKeyPair()

	keyPair := models.ServiceKeyPair{
		ServiceID: "victim-service",
		PublicKey: realKP.GetPublicKeyBytes(), // Real service's key
	}

	// Attacker signs with their own key (not matching PublicKey)
	regPayload, _ := json.Marshal(keyPair)
	sig, _ := attackerKP.Sign(regPayload)
	keyPair.Signature = sig

	payload, _ := json.Marshal(keyPair)
	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	as.registerService(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("registration with wrong signature returned %d, want 401", rr.Code)
	}
}
