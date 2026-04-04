package pqc

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

// Metrics tracks PQC operation counts.
type Metrics struct {
	SignCount         atomic.Int64
	VerifyCount       atomic.Int64
	VerifyFailCount   atomic.Int64
	EncryptCount      atomic.Int64
	DecryptCount      atomic.Int64
	EncapsulateCount  atomic.Int64
	DecapsulateCount  atomic.Int64
	KeyExchangeCount  atomic.Int64
	RequestCount      atomic.Int64
	ReplayRejectCount atomic.Int64
}

// GlobalMetrics is the singleton metrics instance.
var GlobalMetrics = &Metrics{}

// MetricsHandler returns an HTTP handler that serves Prometheus-format metrics.
func MetricsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		m := GlobalMetrics
		fmt.Fprintf(w, "# HELP pqc_sign_total Total Dilithium sign operations\n")
		fmt.Fprintf(w, "# TYPE pqc_sign_total counter\n")
		fmt.Fprintf(w, "pqc_sign_total %d\n", m.SignCount.Load())
		fmt.Fprintf(w, "# HELP pqc_verify_total Total Dilithium verify operations\n")
		fmt.Fprintf(w, "# TYPE pqc_verify_total counter\n")
		fmt.Fprintf(w, "pqc_verify_total %d\n", m.VerifyCount.Load())
		fmt.Fprintf(w, "# HELP pqc_verify_fail_total Total failed signature verifications\n")
		fmt.Fprintf(w, "# TYPE pqc_verify_fail_total counter\n")
		fmt.Fprintf(w, "pqc_verify_fail_total %d\n", m.VerifyFailCount.Load())
		fmt.Fprintf(w, "# HELP pqc_encrypt_total Total AES-GCM encrypt operations\n")
		fmt.Fprintf(w, "# TYPE pqc_encrypt_total counter\n")
		fmt.Fprintf(w, "pqc_encrypt_total %d\n", m.EncryptCount.Load())
		fmt.Fprintf(w, "# HELP pqc_decrypt_total Total AES-GCM decrypt operations\n")
		fmt.Fprintf(w, "# TYPE pqc_decrypt_total counter\n")
		fmt.Fprintf(w, "pqc_decrypt_total %d\n", m.DecryptCount.Load())
		fmt.Fprintf(w, "# HELP pqc_encapsulate_total Total Kyber encapsulate operations\n")
		fmt.Fprintf(w, "# TYPE pqc_encapsulate_total counter\n")
		fmt.Fprintf(w, "pqc_encapsulate_total %d\n", m.EncapsulateCount.Load())
		fmt.Fprintf(w, "# HELP pqc_decapsulate_total Total Kyber decapsulate operations\n")
		fmt.Fprintf(w, "# TYPE pqc_decapsulate_total counter\n")
		fmt.Fprintf(w, "pqc_decapsulate_total %d\n", m.DecapsulateCount.Load())
		fmt.Fprintf(w, "# HELP pqc_key_exchange_total Total Kyber key exchanges\n")
		fmt.Fprintf(w, "# TYPE pqc_key_exchange_total counter\n")
		fmt.Fprintf(w, "pqc_key_exchange_total %d\n", m.KeyExchangeCount.Load())
		fmt.Fprintf(w, "# HELP pqc_request_total Total requests processed\n")
		fmt.Fprintf(w, "# TYPE pqc_request_total counter\n")
		fmt.Fprintf(w, "pqc_request_total %d\n", m.RequestCount.Load())
		fmt.Fprintf(w, "# HELP pqc_replay_reject_total Total replay attacks rejected\n")
		fmt.Fprintf(w, "# TYPE pqc_replay_reject_total counter\n")
		fmt.Fprintf(w, "pqc_replay_reject_total %d\n", m.ReplayRejectCount.Load())
	}
}
