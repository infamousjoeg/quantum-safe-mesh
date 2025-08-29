# Quantum-Safe Service Mesh Authentication Demo

A comprehensive demonstration of Post-Quantum Cryptography (PQC) in a microservices architecture, showcasing quantum-resistant authentication and communication between services.

## 🌟 Overview

This project implements a minimal service mesh with three microservices that authenticate each other using post-quantum cryptography algorithms. It demonstrates how to build quantum-resistant systems that will remain secure even against future quantum computer attacks.

## 🏗️ Architecture

```
Client Request
      ↓
🌐 API Gateway (port 8081)
      ↓ (PQC-signed request)
🧠 Backend Service (port 8082)
      ↑ (PQC-signed response)
      
🔐 Auth Service (port 8080)
   (Key registry & validation)
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
- Go 1.21 or later
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

#### Local Testing
```bash
# Health Check
curl http://localhost:8080/health  # Auth service
curl http://localhost:8081/health  # Gateway
curl http://localhost:8082/health  # Backend

# Echo Test
curl -X POST http://localhost:8081/echo \
  -H "Content-Type: application/json" \
  -H "X-Client-ID: test-client" \
  -d '{"message": "Hello Quantum World!"}'

# Data Processing Test
curl -X POST http://localhost:8081/process \
  -H "Content-Type: application/json" \
  -H "X-Client-ID: test-client" \
  -d '{"data": "quantum processing test", "algorithm": "advanced"}'
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

## 🔐 Security Flow

### 1. Service Registration
```
Service → Auth Service: Register public key
Auth Service → Service: Registration confirmation (signed)
```

### 2. Request Authentication  
```
Client → Gateway: Request
Gateway → Backend: PQC-signed request
Backend: Verifies Gateway signature using Auth Service public key registry
Backend → Gateway: PQC-signed response
Gateway: Verifies Backend signature
Gateway → Client: Response
```

### 3. Key Exchange (Kyber768)
```
Service A → Auth Service: Key exchange request (signed with Dilithium)
Auth Service: Verifies request signature
Auth Service → Service A: Encrypted shared secret (Kyber encapsulation)
Service A: Decapsulates shared secret
```

## 📊 Performance Comparison

The system includes built-in benchmarking to compare PQC algorithms with traditional cryptography:

```bash
make benchmark
```

### Expected Performance Characteristics

| Algorithm | Operation | Traditional (RSA-2048) | Post-Quantum | Ratio |
|-----------|-----------|----------------------|--------------|-------|
| Sign      | Time      | ~1ms                 | ~0.1ms       | 0.1x  |
| Verify    | Time      | ~0.05ms              | ~0.05ms      | 1.0x  |
| Key Size  | Public    | ~270 bytes           | ~1952 bytes  | 7.2x  |
| Key Size  | Private   | ~1190 bytes          | ~4000 bytes  | 3.4x  |
| Signature | Size      | ~256 bytes           | ~3293 bytes  | 12.9x |

**Key Insights:**
- ✅ **Dilithium signing is faster** than RSA
- ✅ **Verification speeds are comparable**
- ⚠️ **Larger key and signature sizes** (acceptable trade-off for quantum resistance)

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

## 🧪 Testing & Validation

### Automated Testing
```bash
make test          # Run unit tests
make demo          # Integration test
make load-test     # Basic load testing
```

### Manual Verification
1. **Signature Verification**: All responses include PQC signatures that are verified
2. **Key Exchange**: Services perform Kyber key exchange for secure channels  
3. **Authentication Flow**: Each request goes through full PQC authentication
4. **Performance Monitoring**: Built-in latency and throughput measurements

## 📝 Educational Value

### Understanding PQC Concepts

1. **Digital Signatures**: How Dilithium provides quantum-safe message authentication
2. **Key Encapsulation**: How Kyber enables secure key exchange
3. **Service Mesh Security**: Applying PQC in distributed systems
4. **Performance Trade-offs**: Understanding the costs of quantum resistance

### Code Structure
```
pkg/
├── pqc/           # PQC cryptographic operations
│   ├── dilithium.go   # Digital signature functions
│   ├── kyber.go       # Key encapsulation functions  
│   └── utils.go       # Key management and benchmarks
├── models/        # Data structures
└── ...
cmd/
├── auth/          # Authentication service
├── gateway/       # API Gateway service  
└── backend/       # Backend processing service
```

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

## 🚧 Production Considerations

### What This Demo Shows
✅ PQC algorithm integration  
✅ Service authentication flow  
✅ Performance characteristics  
✅ Error handling and logging  
✅ Kubernetes-native deployment  
✅ Container security best practices  
✅ Horizontal scaling capabilities  
✅ Service mesh architecture  

### Additional Production Requirements
- [ ] Certificate management and rotation
- [ ] Hardware security module (HSM) integration
- [ ] Comprehensive audit logging
- [ ] Rate limiting and DoS protection
- [ ] Advanced monitoring and alerting
- [ ] Disaster recovery procedures
- [ ] Security scanning and vulnerability management
- [ ] Compliance reporting (SOC2, ISO27001)

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
http://auth-service:8080/metrics
http://gateway-service:8081/metrics  
http://backend-service:8082/metrics

# Custom PQC metrics
pqc_signature_operations_total
pqc_key_operations_total
pqc_signature_verification_duration_seconds
```

### Environment Variables
```yaml
# Service discovery
AUTH_SERVICE_URL: "http://auth-service.quantum-safe-mesh.svc.cluster.local:8080"
BACKEND_SERVICE_URL: "http://backend-service.quantum-safe-mesh.svc.cluster.local:8082"

# Pod information
POD_NAMESPACE: valueFrom fieldRef metadata.namespace
NODE_NAME: valueFrom fieldRef spec.nodeName
```

---

**🌟 This demo showcases the future of cryptographic security in cloud-native distributed systems. Deploy, scale, and prepare for the quantum era with Kubernetes!**