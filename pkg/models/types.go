package models

import (
	"encoding/json"
	"time"
)

type ServiceKeyPair struct {
	ServiceID  string `json:"service_id"`
	PublicKey  []byte `json:"public_key"`
	PrivateKey []byte `json:"private_key,omitempty"`
}

type ServiceRequest struct {
	ServiceID string            `json:"service_id"`
	Timestamp time.Time         `json:"timestamp"`
	Data      json.RawMessage   `json:"data"`
	Signature []byte            `json:"signature"`
	Headers   map[string]string `json:"headers,omitempty"`
}

type ServiceResponse struct {
	ServiceID string            `json:"service_id"`
	Timestamp time.Time         `json:"timestamp"`
	Data      json.RawMessage   `json:"data"`
	Signature []byte            `json:"signature"`
	Headers   map[string]string `json:"headers,omitempty"`
	Success   bool              `json:"success"`
	Error     string            `json:"error,omitempty"`
}

type AuthToken struct {
	ServiceID string    `json:"service_id"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Signature []byte    `json:"signature"`
}

type KeyExchangeRequest struct {
	ServiceID      string    `json:"service_id"`
	KyberPublicKey []byte    `json:"kyber_public_key"`
	Timestamp      time.Time `json:"timestamp"`
	Signature      []byte    `json:"signature"`
}

type KeyExchangeResponse struct {
	Ciphertext   []byte    `json:"ciphertext"`
	SharedSecret []byte    `json:"shared_secret,omitempty"`
	Timestamp    time.Time `json:"timestamp"`
	Signature    []byte    `json:"signature"`
}
