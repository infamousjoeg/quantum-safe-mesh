package pqc

import (
	"log/slog"
	"os"
)

// Logger is a structured logger for PQC operations.
var Logger *slog.Logger

func init() {
	Logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

// LogKeyGen logs a key generation event.
func LogKeyGen(algorithm string, pubKeySize, privKeySize int) {
	Logger.Info("key_generated",
		"algorithm", algorithm,
		"public_key_bytes", pubKeySize,
		"private_key_bytes", privKeySize,
	)
}

// LogSignature logs a signature operation.
func LogSignature(operation, serviceID string, dataSize, sigSize int) {
	Logger.Info("signature_operation",
		"operation", operation,
		"service_id", serviceID,
		"data_bytes", dataSize,
		"signature_bytes", sigSize,
	)
}

// LogEncryption logs an encryption/decryption operation.
func LogEncryption(operation string, inputSize, outputSize int) {
	Logger.Info("encryption_operation",
		"operation", operation,
		"input_bytes", inputSize,
		"output_bytes", outputSize,
	)
}

// LogKeyExchange logs a key exchange event.
func LogKeyExchange(event, serviceID string, secretSize int) {
	Logger.Info("key_exchange",
		"event", event,
		"service_id", serviceID,
		"secret_bytes", secretSize,
	)
}

// LogRequest logs a request processing event.
func LogRequest(event, serviceID, nonce string) {
	Logger.Info("request",
		"event", event,
		"service_id", serviceID,
		"nonce", nonce,
	)
}
