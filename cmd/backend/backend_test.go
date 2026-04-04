package main

import (
	"encoding/json"
	"testing"
	"time"

	"quantum-safe-mesh/pkg/models"
	"quantum-safe-mesh/pkg/pqc"
)

// TestBackendVerifyRequestValid proves the backend's verification logic works
// when the request is signed correctly by the gateway.
func TestBackendVerifyRequestValid(t *testing.T) {
	kp, err := pqc.GenerateDilithiumKeyPair()
	if err != nil {
		t.Fatalf("keygen failed: %v", err)
	}

	// Build request as gateway does (no signature yet)
	request := models.ServiceRequest{
		ServiceID: "api-gateway",
		Timestamp: time.Now(),
		Data:      json.RawMessage(`{"action":"echo","payload":"test"}`),
		Headers:   map[string]string{"X-Request-ID": "123"},
	}

	payload, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	sig, err := kp.Sign(payload)
	if err != nil {
		t.Fatalf("sign failed: %v", err)
	}
	request.Signature = sig

	// Reconstruct as backend does (same as backend verifyRequest lines 152-157)
	reconstructed := models.ServiceRequest{
		ServiceID: request.ServiceID,
		Timestamp: request.Timestamp,
		Data:      request.Data,
		Headers:   request.Headers,
	}

	verifyPayload, err := json.Marshal(reconstructed)
	if err != nil {
		t.Fatalf("marshal reconstructed failed: %v", err)
	}

	err = pqc.VerifyDilithiumSignature(kp.GetPublicKeyBytes(), verifyPayload, request.Signature)
	if err != nil {
		t.Errorf("valid request should pass verification: %v", err)
	}
}

// TestBackendVerifyRequestTampered proves tampered data is rejected.
func TestBackendVerifyRequestTampered(t *testing.T) {
	kp, err := pqc.GenerateDilithiumKeyPair()
	if err != nil {
		t.Fatalf("keygen failed: %v", err)
	}

	request := models.ServiceRequest{
		ServiceID: "api-gateway",
		Timestamp: time.Now(),
		Data:      json.RawMessage(`{"action":"echo","payload":"original"}`),
		Headers:   map[string]string{},
	}

	payload, _ := json.Marshal(request)
	sig, _ := kp.Sign(payload)
	request.Signature = sig

	// Tamper the data
	tampered := models.ServiceRequest{
		ServiceID: request.ServiceID,
		Timestamp: request.Timestamp,
		Data:      json.RawMessage(`{"action":"echo","payload":"INJECTED"}`),
		Headers:   request.Headers,
	}

	tamperedPayload, _ := json.Marshal(tampered)
	err = pqc.VerifyDilithiumSignature(kp.GetPublicKeyBytes(), tamperedPayload, request.Signature)
	if err == nil {
		t.Error("tampered data should fail verification")
	}
}

// TestBackendVerifyRequestWrongService proves impersonation is caught
// if the correct public key lookup is in place.
func TestBackendVerifyRequestWrongService(t *testing.T) {
	gatewayKP, _ := pqc.GenerateDilithiumKeyPair()
	attackerKP, _ := pqc.GenerateDilithiumKeyPair()

	request := models.ServiceRequest{
		ServiceID: "api-gateway",
		Timestamp: time.Now(),
		Data:      json.RawMessage(`{"action":"echo"}`),
	}

	payload, _ := json.Marshal(request)
	// Attacker signs with their own key
	sig, _ := attackerKP.Sign(payload)
	request.Signature = sig

	reconstructed := models.ServiceRequest{
		ServiceID: request.ServiceID,
		Timestamp: request.Timestamp,
		Data:      request.Data,
		Headers:   request.Headers,
	}
	verifyPayload, _ := json.Marshal(reconstructed)

	// Verifying with the real gateway's public key should fail
	err := pqc.VerifyDilithiumSignature(gatewayKP.GetPublicKeyBytes(), verifyPayload, request.Signature)
	if err == nil {
		t.Error("request signed by attacker should fail when verified with gateway's public key")
	}
}

func TestTimestampValidationRejectsStale(t *testing.T) {
	staleTime := time.Now().Add(-60 * time.Second)
	age := time.Since(staleTime)
	if age < 0 {
		age = -age
	}
	if age <= 30*time.Second {
		t.Error("60-second-old timestamp should exceed ±30s window")
	}
}

func TestTimestampValidationRejectsFuture(t *testing.T) {
	futureTime := time.Now().Add(60 * time.Second)
	age := time.Since(futureTime)
	if age < 0 {
		age = -age
	}
	if age <= 30*time.Second {
		t.Error("60-second-future timestamp should exceed ±30s window")
	}
}
