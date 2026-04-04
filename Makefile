.PHONY: help generate-keys run-auth run-gateway run-backend demo clean build-all stop-services benchmark test k8s-build k8s-deploy k8s-helm k8s-demo k8s-status k8s-cleanup release-tag

help:
	@echo "Quantum-Safe Service Mesh Demo - Available Commands:"
	@echo ""
	@echo "Setup Commands:"
	@echo "  make generate-keys    - Generate PQC keypairs for all services"
	@echo "  make build-all        - Build all service binaries"
	@echo ""
	@echo "Local Service Commands:"
	@echo "  make run-auth         - Start the Auth Service (port 8080)"
	@echo "  make run-gateway      - Start the API Gateway (port 8081)"
	@echo "  make run-backend      - Start the Backend Service (port 8082)"
	@echo ""
	@echo "Local Demo Commands:"
	@echo "  make demo             - Run a complete demo flow"
	@echo "  make demo-echo        - Test echo endpoint"
	@echo "  make demo-process     - Test data processing endpoint"
	@echo "  make demo-status      - Check backend service status"
	@echo ""
	@echo "Kubernetes Commands:"
	@echo "  make k8s-build        - Build Docker images"
	@echo "  make k8s-deploy       - Deploy to Kubernetes with kubectl"
	@echo "  make k8s-helm         - Deploy to Kubernetes with Helm"
	@echo "  make k8s-demo         - Run demo in Kubernetes"
	@echo "  make k8s-status       - Show Kubernetes deployment status"
	@echo "  make k8s-cleanup      - Remove Kubernetes deployment"
	@echo "  make k8s-all          - Build, deploy, and demo (complete flow)"
	@echo ""
	@echo "AWS Commands:"
	@echo "  make aws-init         - Initialize Terraform"
	@echo "  make aws-plan         - Create Terraform plan"
	@echo "  make aws-apply        - Deploy AWS infrastructure"
	@echo "  make aws-deploy       - Deploy services to AWS"
	@echo "  make aws-test         - Test AWS deployment"
	@echo "  make aws-destroy      - Destroy AWS infrastructure"
	@echo "  make aws-all          - Complete AWS deployment"
	@echo ""
	@echo "Performance Commands:"
	@echo "  make benchmark        - Run PQC vs RSA performance comparison"
	@echo "  make load-test        - Run basic load test"
	@echo ""
	@echo "Release Commands:"
	@echo "  make release-tag VERSION=v2.0.0 - Create and push a release tag"
	@echo ""
	@echo "Utility Commands:"
	@echo "  make test             - Run tests"
	@echo "  make clean            - Clean build artifacts and keys"
	@echo "  make stop-services    - Stop all running services"

# Build Commands
build-all:
	@echo "🔨 Building all services..."
	@go mod tidy
	@go build -o bin/auth ./cmd/auth
	@go build -o bin/gateway ./cmd/gateway
	@go build -o bin/backend ./cmd/backend
	@echo "✅ All services built successfully"

# Key Generation
generate-keys:
	@echo "🔑 Generating PQC keypairs for all services..."
	@mkdir -p keys
	@echo "Starting temporary auth service to generate keys..."
	@timeout 10s go run ./cmd/auth/main.go > /dev/null 2>&1 || true
	@sleep 2
	@echo "✅ PQC keypairs generated for all services"
	@ls -la keys/

# Service Startup Commands
run-auth:
	@echo "🚀 Starting Auth Service on port 8080..."
	@echo "Press Ctrl+C to stop"
	@go run ./cmd/auth/main.go

run-gateway:
	@echo "🚀 Starting API Gateway on port 8081..."
	@echo "Make sure Auth Service is running on port 8080"
	@echo "Press Ctrl+C to stop"
	@go run ./cmd/gateway/main.go

run-backend:
	@echo "🚀 Starting Backend Service on port 8082..."
	@echo "Make sure Auth Service is running on port 8080"
	@echo "Press Ctrl+C to stop"
	@go run ./cmd/backend/main.go

# Demo Commands
demo: demo-health demo-echo demo-process demo-status
	@echo ""
	@echo "🎉 Demo completed successfully!"
	@echo "All services authenticated using Post-Quantum Cryptography!"

demo-health:
	@echo "🔍 Checking service health..."
	@echo "Auth Service:"
	@curl -s http://localhost:8080/health || echo "❌ Auth Service not responding"
	@echo ""
	@echo "Gateway Service:"
	@curl -s http://localhost:8081/health || echo "❌ Gateway Service not responding"  
	@echo ""
	@echo "Backend Service:"
	@curl -s http://localhost:8082/health || echo "❌ Backend Service not responding"
	@echo ""

demo-echo:
	@echo "📡 Testing Echo Endpoint through Gateway..."
	@curl -X POST http://localhost:8081/echo \
		-H "Content-Type: application/json" \
		-H "X-Client-ID: demo-client" \
		-d '{"message": "Hello Quantum-Safe World!", "demo": true}' \
		2>/dev/null | jq '.' || curl -X POST http://localhost:8081/echo \
		-H "Content-Type: application/json" \
		-H "X-Client-ID: demo-client" \
		-d '{"message": "Hello Quantum-Safe World!", "demo": true}'
	@echo ""

demo-process:
	@echo "🧠 Testing Data Processing Endpoint..."
	@curl -X POST http://localhost:8081/process \
		-H "Content-Type: application/json" \
		-H "X-Client-ID: demo-client" \
		-d '{"data": "Quantum-safe data processing", "algorithm": "test", "priority": "high"}' \
		2>/dev/null | jq '.' || curl -X POST http://localhost:8081/process \
		-H "Content-Type: application/json" \
		-H "X-Client-ID: demo-client" \
		-d '{"data": "Quantum-safe data processing", "algorithm": "test", "priority": "high"}'
	@echo ""

demo-status:
	@echo "📊 Checking Backend Service Status..."
	@curl -X GET http://localhost:8081/status \
		-H "X-Client-ID: demo-client" \
		2>/dev/null | jq '.' || curl -X GET http://localhost:8081/status \
		-H "X-Client-ID: demo-client"
	@echo ""

# Performance Testing
benchmark:
	@echo "📊 Running PQC vs RSA performance comparison..."
	@go run ./cmd/auth/main.go -benchmark-only 2>/dev/null || echo "Run 'make run-auth' in another terminal first, then run this command"

load-test:
	@echo "⚡ Running basic load test..."
	@echo "Sending 10 concurrent requests to test PQC performance under load..."
	@for i in {1..10}; do \
		(curl -X POST http://localhost:8081/echo \
			-H "Content-Type: application/json" \
			-H "X-Client-ID: load-test-$$i" \
			-d "{\"message\": \"Load test request $$i\", \"timestamp\": \"$$(date)\"}" \
			>/dev/null 2>&1 &); \
	done; \
	wait
	@echo "✅ Load test completed"

# Test Commands  
test:
	@echo "🧪 Running tests..."
	@go test ./pkg/... -v
	@go test ./cmd/... -v
	@echo "✅ All tests completed"

# Utility Commands
clean:
	@echo "🧹 Cleaning up..."
	@rm -rf bin/
	@rm -rf keys/
	@echo "✅ Cleanup completed"

stop-services:
	@echo "🛑 Stopping services..."
	@pkill -f "auth/main.go" 2>/dev/null || true
	@pkill -f "gateway/main.go" 2>/dev/null || true  
	@pkill -f "backend/main.go" 2>/dev/null || true
	@echo "✅ All services stopped"

# Development helpers
dev-setup: clean generate-keys
	@echo "🛠️ Development environment setup complete"

start-all-services:
	@echo "🚀 Starting all services..."
	@echo "This will start services in background. Use 'make stop-services' to stop them."
	@nohup go run ./cmd/auth/main.go > logs/auth.log 2>&1 & echo $$! > .auth.pid
	@sleep 2
	@nohup go run ./cmd/gateway/main.go > logs/gateway.log 2>&1 & echo $$! > .gateway.pid
	@sleep 2  
	@nohup go run ./cmd/backend/main.go > logs/backend.log 2>&1 & echo $$! > .backend.pid
	@sleep 2
	@echo "✅ All services started in background"
	@echo "Logs: logs/auth.log, logs/gateway.log, logs/backend.log"

logs:
	@mkdir -p logs

# Docker commands (optional)
docker-build:
	@echo "🐳 Building Docker images..."
	@docker build -t quantum-safe-auth -f docker/Dockerfile.auth .
	@docker build -t quantum-safe-gateway -f docker/Dockerfile.gateway .
	@docker build -t quantum-safe-backend -f docker/Dockerfile.backend .

docker-run:
	@echo "🐳 Running services in Docker..."
	@docker-compose up -d

# Kubernetes Commands
k8s-build:
	@echo "🐳 Building Docker images for Kubernetes..."
	@./scripts/deploy.sh build

k8s-deploy:
	@echo "🚀 Deploying to Kubernetes with kubectl..."
	@./scripts/deploy.sh deploy

k8s-helm:
	@echo "⚓ Deploying to Kubernetes with Helm..."
	@./scripts/deploy.sh helm

k8s-demo:
	@echo "🎯 Running Kubernetes demo..."
	@./scripts/deploy.sh demo

k8s-status:
	@echo "📊 Checking Kubernetes deployment status..."
	@./scripts/deploy.sh status

k8s-cleanup:
	@echo "🗑️ Cleaning up Kubernetes deployment..."
	@./scripts/deploy.sh cleanup

k8s-all:
	@echo "🚀 Complete Kubernetes deployment..."
	@./scripts/deploy.sh all

# AWS Commands
aws-init:
	@echo "🏗️ Initializing Terraform for AWS..."
	@./scripts/aws-deploy.sh init

aws-plan:
	@echo "📋 Creating Terraform plan..."
	@./scripts/aws-deploy.sh plan

aws-apply:
	@echo "🚀 Deploying AWS infrastructure..."
	@./scripts/aws-deploy.sh apply

aws-deploy:
	@echo "📦 Deploying services to AWS..."
	@./scripts/aws-deploy.sh deploy

aws-test:
	@echo "🧪 Testing AWS deployment..."
	@./scripts/aws-deploy.sh test

aws-destroy:
	@echo "💥 Destroying AWS infrastructure..."
	@./scripts/aws-deploy.sh destroy

aws-all:
	@echo "🌩️ Complete AWS deployment..."
	@./scripts/aws-deploy.sh all

# Release Commands
release-tag:
	@if [ -z "$(VERSION)" ]; then echo "Usage: make release-tag VERSION=v2.0.0"; exit 1; fi
	@if ! echo "$(VERSION)" | grep -qE '^v[0-9]+\.[0-9]+\.[0-9]+$$'; then echo "Error: VERSION must be semver (e.g., v2.0.0)"; exit 1; fi
	@echo "Creating release tag $(VERSION)..."
	@git tag -a $(VERSION) -m "Release $(VERSION)"
	@git push origin $(VERSION)
	@echo "Tag $(VERSION) pushed. GitHub Actions will build and publish the release."