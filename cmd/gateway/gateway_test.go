package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"quantum-safe-mesh/pkg/models"
	"quantum-safe-mesh/pkg/pqc"
)

// TestGatewaySignatureVerifiableByBackend proves the active code path works:
// Gateway constructs a ServiceRequest with nil Signature, marshals it, signs,
// then sets Signature. Backend reconstructs without Signature and verifies.
func TestGatewaySignatureVerifiableByBackend(t *testing.T) {
	gatewayKP, err := pqc.GenerateDilithiumKeyPair()
	if err != nil {
		t.Fatalf("keygen failed: %v", err)
	}

	// Gateway builds request (same as forwardToBackend in main.go:192-212)
	requestData := models.ServiceRequest{
		ServiceID: "api-gateway",
		Timestamp: time.Now(),
		Data:      json.RawMessage(`{"message":"hello"}`),
		Headers:   map[string]string{"Content-Type": "application/json"},
	}

	// Marshal WITHOUT signature (Signature is nil/zero-value)
	payload, err := json.Marshal(requestData)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	// Sign the payload
	sig, err := gatewayKP.Sign(payload)
	if err != nil {
		t.Fatalf("sign failed: %v", err)
	}

	// Set signature (gateway sends this to backend)
	requestData.Signature = sig

	// --- Backend side ---
	// Backend reconstructs request without Signature (same as verifyRequest in backend main.go:152-157)
	reconstructed := models.ServiceRequest{
		ServiceID: requestData.ServiceID,
		Timestamp: requestData.Timestamp,
		Data:      requestData.Data,
		Headers:   requestData.Headers,
	}

	verifyPayload, err := json.Marshal(reconstructed)
	if err != nil {
		t.Fatalf("marshal reconstructed failed: %v", err)
	}

	// Verify: this proves the gateway→backend signature path works
	err = pqc.VerifyDilithiumSignature(gatewayKP.GetPublicKeyBytes(), verifyPayload, requestData.Signature)
	if err != nil {
		t.Errorf("backend verification of gateway signature failed: %v", err)
	}
}

// TestGatewaySignatureDetectsTampering proves the signature catches modifications.
func TestGatewaySignatureDetectsTampering(t *testing.T) {
	gatewayKP, err := pqc.GenerateDilithiumKeyPair()
	if err != nil {
		t.Fatalf("keygen failed: %v", err)
	}

	requestData := models.ServiceRequest{
		ServiceID: "api-gateway",
		Timestamp: time.Now(),
		Data:      json.RawMessage(`{"message":"hello"}`),
		Headers:   map[string]string{},
	}

	payload, _ := json.Marshal(requestData)
	sig, _ := gatewayKP.Sign(payload)
	requestData.Signature = sig

	// Tamper: change the data
	tampered := models.ServiceRequest{
		ServiceID: requestData.ServiceID,
		Timestamp: requestData.Timestamp,
		Data:      json.RawMessage(`{"message":"TAMPERED"}`),
		Headers:   requestData.Headers,
	}

	tamperedPayload, _ := json.Marshal(tampered)
	err = pqc.VerifyDilithiumSignature(gatewayKP.GetPublicKeyBytes(), tamperedPayload, requestData.Signature)
	if err == nil {
		t.Error("tampered request should fail signature verification")
	}
}

// TestVerifyRequestBase64Bug demonstrates the dormant bug in gateway's verifyRequest:
// []byte(base64string) produces the UTF-8 bytes of the base64 text, not the decoded bytes.
func TestVerifyRequestBase64Bug(t *testing.T) {
	kp, err := pqc.GenerateDilithiumKeyPair()
	if err != nil {
		t.Fatalf("keygen failed: %v", err)
	}

	data := []byte("test data")
	sig, _ := kp.Sign(data)

	// This is what a real HTTP header would contain: base64-encoded signature
	base64Sig := base64.StdEncoding.EncodeToString(sig)

	// BUG: gateway line 164 does []byte(signature) which gives UTF-8 bytes of the base64 string
	buggyBytes := []byte(base64Sig)

	// Correct: decode the base64 to get actual signature bytes
	correctBytes, _ := base64.StdEncoding.DecodeString(base64Sig)

	// The buggy conversion produces wrong bytes
	if len(buggyBytes) == len(correctBytes) {
		t.Error("base64 string bytes should differ in length from decoded bytes")
	}

	// Verification with buggy bytes fails
	err = pqc.VerifyDilithiumSignature(kp.GetPublicKeyBytes(), data, buggyBytes)
	if err == nil {
		t.Error("verification with []byte(base64string) should fail — this is the dormant bug")
	}

	// Verification with correct bytes succeeds
	err = pqc.VerifyDilithiumSignature(kp.GetPublicKeyBytes(), data, correctBytes)
	if err != nil {
		t.Errorf("verification with properly decoded base64 should succeed: %v", err)
	}
}

// newTestGateway creates a gateway with a pre-populated public key cache for testing.
func newTestGateway(clientID string, clientPubKey []byte) *APIGateway {
	dkp, _ := pqc.GenerateDilithiumKeyPair()
	kkp, _ := pqc.GenerateKyberKeyPair()
	return &APIGateway{
		dilithiumKeyPair:  dkp,
		kyberKeyPair:      kkp,
		serviceID:         "api-gateway",
		publicKeyCache:    map[string][]byte{clientID: clientPubKey},
		sharedSecrets:     make(map[string][]byte),
	}
}

// TestVerifyRequestValidSignature proves the fixed verifyRequest method works end-to-end:
// client signs body, base64-encodes sig into header, gateway decodes and verifies.
func TestVerifyRequestValidSignature(t *testing.T) {
	clientKP, _ := pqc.GenerateDilithiumKeyPair()
	gw := newTestGateway("test-client", clientKP.GetPublicKeyBytes())

	body := []byte(`{"message":"hello"}`)
	sig, _ := clientKP.Sign(body)
	base64Sig := base64.StdEncoding.EncodeToString(sig)

	req := httptest.NewRequest("POST", "/echo", bytes.NewBuffer(body))
	req.Header.Set("X-Signature", base64Sig)

	err := gw.verifyRequest(req, "test-client")
	if err != nil {
		t.Errorf("verifyRequest() should pass with valid signature: %v", err)
	}
}

// TestVerifyRequestMissingSignature proves missing header is rejected.
func TestVerifyRequestMissingSignature(t *testing.T) {
	clientKP, _ := pqc.GenerateDilithiumKeyPair()
	gw := newTestGateway("test-client", clientKP.GetPublicKeyBytes())

	req := httptest.NewRequest("POST", "/echo", bytes.NewBuffer([]byte(`{}`)))
	// No X-Signature header

	err := gw.verifyRequest(req, "test-client")
	if err == nil {
		t.Error("verifyRequest() should fail with missing signature header")
	}
}

// TestVerifyRequestWrongKey proves wrong client key is rejected.
func TestVerifyRequestWrongKey(t *testing.T) {
	signerKP, _ := pqc.GenerateDilithiumKeyPair()
	wrongKP, _ := pqc.GenerateDilithiumKeyPair()
	// Gateway has wrongKP's public key cached for this client
	gw := newTestGateway("test-client", wrongKP.GetPublicKeyBytes())

	body := []byte(`{"message":"hello"}`)
	sig, _ := signerKP.Sign(body) // Signed with different key
	base64Sig := base64.StdEncoding.EncodeToString(sig)

	req := httptest.NewRequest("POST", "/echo", bytes.NewBuffer(body))
	req.Header.Set("X-Signature", base64Sig)

	err := gw.verifyRequest(req, "test-client")
	if err == nil {
		t.Error("verifyRequest() should fail when signature doesn't match cached public key")
	}
}

// TestHandleRequestRejects401 proves the full HTTP handler rejects unsigned requests.
func TestHandleRequestRejects401(t *testing.T) {
	clientKP, _ := pqc.GenerateDilithiumKeyPair()
	gw := newTestGateway("test-client", clientKP.GetPublicKeyBytes())

	req := httptest.NewRequest("POST", "/echo", bytes.NewBuffer([]byte(`{"test":true}`)))
	req.Header.Set("X-Client-ID", "test-client")
	// No X-Signature → should get 401

	rr := httptest.NewRecorder()
	gw.handleRequest(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("handleRequest() status = %d, want 401", rr.Code)
	}
}

// TestHandleRequestAllowsHealth proves /health bypasses verification.
func TestHandleRequestAllowsHealth(t *testing.T) {
	gw := newTestGateway("nobody", nil)

	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()
	gw.handleRequest(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("handleRequest(/health) status = %d, want 200", rr.Code)
	}
}
