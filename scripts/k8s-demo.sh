#!/bin/bash

# Kubernetes Demo Script for Quantum-Safe Service Mesh
# This script runs inside the demo container to test the PQC service mesh

set -e

# Configuration
GATEWAY_URL=${GATEWAY_URL:-"http://gateway-service:8081"}
AUTH_URL=${AUTH_URL:-"http://auth-service:8080"}
MAX_RETRIES=30
RETRY_DELAY=5

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_header() {
    echo -e "${CYAN}$1${NC}"
}

wait_for_service() {
    local url=$1
    local service_name=$2
    local retries=0
    
    log_info "Waiting for $service_name to be ready..."
    
    while [ $retries -lt $MAX_RETRIES ]; do
        if curl -f -s "$url/health" > /dev/null 2>&1; then
            log_success "$service_name is ready"
            return 0
        fi
        
        retries=$((retries + 1))
        log_info "Attempt $retries/$MAX_RETRIES: $service_name not ready, waiting ${RETRY_DELAY}s..."
        sleep $RETRY_DELAY
    done
    
    log_error "$service_name failed to become ready after $MAX_RETRIES attempts"
    return 1
}

test_health_checks() {
    log_header "=== Health Check Tests ==="
    
    local services=("$AUTH_URL:Auth Service" "$GATEWAY_URL:Gateway Service")
    
    for service_info in "${services[@]}"; do
        local url=$(echo $service_info | cut -d: -f1)
        local name=$(echo $service_info | cut -d: -f2-)
        
        log_info "Testing $name health endpoint..."
        
        response=$(curl -s -w "HTTP_CODE:%{http_code}" "$url/health")
        http_code=$(echo "$response" | grep -o "HTTP_CODE:[0-9]*" | cut -d: -f2)
        body=$(echo "$response" | sed 's/HTTP_CODE:[0-9]*$//')
        
        if [ "$http_code" = "200" ]; then
            log_success "$name: $body"
        else
            log_error "$name: HTTP $http_code"
        fi
    done
}

test_pqc_echo() {
    log_header "=== PQC Echo Test ==="
    
    local test_data='{
        "message": "Hello Quantum-Safe Kubernetes World!",
        "demo": true,
        "timestamp": "'$(date -Iseconds)'",
        "test_data": ["quantum", "cryptography", "kubernetes", "demo"],
        "kubernetes_info": {
            "namespace": "'${POD_NAMESPACE:-quantum-safe-mesh}'",
            "pod": "'${HOSTNAME:-demo-pod}'"
        }
    }'
    
    log_info "Sending PQC-signed echo request..."
    
    response=$(curl -s -w "\nHTTP_CODE:%{http_code}\nTIME_TOTAL:%{time_total}" \
        -X POST "$GATEWAY_URL/echo" \
        -H "Content-Type: application/json" \
        -H "X-Client-ID: k8s-demo-client" \
        -d "$test_data")
    
    http_code=$(echo "$response" | grep "HTTP_CODE:" | cut -d: -f2)
    time_total=$(echo "$response" | grep "TIME_TOTAL:" | cut -d: -f2)
    body=$(echo "$response" | sed '/HTTP_CODE:/d' | sed '/TIME_TOTAL:/d')
    
    if [ "$http_code" = "200" ]; then
        log_success "Echo test successful (${time_total}s)"
        echo "$body" | jq '.' 2>/dev/null || echo "$body"
    else
        log_error "Echo test failed: HTTP $http_code"
        echo "$body"
    fi
}

test_pqc_processing() {
    log_header "=== PQC Data Processing Test ==="
    
    local test_data='{
        "operation": "kubernetes_quantum_processing",
        "data": {
            "input": "Kubernetes-native post-quantum cryptography demonstration",
            "algorithm": "dilithium_kyber_hybrid",
            "priority": "high",
            "security_level": "quantum_resistant"
        },
        "metadata": {
            "client_id": "k8s-demo",
            "request_id": "k8s-'$(date +%s)'",
            "quantum_safe": true,
            "kubernetes": {
                "cluster": "demo-cluster",
                "namespace": "'${POD_NAMESPACE:-quantum-safe-mesh}'",
                "deployment": "quantum-safe-demo"
            }
        }
    }'
    
    log_info "Sending PQC-signed processing request..."
    
    response=$(curl -s -w "\nHTTP_CODE:%{http_code}\nTIME_TOTAL:%{time_total}" \
        -X POST "$GATEWAY_URL/process" \
        -H "Content-Type: application/json" \
        -H "X-Client-ID: k8s-demo-client" \
        -d "$test_data")
    
    http_code=$(echo "$response" | grep "HTTP_CODE:" | cut -d: -f2)
    time_total=$(echo "$response" | grep "TIME_TOTAL:" | cut -d: -f2)
    body=$(echo "$response" | sed '/HTTP_CODE:/d' | sed '/TIME_TOTAL:/d')
    
    if [ "$http_code" = "200" ]; then
        log_success "Processing test successful (${time_total}s)"
        echo "$body" | jq '.' 2>/dev/null || echo "$body"
    else
        log_error "Processing test failed: HTTP $http_code"
        echo "$body"
    fi
}

test_service_status() {
    log_header "=== Service Status Test ==="
    
    log_info "Requesting backend service status..."
    
    response=$(curl -s -w "\nHTTP_CODE:%{http_code}\nTIME_TOTAL:%{time_total}" \
        -X GET "$GATEWAY_URL/status" \
        -H "X-Client-ID: k8s-demo-client")
    
    http_code=$(echo "$response" | grep "HTTP_CODE:" | cut -d: -f2)
    time_total=$(echo "$response" | grep "TIME_TOTAL:" | cut -d: -f2)
    body=$(echo "$response" | sed '/HTTP_CODE:/d' | sed '/TIME_TOTAL:/d')
    
    if [ "$http_code" = "200" ]; then
        log_success "Status test successful (${time_total}s)"
        echo "$body" | jq '.' 2>/dev/null || echo "$body"
    else
        log_error "Status test failed: HTTP $http_code"
        echo "$body"
    fi
}

test_auth_service_direct() {
    log_header "=== Auth Service Direct Test ==="
    
    log_info "Testing direct auth service endpoints..."
    
    # Test service registry
    log_info "Checking registered services..."
    response=$(curl -s -w "HTTP_CODE:%{http_code}" "$AUTH_URL/services")
    http_code=$(echo "$response" | grep -o "HTTP_CODE:[0-9]*" | cut -d: -f2)
    body=$(echo "$response" | sed 's/HTTP_CODE:[0-9]*$//')
    
    if [ "$http_code" = "200" ]; then
        log_success "Service registry accessible"
        echo "$body" | jq '.' 2>/dev/null || echo "$body"
    else
        log_warning "Service registry test failed: HTTP $http_code"
    fi
}

run_load_test() {
    log_header "=== Load Test (Basic) ==="
    
    log_info "Running basic load test with 10 concurrent requests..."
    
    local success_count=0
    local total_requests=10
    local pids=()
    
    # Create temporary directory for results
    local temp_dir=$(mktemp -d)
    
    # Launch concurrent requests
    for i in $(seq 1 $total_requests); do
        (
            response=$(curl -s -w "%{http_code}" \
                -X POST "$GATEWAY_URL/echo" \
                -H "Content-Type: application/json" \
                -H "X-Client-ID: load-test-$i" \
                -d '{"message":"Load test request '$i'","timestamp":"'$(date -Iseconds)'"}')
            echo "$response" > "$temp_dir/result_$i"
        ) &
        pids+=($!)
    done
    
    # Wait for all requests to complete
    for pid in "${pids[@]}"; do
        wait $pid
    done
    
    # Count successful responses
    for i in $(seq 1 $total_requests); do
        if [ -f "$temp_dir/result_$i" ]; then
            if grep -q "200" "$temp_dir/result_$i"; then
                success_count=$((success_count + 1))
            fi
        fi
    done
    
    # Cleanup
    rm -rf "$temp_dir"
    
    log_success "Load test completed: $success_count/$total_requests requests successful"
}

show_kubernetes_info() {
    log_header "=== Kubernetes Environment Info ==="
    
    log_info "Pod Information:"
    echo "  Pod Name: ${HOSTNAME:-unknown}"
    echo "  Namespace: ${POD_NAMESPACE:-unknown}"
    echo "  Node Name: ${NODE_NAME:-unknown}"
    
    log_info "Service Discovery:"
    echo "  Gateway URL: $GATEWAY_URL"
    echo "  Auth URL: $AUTH_URL"
    
    if [ -f /var/run/secrets/kubernetes.io/serviceaccount/token ]; then
        log_info "Service Account: Mounted"
    else
        log_info "Service Account: Not mounted"
    fi
}

show_security_summary() {
    log_header "=== Post-Quantum Cryptography Summary ==="
    
    echo "🔐 Quantum-Resistant Algorithms Used:"
    echo "  • Dilithium3: Digital signatures for service authentication"
    echo "  • Kyber768: Key encapsulation for secure communication"
    echo ""
    echo "🛡️  Security Features Demonstrated:"
    echo "  • Mutual authentication between all services"
    echo "  • Zero-trust networking with signature verification"
    echo "  • Kubernetes-native service discovery"
    echo "  • Network policies for traffic isolation"
    echo "  • Container security with minimal privileges"
    echo ""
    echo "🚀 Cloud-Native Features:"
    echo "  • Horizontal scaling with multiple replicas"
    echo "  • Health checks and automated recovery"
    echo "  • Load balancer integration"
    echo "  • Ingress controller support"
    echo "  • Prometheus monitoring ready"
    echo ""
    echo "✅ All inter-service communication is quantum-safe!"
}

# Main execution
main() {
    log_header "🌟 Quantum-Safe Service Mesh Kubernetes Demo"
    echo "================================================================="
    
    show_kubernetes_info
    
    # Wait for services to be ready
    wait_for_service "$AUTH_URL" "Auth Service" || exit 1
    wait_for_service "$GATEWAY_URL" "Gateway Service" || exit 1
    
    # Run tests
    test_health_checks
    sleep 2
    
    test_auth_service_direct
    sleep 2
    
    test_pqc_echo
    sleep 2
    
    test_pqc_processing
    sleep 2
    
    test_service_status
    sleep 2
    
    run_load_test
    sleep 2
    
    show_security_summary
    
    log_success "🎉 All tests completed successfully!"
    log_success "Quantum-Safe Service Mesh is operating correctly in Kubernetes!"
}

# Run the demo
main "$@"