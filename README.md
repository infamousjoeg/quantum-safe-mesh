# Quantum-Safe Service Mesh Authentication Demo

A comprehensive demonstration of Post-Quantum Cryptography (PQC) in a microservices architecture, showcasing quantum-resistant authentication and communication between services.

## 🌟 Overview

This project implements a minimal service mesh with three microservices that authenticate each other using post-quantum cryptography algorithms. It demonstrates how to build quantum-resistant systems that will remain secure even against future quantum computer attacks.

## 🏗️ Architecture

```
Client (Dilithium-signed) ──[TLS]──▶ 🌐 API Gateway (port 8081)
                                           │
                                    [TLS + AES-256-GCM + Dilithium]
                                           │
                                           ▼
                                     🧠 Backend Service (port 8082)
                                    
🔐 Auth Service (port 8080)
   ├── Key registry (Dilithium + Kyber public keys)
   ├── Authenticated registration (signature-verified)
   ├── Kyber768 key exchange for all services
   └── TLS certificate authority
```

### Services Overview

- **Auth Service** (`cmd/auth/`): Central authentication authority that manages service public keys and handles key exchange
- **API Gateway** (`cmd/gateway/`): Entry point that validates and forwards requests to backend services
- **Backend Service** (`cmd/backend/`): Processing service that handles business logic and returns signed responses

## 🔒 Post-Quantum Cryptography Algorithms

### Dilithium3 (Digital Signatures)
- **Purpose**: Service identity verification and message authentication
- **Key Size**: Public key ~1952 bytes, Private key ~4000 bytes
- **Signature Size**: ~3293 bytes
- **Security Level**: Equivalent to AES-192
- **Why**: Provides quantum-resistant digital signatures based on lattice problems

### Kyber768 (Key Encapsulation Mechanism)
- **Purpose**: Establishing shared secrets for encrypted communication channels  
- **Key Size**: Public key ~1184 bytes, Private key ~2400 bytes
- **Ciphertext Size**: ~1088 bytes
- **Security Level**: Equivalent to AES-192
- **Why**: Enables quantum-safe key exchange for session encryption

## 🚀 Quick Start

### Prerequisites

**For Local Development:**
- Go 1.22 or later
- Make (for using Makefile commands)

**For Kubernetes Deployment:**
- Docker
- Kubernetes cluster (local: kind/minikube, cloud: EKS/GKE/AKS)
- kubectl configured
- Helm 3.x (optional but recommended)

**For AWS Deployment:**
- Terraform (>= 1.0)
- AWS CLI (>= 2.0) configured with credentials
- AWS account with appropriate permissions

### Local Development Setup

#### 1. Clone and Setup
```bash
git clone <repository-url>
cd quantum-safe-mesh
go mod tidy
```

#### 2. Generate PQC Keys
```bash
make generate-keys
```

#### 3. Start Services (in separate terminals)
```bash
# Terminal 1: Start Auth Service
make run-auth

# Terminal 2: Start API Gateway  
make run-gateway

# Terminal 3: Start Backend Service
make run-backend
```

#### 4. Run Demo
```bash
# Terminal 4: Run the demo
make demo
```

### Kubernetes Deployment

#### Quick Deploy (Recommended)
```bash
# Complete automated deployment
./scripts/deploy.sh all
```

#### Manual Deployment Options

**Option 1: Using Helm (Recommended)**
```bash
# Build and deploy with Helm
./scripts/deploy.sh helm

# Or manually:
helm install quantum-safe-demo helm/quantum-safe-mesh \
  --namespace quantum-safe-mesh \
  --create-namespace
```

**Option 2: Using kubectl**
```bash
# Build and deploy with kubectl
./scripts/deploy.sh deploy

# Or manually:
kubectl apply -f k8s/
```

#### Deployment Script Usage
```bash
./scripts/deploy.sh [command]

Commands:
  build       Build Docker images
  deploy      Deploy using kubectl  
  helm        Deploy using Helm
  demo        Run demonstration
  status      Show deployment status
  cleanup     Remove deployment
  all         Build, deploy, and run demo
  help        Show help message
```

### AWS Deployment

#### Quick Deploy to AWS
```bash
# Complete AWS deployment with Terraform
./scripts/aws-deploy.sh all
```

#### Manual AWS Deployment
```bash
# Initialize and deploy infrastructure
./scripts/aws-deploy.sh init
./scripts/aws-deploy.sh plan
./scripts/aws-deploy.sh apply

# Deploy services to EC2 instance
./scripts/aws-deploy.sh deploy

# Setup monitoring
./scripts/aws-deploy.sh monitor

# Run tests
./scripts/aws-deploy.sh test
```

#### AWS Configuration
```bash
# Set environment variables
export AWS_REGION="us-west-2"
export ENVIRONMENT="dev"
export ALERT_EMAIL="alerts@example.com"

# Create terraform.tfvars
cd terraform
cat > terraform.tfvars << EOF
aws_region = "us-west-2"
environment = "dev"
instance_type = "t3.large"
allowed_cidr_blocks = ["0.0.0.0/0"]
alert_email = "your-email@example.com"
EOF
```

## 📖 Usage Guide

### Local Development Commands

```bash
make help              # Show all available commands
make generate-keys     # Generate PQC keypairs for all services
make run-auth         # Start auth service
make run-gateway      # Start gateway service  
make run-backend      # Start backend service
make demo             # Run complete demo flow
make benchmark        # PQC vs RSA performance comparison
make clean            # Clean build artifacts and keys
```

### Kubernetes Commands

```bash
# Deployment
./scripts/deploy.sh all          # Complete deployment + demo
./scripts/deploy.sh build        # Build Docker images only
./scripts/deploy.sh helm         # Deploy with Helm
./scripts/deploy.sh deploy       # Deploy with kubectl
./scripts/deploy.sh demo         # Run demo in cluster
./scripts/deploy.sh status       # Check deployment status
./scripts/deploy.sh cleanup      # Remove everything

# Helm-specific commands
helm list -n quantum-safe-mesh                    # List releases
helm status quantum-safe-demo -n quantum-safe-mesh  # Release status
helm upgrade quantum-safe-demo helm/quantum-safe-mesh  # Update
helm uninstall quantum-safe-demo -n quantum-safe-mesh  # Remove

# AWS deployment commands
./scripts/aws-deploy.sh all       # Complete AWS deployment
./scripts/aws-deploy.sh deploy    # Deploy services only
./scripts/aws-deploy.sh monitor   # Setup monitoring
./scripts/aws-deploy.sh test      # Run tests
./scripts/aws-deploy.sh info      # Show access information
./scripts/aws-deploy.sh destroy   # Destroy infrastructure
```

### Manual Testing

> **Note**: Services now use HTTPS with self-signed certificates. The demo client (`go run demo.go`) handles TLS and request signing automatically. For manual curl testing, use `-k` to skip certificate verification:

#### Health Checks (no signature required)
```bash
curl -k https://localhost:8080/health  # Auth service
curl -k https://localhost:8081/health  # Gateway
curl -k https://localhost:8082/health  # Backend
```

#### Prometheus Metrics
```bash
curl -k https://localhost:8081/metrics  # Gateway metrics
# Output includes: pqc_sign_total, pqc_verify_total, pqc_encrypt_total, etc.
```

#### Full Demo (recommended for testing)
```bash
# The demo client generates a keypair, registers, signs requests, and uses TLS
go run demo.go
```

#### Kubernetes Testing
```bash
# Port forwarding for external access
kubectl port-forward svc/gateway-service 8081:8081 -n quantum-safe-mesh

# Then use same curl commands as local testing
# Or use the ingress (if configured):
curl http://quantum-safe-mesh.local/echo \
  -H "Content-Type: application/json" \
  -d '{"message": "Hello K8s Quantum World!"}'

# View logs
kubectl logs -f deployment/auth-service -n quantum-safe-mesh
kubectl logs -f deployment/gateway-service -n quantum-safe-mesh  
kubectl logs -f deployment/backend-service -n quantum-safe-mesh

# Check demo results
kubectl logs job/quantum-safe-demo -n quantum-safe-mesh
```

#### AWS Testing
```bash
# After AWS deployment, get the instance IP from output
INSTANCE_IP=$(cd terraform && terraform output -raw instance_public_ip)

# Test services
curl http://$INSTANCE_IP:8081/health
curl -X POST http://$INSTANCE_IP:8081/echo \
  -H "Content-Type: application/json" \
  -d '{"message": "AWS deployment test"}'

# SSH into instance
ssh -i terraform/quantum-safe-mesh-key.pem ubuntu@$INSTANCE_IP

# View AWS monitoring
# - CloudWatch Dashboard: AWS Console → CloudWatch
# - Prometheus: http://$INSTANCE_IP:30090
# - Grafana: http://$INSTANCE_IP:30300 (admin/admin123)
```

## 🔐 Security Architecture — Seven Layers of Quantum Protection

This project implements defense-in-depth with seven security layers. Each layer is explained below in plain English, then with technical details.

### Layer 1: TLS Transport Encryption
**What it means**: Every network connection between services is encrypted, like a private tunnel. Even someone watching the network sees only gibberish.

**How it works**: A self-signed Certificate Authority (CA) generates TLS certificates for each service at startup. All services communicate over HTTPS. Plain HTTP is rejected with a 400 error.

### Layer 2: Client Authentication (Dilithium Signatures)
**What it means**: Before the gateway processes any request, the client must prove its identity with a quantum-proof digital signature — like showing a passport that no quantum computer can forge.

**How it works**: Clients generate a Dilithium3 keypair, register their public key with the Auth Service, then sign every request body with their private key. The gateway decodes the base64 signature from the `X-Signature` header and verifies it against the registered public key. Unsigned requests receive 401 Unauthorized.

### Layer 3: Authenticated Service Registration
**What it means**: When a service joins the mesh, it must prove it owns the private key matching the public key it's registering. An attacker can't register a fake key under someone else's name.

**How it works**: The registration payload is signed with the registrant's Dilithium private key. The Auth Service verifies the signature against the included public key before accepting registration.

### Layer 4: Kyber768 Key Exchange → AES-256-GCM Encryption
**What it means**: The gateway and backend establish a shared secret using quantum-safe key exchange, then use that secret to encrypt the actual data inside every request. Even if TLS were somehow broken, the payload is still encrypted.

**How it works**: 
1. Gateway fetches backend's Kyber768 public key from Auth Service
2. Gateway encapsulates: produces a ciphertext + 32-byte shared secret
3. Gateway sends ciphertext to backend's `/key-exchange` endpoint
4. Backend decapsulates the ciphertext to derive the same 32-byte shared secret
5. Both sides now use this secret as an AES-256-GCM key to encrypt/decrypt the `Data` field in every request and response

**Why two layers of encryption?** TLS protects the transport. AES-256-GCM protects the payload. This is defense-in-depth — if one layer is compromised, the other still protects the data.

### Layer 5: Digital Signatures (Dilithium3)
**What it means**: Every request and response is signed by the sender, proving who sent it and that it wasn't tampered with in transit.

**How it works**: The gateway signs the full `ServiceRequest` (including the encrypted data) with Dilithium3 before sending to backend. The backend verifies the signature, then signs its response. The gateway verifies the response signature before returning it to the client.

### Layer 6: Timestamp Validation (±30 seconds)
**What it means**: Every request includes a timestamp. If the request is more than 30 seconds old (or 30 seconds in the future), it's rejected. This prevents attackers from saving a valid request and replaying it later.

**How it works**: The `Timestamp` field in `ServiceRequest` is covered by the Dilithium signature, so it can't be modified. The backend checks `time.Since(request.Timestamp)` and rejects anything outside the ±30s window.

### Layer 7: Nonce Replay Protection
**What it means**: Every request carries a unique random ID (nonce). If the backend sees the same nonce twice within 60 seconds, it rejects the duplicate. Even if an attacker replays a request within the 30-second timestamp window, the nonce catches it.

**How it works**: The gateway generates a 16-byte random nonce for each request, included in the signed payload. The backend tracks seen nonces in a map with automatic 60-second expiry.

## 📊 Performance Comparison

Run benchmarks with proper Go testing framework:

```bash
go test -bench=. -benchmem ./pkg/pqc/
```

### Benchmark Results (Apple M-series, 16 cores)

| Operation | Time | vs RSA-2048 | Allocations |
|-----------|-----:|:-----------:|:-----------:|
| **RSA-2048 Sign** | 1,568 µs | baseline | 2 allocs |
| **RSA-2048 Verify** | 45 µs | baseline | 9 allocs |
| **Dilithium3 KeyGen** | 179 µs | — | 6 allocs |
| **Dilithium3 Sign** | 444 µs | **3.5x faster** | 5 allocs |
| **Dilithium3 Verify** | 108 µs | 2.4x slower | 6 allocs |
| **Kyber768 KeyGen** | 59 µs | — | 8 allocs |
| **Kyber768 Encapsulate** | 35 µs | — | 5 allocs |
| **Kyber768 Decapsulate** | 37 µs | — | 4 allocs |

### Key Size Comparison

| | RSA-2048 | Dilithium3 | Ratio |
|---|---:|---:|:---:|
| Public Key | 294 bytes | 1,952 bytes | 6.6x |
| Private Key | 1,192 bytes | 4,000 bytes | 3.4x |
| Signature | 256 bytes | 3,293 bytes | 12.9x |

**Key Insights:**
- ✅ **Dilithium signing is 3.5x faster** than RSA-2048
- ✅ **Kyber operations are extremely fast** (35-59 µs)
- ⚠️ **Larger key and signature sizes** — an acceptable trade-off for quantum resistance
- ✅ **AES-256-GCM encryption overhead is negligible** (< 1 µs per operation)

## 🛡️ Security Benefits

### Quantum Resistance
- **Future-proof**: Secure against both classical and quantum computer attacks
- **NIST Standards**: Uses NIST-approved PQC algorithms
- **Forward Security**: Key exchange provides forward secrecy

### Zero-Trust Architecture
- **Mutual Authentication**: All services verify each other's signatures
- **Request Integrity**: Every message is cryptographically signed
- **Service Identity**: Strong service identity verification
- **Tamper Detection**: Any modification to messages is detected

### Security Properties Summary

| Property | Status | Implementation |
|----------|:------:|----------------|
| Quantum-resistant signatures | ✅ | Dilithium3 (NIST FIPS 204 / ML-DSA) |
| Quantum-resistant key exchange | ✅ | Kyber768 (NIST FIPS 203 / ML-KEM) |
| Payload encryption | ✅ | AES-256-GCM with Kyber-derived 32-byte key |
| Client → Gateway auth | ✅ | Dilithium-signed requests, 401 for unsigned |
| Gateway → Backend auth | ✅ | Dilithium-signed ServiceRequests |
| Backend → Gateway auth | ✅ | Dilithium-signed ServiceResponses |
| Transport encryption | ✅ | TLS with self-signed CA certificates |
| Replay protection | ✅ | ±30s timestamps + unique nonces |
| Authenticated registration | ✅ | Dilithium signature proof of key ownership |
| Key rotation | ✅ | Hourly Kyber shared secret rotation |
| Observability | ✅ | Prometheus /metrics on all services |

## 🧪 Testing & Validation

### Automated Test Suite (40 tests)
```bash
go test ./... -v          # Run all 40 tests
go test -bench=. ./pkg/pqc/  # Run PQC benchmarks
go vet ./...              # Static analysis
```

### Test Coverage
- **pkg/pqc/** (24 tests): Dilithium sign/verify, Kyber encap/decap, AES-GCM encrypt/decrypt, TLS cert generation, key persistence, public key deserialization, RSA vs PQC benchmarks
- **cmd/gateway/** (8 tests): Client signature verification (valid/missing/wrong key), HTTP 401 enforcement, health bypass, gateway→backend signing, base64 decode validation
- **cmd/backend/** (5 tests): Request verification, tamper detection, impersonation rejection, timestamp validation, replay protection
- **cmd/auth/** (3 tests): Signed registration accepted, unsigned rejected (401), wrong-key signature rejected (401)

## 📝 Educational Value

### Understanding PQC Concepts

1. **Digital Signatures**: How Dilithium provides quantum-safe message authentication
2. **Key Encapsulation**: How Kyber enables secure key exchange
3. **Service Mesh Security**: Applying PQC in distributed systems
4. **Performance Trade-offs**: Understanding the costs of quantum resistance

### Code Structure
```
pkg/
├── pqc/               # PQC cryptographic operations
│   ├── dilithium.go       # Dilithium3 digital signatures
│   ├── kyber.go           # Kyber768 key encapsulation
│   ├── aead.go            # AES-256-GCM encryption (Kyber-derived keys)
│   ├── tlsgen.go          # Self-signed CA and TLS certificate generation
│   ├── metrics.go         # Prometheus metrics (sign/verify/encrypt counters)
│   ├── logger.go          # Structured JSON logging via slog
│   └── utils.go           # Key persistence and RSA benchmark
├── models/            # Data structures (ServiceRequest, ServiceResponse, etc.)
└── ...
cmd/
├── auth/              # Auth Service: key registry, registration, key exchange
├── gateway/           # API Gateway: client auth, encryption, forwarding
└── backend/           # Backend: verification, decryption, processing
```

## 🧠 Understanding the Cryptography

> This section explains the cryptography in plain English. No math degree required.

### What Is a Lattice?
Imagine a grid of dots extending infinitely in all directions. A lattice is a mathematical version of this grid, but in hundreds of dimensions. Finding the closest point on this grid is easy if you know the structure, but incredibly hard if you don't — even for quantum computers. This difficulty is what makes lattice-based cryptography secure.

### Why Can't Quantum Computers Break Lattices?
RSA security relies on factoring large numbers — something quantum computers can do efficiently using Shor's algorithm. Lattice problems are fundamentally different. No known quantum algorithm can efficiently solve them. This is why NIST chose lattice-based algorithms as the post-quantum standard.

### What Is a Key Encapsulation Mechanism (KEM)?
Think of it like a lockbox exchange:
1. Backend creates a special lockbox (Kyber public key) and shares it
2. Gateway puts a random secret in the lockbox and locks it (encapsulation)
3. Gateway sends the locked box to backend
4. Backend opens it with their private key (decapsulation)
5. Now both sides have the same secret, and nobody else does

This secret becomes the AES-256-GCM encryption key.

### Why Both KEM and AES?
Kyber (the KEM) is used to **agree on a key**. AES-256-GCM is used to **encrypt data with that key**. Kyber can't directly encrypt large payloads — it only produces a 32-byte shared secret. AES-256-GCM takes that 32-byte secret and uses it to encrypt any amount of data, fast.

### What Is a Digital Signature?
Dilithium3 is like an unforgeable stamp. When the gateway signs a request:
1. It creates a unique stamp from the message content + its private key
2. Anyone with the public key can verify the stamp is genuine
3. But nobody can create a valid stamp without the private key
4. A quantum computer can't forge the stamp (unlike RSA signatures)

### Why Do Nonces Matter?
Imagine you intercept a valid, signed bank transfer for $1,000. Without nonces, you could send that exact same signed request again and again — each time transferring another $1,000. The signature is valid each time because the message hasn't changed.

A nonce (number used once) is a unique random ID in each request. The backend remembers every nonce it's seen. If it sees the same nonce twice, it knows it's a replay attack and rejects it.

## 🔮 The "Harvest Now, Decrypt Later" Threat Timeline

### 📊 Current State (2025)
- **Quantum Computing Progress**: 1000+ qubit systems operational (IBM, Google, IonQ)
- **Cryptanalytic-Relevant Quantum Computers (CRQC)**: Not yet achieved
- **Data Harvesting**: Adversaries actively collecting encrypted data for future decryption
- **Current Risk**: All RSA, ECC, and DH-encrypted data vulnerable to future quantum attacks

### ⚠️ Critical Timeline Milestones

#### **2025-2030: The Harvest Window**
- **What's Happening**: Nation-states and advanced adversaries are harvesting encrypted communications
- **Target Data**: Banking transactions, healthcare records, government communications, intellectual property
- **Storage Capacity**: Cloud storage and quantum-ready infrastructure making mass data collection feasible
- **Risk Level**: 🔴 **CRITICAL** - All current encrypted data will be readable by future quantum computers

#### **2030-2035: Quantum Advantage Achieved**
- **CRQC Development**: First cryptanalytic-relevant quantum computers expected
- **RSA-2048 Broken**: Quantum computers capable of breaking 2048-bit RSA in hours/days
- **ECC Vulnerability**: Elliptic curve cryptography rendered obsolete
- **Mass Decryption**: Previously harvested data becomes readable

#### **2035+: Post-Quantum Era**
- **Widespread CRQC**: Multiple nations and organizations possess quantum computers
- **Legacy Systems Exposed**: Any system still using classical cryptography fully compromised
- **Data Retroactively Compromised**: 10+ years of harvested data becomes accessible

### 🛡️ Post-Quantum Cryptography Defense

#### **Mathematical Foundation**
- **Lattice-Based**: Dilithium signatures resist quantum attacks via Learning With Errors (LWE) problem
- **Code-Based**: Kyber key exchange based on Module Learning With Errors (MLWE)
- **Quantum-Safe Timeline**: Secure against both classical and quantum computer attacks

#### **Migration Urgency**
- **NIST Standards**: Published final standards in 2024 (FIPS 203, 204, 205)
- **Industry Adoption**: Major cloud providers implementing PQC by 2025-2026
- **Regulatory Compliance**: Government mandates for PQC migration by 2030-2035
- **Protection Window**: Migrate now to protect against future quantum decryption

### 🎯 Service Mesh Context
- **Microservices Security**: Critical for zero-trust architectures in quantum era
- **Inter-service Communication**: Quantum-safe authentication for all service-to-service traffic
- **Data in Transit**: Protect API communications from harvest attacks
- **Future-proofing**: Deploy quantum-resistant cryptography before CRQC emergence

### 📈 Industry Impact Assessment
- **Healthcare**: 25+ year data retention requirements demand immediate PQC adoption
- **Financial Services**: Real-time quantum-safe authentication for trading systems
- **Government**: Classified data protection requiring immediate quantum resistance
- **Critical Infrastructure**: Power grids, telecommunications, transportation systems at risk

> **🚨 Key Insight**: The window for safe migration is narrowing. Organizations must implement post-quantum cryptography **before** quantum computers become capable, not after. Every day of delay increases the risk of retroactive data compromise.

## 🏭 Industry-Specific Implementation Guides

Comprehensive deployment guides tailored for specific industries are available in the [`examples/`](./examples/) directory:

- **[Healthcare](./examples/HEALTHCARE.md)** - Patient data protection, HIPAA compliance, telemedicine security
- **[Financial Services](./examples/FINANCIAL_SERVICES.md)** - Trading systems, payment networks, regulatory compliance  
- **[Government & Defense](./examples/GOVERNMENT.md)** - Classified systems, national security, multi-level security
- **[Technology](./examples/TECHNOLOGY.md)** - Cloud platforms, SaaS, AI/ML model protection
- **[Critical Infrastructure](./examples/CRITICAL_INFRASTRUCTURE.md)** - SCADA systems, power grids, transportation networks

Each guide includes:
- Industry-specific threat assessment and timeline
- Regulatory compliance frameworks
- Implementation strategies and code examples
- Deployment scenarios and performance requirements
- Business case analysis and ROI calculations

## 🎯 Cloud-Native Features

### Kubernetes Integration
✅ **Container Orchestration**: Multi-replica deployments with auto-scaling  
✅ **Service Discovery**: DNS-based service resolution  
✅ **Load Balancing**: Built-in Kubernetes service load balancing  
✅ **Health Checks**: Liveness and readiness probes  
✅ **Secret Management**: Kubernetes secrets for key storage  
✅ **Network Security**: Network policies for zero-trust networking  
✅ **Ingress Support**: External access via ingress controllers  
✅ **Monitoring Ready**: Prometheus metrics and Grafana dashboards  

### Helm Chart Features
- **Configurable Values**: Easy customization via values.yaml
- **Environment Support**: Dev, staging, production configurations  
- **Resource Management**: CPU/memory limits and requests
- **Storage Options**: Persistent vs. ephemeral key storage
- **Security Policies**: Pod security policies and contexts
- **Upgrade Strategy**: Rolling updates with zero downtime

### Deployment Flexibility
- **Multi-Cloud Support**: Works on any Kubernetes distribution
- **Local Development**: Kind, minikube, Docker Desktop
- **Cloud Platforms**: EKS, GKE, AKS, OpenShift
- **Edge Computing**: K3s, MicroK8s support
- **GitOps Ready**: ArgoCD/Flux compatible

## 🚧 What's Included vs. Future Work

### ✅ What This Project Implements
- PQC algorithm integration (Dilithium3 + Kyber768)
- Zero-trust client and service authentication
- AES-256-GCM payload encryption using Kyber-derived keys
- TLS on all inter-service communication
- Timestamp validation (±30s) and nonce replay protection
- Authenticated service registration (signed proof of key ownership)
- Periodic key rotation (1 hour interval)
- Prometheus metrics on all services
- 40 automated tests with proper Go benchmarks
- GitHub Actions CI (test + vet + benchmark)
- Kubernetes-native deployment (Helm + kubectl)
- AWS deployment with Terraform

### 🔮 Future Production Requirements
- [ ] Hardware security module (HSM) integration for key storage
- [ ] Certificate management with cert-manager
- [ ] Comprehensive audit logging pipeline
- [ ] Rate limiting and DoS protection
- [ ] Advanced alerting (PagerDuty, Slack integration)
- [ ] Disaster recovery procedures
- [ ] Security scanning and vulnerability management
- [ ] Compliance reporting (SOC2, ISO 27001)

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## 📚 References

- [NIST Post-Quantum Cryptography Standardization](https://csrc.nist.gov/projects/post-quantum-cryptography)
- [Cloudflare CIRCL Library](https://github.com/cloudflare/circl)
- [Dilithium Algorithm Specification](https://pq-crystals.org/dilithium/)
- [Kyber Algorithm Specification](https://pq-crystals.org/kyber/)

## 📄 License

This project is provided for educational and demonstration purposes. Please review the individual algorithm licenses and compliance requirements for production use.

## 🔧 Kubernetes Architecture Details

### Pod Security
```yaml
securityContext:
  allowPrivilegeEscalation: false
  runAsNonRoot: true
  runAsUser: 1000
  capabilities:
    drop: [ALL]
```

### Resource Requirements
- **Auth Service**: 64Mi RAM, 50m CPU (min) / 128Mi RAM, 200m CPU (max)
- **Gateway Service**: 64Mi RAM, 50m CPU (min) / 128Mi RAM, 200m CPU (max)  
- **Backend Service**: 64Mi RAM, 50m CPU (min) / 128Mi RAM, 200m CPU (max)

### Storage Configuration
```yaml
# Ephemeral storage (default)
volumes:
- name: keys-storage
  emptyDir: {}

# Persistent storage (optional)
volumes:
- name: keys-storage
  persistentVolumeClaim:
    claimName: quantum-safe-keys
```

### Network Policies (Zero Trust)
- Auth Service: Accepts connections from Gateway and Backend only
- Gateway Service: Accepts external traffic, connects to Auth and Backend
- Backend Service: Accepts connections from Gateway only, connects to Auth

### Monitoring Integration
```bash
# Prometheus metrics endpoints
https://auth-service:8080/metrics
https://gateway-service:8081/metrics
https://backend-service:8082/metrics

# PQC operation metrics
pqc_sign_total
pqc_verify_total
pqc_verify_fail_total
pqc_encrypt_total
pqc_decrypt_total
pqc_encapsulate_total
pqc_decapsulate_total
pqc_key_exchange_total
pqc_request_total
pqc_replay_reject_total
```

### Environment Variables
```yaml
# Service discovery
AUTH_SERVICE_URL: "https://auth-service.quantum-safe-mesh.svc.cluster.local:8080"
BACKEND_SERVICE_URL: "https://backend-service.quantum-safe-mesh.svc.cluster.local:8082"

# Pod information
POD_NAMESPACE: valueFrom fieldRef metadata.namespace
NODE_NAME: valueFrom fieldRef spec.nodeName
```

---

**🌟 This demo showcases the future of cryptographic security in cloud-native distributed systems. Deploy, scale, and prepare for the quantum era with Kubernetes!**