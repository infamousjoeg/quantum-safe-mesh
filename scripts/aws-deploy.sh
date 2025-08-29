#!/bin/bash

# AWS Deployment Script for Quantum-Safe Service Mesh
# This script automates the complete deployment on AWS using Terraform

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
TERRAFORM_DIR="terraform"
AWS_REGION=${AWS_REGION:-"us-west-2"}
ENVIRONMENT=${ENVIRONMENT:-"dev"}
PROJECT_NAME=${PROJECT_NAME:-"quantum-safe-mesh"}

# Functions
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

check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check if terraform is available
    if ! command -v terraform &> /dev/null; then
        log_error "Terraform is not installed or not in PATH"
        log_info "Install from: https://www.terraform.io/downloads"
        exit 1
    fi
    
    # Check if AWS CLI is available
    if ! command -v aws &> /dev/null; then
        log_error "AWS CLI is not installed or not in PATH"
        log_info "Install from: https://docs.aws.amazon.com/cli/latest/userguide/install-cliv2.html"
        exit 1
    fi
    
    # Check AWS credentials
    if ! aws sts get-caller-identity &> /dev/null; then
        log_error "AWS credentials not configured"
        log_info "Run: aws configure"
        exit 1
    fi
    
    # Check if kubectl is available
    if ! command -v kubectl &> /dev/null; then
        log_warning "kubectl is not installed - some operations will not be available"
    fi
    
    log_success "Prerequisites check completed"
}

terraform_init() {
    log_info "Initializing Terraform..."
    
    cd $TERRAFORM_DIR
    terraform init
    
    log_success "Terraform initialized"
}

terraform_plan() {
    log_info "Creating Terraform plan..."
    
    cd $TERRAFORM_DIR
    terraform plan \
        -var="aws_region=$AWS_REGION" \
        -var="environment=$ENVIRONMENT" \
        -var="project_name=$PROJECT_NAME" \
        -out=tfplan
    
    log_success "Terraform plan created"
}

terraform_apply() {
    log_info "Applying Terraform configuration..."
    
    cd $TERRAFORM_DIR
    terraform apply tfplan
    
    log_success "Infrastructure deployed successfully"
}

terraform_destroy() {
    log_warning "Destroying infrastructure..."
    
    cd $TERRAFORM_DIR
    terraform destroy \
        -var="aws_region=$AWS_REGION" \
        -var="environment=$ENVIRONMENT" \
        -var="project_name=$PROJECT_NAME" \
        -auto-approve
    
    log_success "Infrastructure destroyed"
}

get_outputs() {
    log_info "Getting Terraform outputs..."
    
    cd $TERRAFORM_DIR
    
    # Get key outputs
    INSTANCE_IP=$(terraform output -raw instance_public_ip 2>/dev/null || echo "")
    PRIVATE_KEY=$(terraform output -raw private_key_path 2>/dev/null || echo "")
    SSH_COMMAND=$(terraform output -raw ssh_command 2>/dev/null || echo "")
    GATEWAY_URL=$(terraform output -raw gateway_service_url 2>/dev/null || echo "")
    
    if [ ! -z "$INSTANCE_IP" ]; then
        echo ""
        log_header "=== Deployment Information ==="
        echo "Instance IP: $INSTANCE_IP"
        echo "Private Key: $PRIVATE_KEY"
        echo "Gateway URL: $GATEWAY_URL"
        echo ""
        echo "SSH Command:"
        echo "$SSH_COMMAND"
        echo ""
        
        # Show all outputs
        log_info "All outputs:"
        terraform output
    fi
}

deploy_services() {
    log_info "Deploying quantum-safe services to EC2 instance..."
    
    cd $TERRAFORM_DIR
    
    INSTANCE_IP=$(terraform output -raw instance_public_ip 2>/dev/null || echo "")
    PRIVATE_KEY=$(terraform output -raw private_key_path 2>/dev/null || echo "")
    
    if [ -z "$INSTANCE_IP" ] || [ -z "$PRIVATE_KEY" ]; then
        log_error "Could not get instance information from Terraform"
        exit 1
    fi
    
    log_info "Copying deployment files to EC2 instance..."
    
    # Wait for instance to be ready
    log_info "Waiting for instance to be ready..."
    for i in {1..30}; do
        if ssh -i "$PRIVATE_KEY" -o ConnectTimeout=10 -o StrictHostKeyChecking=no ubuntu@$INSTANCE_IP "echo 'Instance ready'" 2>/dev/null; then
            break
        fi
        echo "Attempt $i/30: Instance not ready, waiting 10s..."
        sleep 10
    done
    
    # Copy project files
    log_info "Copying project files..."
    scp -i "$PRIVATE_KEY" -o StrictHostKeyChecking=no -r ../{pkg,cmd,helm,k8s,scripts,go.mod,go.sum,README.md,Makefile} ubuntu@$INSTANCE_IP:/home/ubuntu/quantum-safe-mesh/
    
    # Deploy services
    log_info "Deploying services on EC2..."
    ssh -i "$PRIVATE_KEY" -o StrictHostKeyChecking=no ubuntu@$INSTANCE_IP << 'EOF'
cd /home/ubuntu/quantum-safe-mesh
chmod +x scripts/*.sh
./deploy-quantum-safe-mesh.sh
EOF
    
    log_success "Services deployed successfully"
}

setup_monitoring() {
    log_info "Setting up monitoring on EC2 instance..."
    
    cd $TERRAFORM_DIR
    
    INSTANCE_IP=$(terraform output -raw instance_public_ip 2>/dev/null || echo "")
    PRIVATE_KEY=$(terraform output -raw private_key_path 2>/dev/null || echo "")
    
    if [ -z "$INSTANCE_IP" ] || [ -z "$PRIVATE_KEY" ]; then
        log_error "Could not get instance information from Terraform"
        exit 1
    fi
    
    ssh -i "$PRIVATE_KEY" -o StrictHostKeyChecking=no ubuntu@$INSTANCE_IP << 'EOF'
/home/ubuntu/setup-monitoring.sh
EOF
    
    log_success "Monitoring setup completed"
}

run_tests() {
    log_info "Running tests on deployed services..."
    
    cd $TERRAFORM_DIR
    
    INSTANCE_IP=$(terraform output -raw instance_public_ip 2>/dev/null || echo "")
    GATEWAY_URL=$(terraform output -raw gateway_service_url 2>/dev/null || echo "")
    
    if [ -z "$GATEWAY_URL" ]; then
        log_error "Could not get gateway URL"
        exit 1
    fi
    
    # Wait for services to be ready
    log_info "Waiting for services to be ready..."
    for i in {1..60}; do
        if curl -f -s "$GATEWAY_URL/health" > /dev/null 2>&1; then
            log_success "Services are ready"
            break
        fi
        echo "Attempt $i/60: Services not ready, waiting 10s..."
        sleep 10
    done
    
    # Run basic tests
    log_info "Running health check..."
    if curl -f -s "$GATEWAY_URL/health" > /dev/null; then
        log_success "Health check passed"
    else
        log_error "Health check failed"
    fi
    
    log_info "Running echo test..."
    response=$(curl -s -w "%{http_code}" -X POST "$GATEWAY_URL/echo" \
        -H "Content-Type: application/json" \
        -H "X-Client-ID: aws-test" \
        -d '{"message": "AWS deployment test"}' \
        -o /tmp/response.json)
    
    if [[ "$response" == "200" ]]; then
        log_success "Echo test passed"
        cat /tmp/response.json | jq '.' 2>/dev/null || cat /tmp/response.json
    else
        log_error "Echo test failed (HTTP $response)"
    fi
    
    rm -f /tmp/response.json
}

show_access_info() {
    log_header "=== Access Information ==="
    
    cd $TERRAFORM_DIR
    
    INSTANCE_IP=$(terraform output -raw instance_public_ip 2>/dev/null || echo "")
    PRIVATE_KEY=$(terraform output -raw private_key_path 2>/dev/null || echo "")
    GATEWAY_URL=$(terraform output -raw gateway_service_url 2>/dev/null || echo "")
    
    echo ""
    echo "🌐 Service URLs:"
    echo "  Gateway Service: $GATEWAY_URL"
    echo "  Prometheus: http://$INSTANCE_IP:30090"
    echo "  Grafana: http://$INSTANCE_IP:30300 (admin/admin123)"
    echo ""
    echo "🔗 SSH Access:"
    echo "  ssh -i $PRIVATE_KEY ubuntu@$INSTANCE_IP"
    echo ""
    echo "📋 Quick Tests:"
    echo "  curl $GATEWAY_URL/health"
    echo "  curl -X POST $GATEWAY_URL/echo -H 'Content-Type: application/json' -d '{\"message\": \"test\"}'"
    echo ""
    echo "🔧 Management Commands:"
    echo "  # Check deployment status"
    echo "  ssh -i $PRIVATE_KEY ubuntu@$INSTANCE_IP 'cd quantum-safe-mesh && ./scripts/deploy.sh status'"
    echo ""
    echo "  # View logs"
    echo "  ssh -i $PRIVATE_KEY ubuntu@$INSTANCE_IP 'kubectl logs job/quantum-safe-demo -n quantum-safe-mesh'"
    echo ""
}

cleanup() {
    log_warning "Cleaning up local files..."
    
    cd $TERRAFORM_DIR
    rm -f tfplan quantum-safe-mesh-key.pem
    
    log_success "Cleanup completed"
}

print_help() {
    echo "AWS Deployment Script for Quantum-Safe Service Mesh"
    echo ""
    echo "Usage: $0 [command]"
    echo ""
    echo "Commands:"
    echo "  init        Initialize Terraform"
    echo "  plan        Create Terraform execution plan"
    echo "  apply       Deploy infrastructure to AWS"
    echo "  deploy      Deploy services to EC2 instance"
    echo "  monitor     Setup monitoring (Prometheus/Grafana)"
    echo "  test        Run tests against deployed services"
    echo "  info        Show access information"
    echo "  destroy     Destroy all AWS resources"
    echo "  cleanup     Clean up local files"
    echo "  all         Complete deployment (init + plan + apply + deploy)"
    echo "  help        Show this help message"
    echo ""
    echo "Environment Variables:"
    echo "  AWS_REGION      AWS region (default: us-west-2)"
    echo "  ENVIRONMENT     Environment name (default: dev)"
    echo "  PROJECT_NAME    Project name (default: quantum-safe-mesh)"
    echo ""
    echo "Examples:"
    echo "  $0 all                           # Complete deployment"
    echo "  AWS_REGION=us-east-1 $0 apply    # Deploy to us-east-1"
    echo "  ENVIRONMENT=prod $0 plan         # Plan for production"
}

# Main script logic
case "${1:-help}" in
    "init")
        check_prerequisites
        terraform_init
        ;;
    "plan")
        check_prerequisites
        terraform_plan
        ;;
    "apply")
        check_prerequisites
        terraform_init
        terraform_plan
        terraform_apply
        get_outputs
        show_access_info
        ;;
    "deploy")
        deploy_services
        ;;
    "monitor")
        setup_monitoring
        ;;
    "test")
        run_tests
        ;;
    "info")
        get_outputs
        show_access_info
        ;;
    "destroy")
        terraform_destroy
        cleanup
        ;;
    "cleanup")
        cleanup
        ;;
    "all")
        check_prerequisites
        terraform_init
        terraform_plan
        terraform_apply
        get_outputs
        deploy_services
        setup_monitoring
        run_tests
        show_access_info
        ;;
    "help"|*)
        print_help
        ;;
esac