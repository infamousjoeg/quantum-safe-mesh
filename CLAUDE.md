# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Architecture

This is a quantum-safe mesh network implementation showcasing Post-Quantum Cryptography (PQC) in a microservices architecture. The system demonstrates quantum-resistant authentication between three microservices:

- **Auth Service** (`cmd/auth/`): Central authentication authority managing service public keys and key exchange
- **API Gateway** (`cmd/gateway/`): Entry point validating and forwarding requests to backend services  
- **Backend Service** (`cmd/backend/`): Processing service handling business logic and returning signed responses

### Post-Quantum Cryptography

The system uses NIST-standardized PQC algorithms:
- **Dilithium3** (`pkg/pqc/dilithium.go`): Digital signatures for service identity verification
- **Kyber768** (`pkg/pqc/kyber.go`): Key encapsulation mechanism for establishing shared secrets

Keys are stored in the `keys/` directory with service-specific naming (e.g., `auth-service_dilithium.key`).

## Common Development Commands

### Local Development
```bash
make generate-keys    # Generate PQC keypairs for all services
make run-auth        # Start Auth Service (port 8080)
make run-gateway     # Start API Gateway (port 8081) 
make run-backend     # Start Backend Service (port 8082)
make demo           # Run complete demo flow
make test           # Run unit tests
make benchmark      # PQC vs RSA performance comparison
make clean          # Clean build artifacts and keys
```

### Kubernetes Deployment
```bash
./scripts/deploy.sh all      # Complete build, deploy, and demo
./scripts/deploy.sh build    # Build Docker images only
./scripts/deploy.sh helm     # Deploy with Helm
./scripts/deploy.sh deploy   # Deploy with kubectl
./scripts/deploy.sh cleanup  # Remove deployment
```

### AWS Deployment
```bash
./scripts/aws-deploy.sh all     # Complete AWS deployment
./scripts/aws-deploy.sh init    # Initialize Terraform
./scripts/aws-deploy.sh apply   # Deploy infrastructure
./scripts/aws-deploy.sh destroy # Destroy infrastructure
```

## Code Structure

### Key Packages
- `pkg/pqc/`: Post-quantum cryptographic operations (Dilithium/Kyber)
- `pkg/models/`: Shared data structures and types
- `cmd/`: Service entry points (auth, gateway, backend)

### Dependencies
- `github.com/cloudflare/circl`: Cloudflare's cryptographic library providing PQC implementations
- `github.com/gorilla/mux`: HTTP router for REST endpoints
- Go 1.22+ required

### Service Communication Flow
1. Services register their public keys with Auth Service
2. Gateway signs requests using Dilithium before forwarding to Backend
3. Backend verifies Gateway signature using Auth Service public key registry
4. Backend signs responses before returning to Gateway
5. Gateway verifies Backend signature before responding to client

### Key Management
- Keys are generated automatically on first service startup
- Key files are saved to `keys/` directory with service-specific names
- Services load existing keys or generate new ones if not found
- Use `make generate-keys` to pre-generate all service keys

## Testing and Validation

The system includes comprehensive testing endpoints:
- Health checks: `/health` on each service
- Echo test: `POST /echo` through Gateway
- Processing test: `POST /process` through Gateway
- Status check: `GET /status` through Gateway

All inter-service communication uses PQC signatures for authentication and integrity verification.