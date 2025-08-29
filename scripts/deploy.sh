#!/bin/bash

# Quantum-Safe Service Mesh Deployment Script
# This script automates the complete deployment of the quantum-safe service mesh on Kubernetes

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
NAMESPACE="quantum-safe-mesh"
CHART_NAME="quantum-safe-mesh"
RELEASE_NAME="quantum-safe-demo"
DOCKER_REGISTRY=${DOCKER_REGISTRY:-""}
IMAGE_TAG=${IMAGE_TAG:-"latest"}

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

check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check if kubectl is available
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl is not installed or not in PATH"
        exit 1
    fi
    
    # Check if docker is available
    if ! command -v docker &> /dev/null; then
        log_error "docker is not installed or not in PATH"
        exit 1
    fi
    
    # Check if helm is available
    if ! command -v helm &> /dev/null; then
        log_error "helm is not installed or not in PATH"
        exit 1
    fi
    
    # Check kubectl connection
    if ! kubectl cluster-info &> /dev/null; then
        log_error "Cannot connect to Kubernetes cluster"
        exit 1
    fi
    
    log_success "All prerequisites satisfied"
}

build_images() {
    log_info "Building Docker images..."
    
    local images=("auth" "gateway" "backend" "demo")
    
    for image in "${images[@]}"; do
        log_info "Building $image service image..."
        docker build -f docker/Dockerfile.$image -t quantum-safe-mesh/$image:$IMAGE_TAG .
        
        if [ $? -eq 0 ]; then
            log_success "Built quantum-safe-mesh/$image:$IMAGE_TAG"
        else
            log_error "Failed to build quantum-safe-mesh/$image:$IMAGE_TAG"
            exit 1
        fi
    done
    
    log_success "All Docker images built successfully"
}

load_images_kind() {
    if kubectl get nodes -o jsonpath='{.items[0].status.nodeInfo.containerRuntimeVersion}' | grep -q containerd; then
        log_info "Loading images into kind cluster..."
        
        local images=("auth" "gateway" "backend" "demo")
        
        for image in "${images[@]}"; do
            log_info "Loading quantum-safe-mesh/$image:$IMAGE_TAG into kind..."
            kind load docker-image quantum-safe-mesh/$image:$IMAGE_TAG
        done
        
        log_success "All images loaded into kind cluster"
    fi
}

deploy_with_kubectl() {
    log_info "Deploying with kubectl..."
    
    # Create namespace
    kubectl apply -f k8s/namespace.yaml
    
    # Deploy services in order
    log_info "Deploying Auth Service..."
    kubectl apply -f k8s/auth-service.yaml
    
    log_info "Waiting for Auth Service to be ready..."
    kubectl wait --for=condition=available --timeout=120s deployment/auth-service -n $NAMESPACE
    
    log_info "Deploying Gateway Service..."
    kubectl apply -f k8s/gateway-service.yaml
    
    log_info "Deploying Backend Service..."
    kubectl apply -f k8s/backend-service.yaml
    
    log_info "Waiting for all services to be ready..."
    kubectl wait --for=condition=available --timeout=120s deployment/gateway-service -n $NAMESPACE
    kubectl wait --for=condition=available --timeout=120s deployment/backend-service -n $NAMESPACE
    
    # Deploy network policies
    if kubectl get crd networkpolicies.networking.k8s.io &> /dev/null; then
        log_info "Deploying Network Policies..."
        kubectl apply -f k8s/network-policy.yaml
    else
        log_warning "NetworkPolicy CRD not found, skipping network policies"
    fi
    
    log_success "Deployment completed successfully"
}

deploy_with_helm() {
    log_info "Deploying with Helm..."
    
    # Check if release exists
    if helm list -n $NAMESPACE | grep -q $RELEASE_NAME; then
        log_info "Upgrading existing Helm release..."
        helm upgrade $RELEASE_NAME helm/$CHART_NAME \
            --namespace $NAMESPACE \
            --set authService.image.tag=$IMAGE_TAG \
            --set gatewayService.image.tag=$IMAGE_TAG \
            --set backendService.image.tag=$IMAGE_TAG \
            --set demoJob.image.tag=$IMAGE_TAG \
            --wait \
            --timeout=300s
    else
        log_info "Installing new Helm release..."
        helm install $RELEASE_NAME helm/$CHART_NAME \
            --namespace $NAMESPACE \
            --create-namespace \
            --set authService.image.tag=$IMAGE_TAG \
            --set gatewayService.image.tag=$IMAGE_TAG \
            --set backendService.image.tag=$IMAGE_TAG \
            --set demoJob.image.tag=$IMAGE_TAG \
            --wait \
            --timeout=300s
    fi
    
    log_success "Helm deployment completed successfully"
}

run_demo() {
    log_info "Running demonstration..."
    
    # Delete any existing demo job
    kubectl delete job quantum-safe-demo -n $NAMESPACE --ignore-not-found=true
    
    # Wait a moment for cleanup
    sleep 5
    
    # Run demo job
    kubectl apply -f k8s/demo-job.yaml
    
    # Wait for job to complete and show logs
    log_info "Waiting for demo to complete..."
    kubectl wait --for=condition=complete --timeout=180s job/quantum-safe-demo -n $NAMESPACE
    
    log_info "Demo results:"
    kubectl logs job/quantum-safe-demo -n $NAMESPACE
    
    log_success "Demo completed successfully"
}

show_status() {
    log_info "Showing deployment status..."
    
    echo ""
    log_info "Namespace: $NAMESPACE"
    kubectl get namespace $NAMESPACE
    
    echo ""
    log_info "Pods:"
    kubectl get pods -n $NAMESPACE -o wide
    
    echo ""
    log_info "Services:"
    kubectl get services -n $NAMESPACE
    
    echo ""
    log_info "Ingresses:"
    kubectl get ingress -n $NAMESPACE
    
    # Get external access information
    echo ""
    log_info "Access Information:"
    
    # Try to get LoadBalancer external IP
    EXTERNAL_IP=$(kubectl get svc gateway-service -n $NAMESPACE -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
    if [ -z "$EXTERNAL_IP" ]; then
        EXTERNAL_IP=$(kubectl get svc gateway-service -n $NAMESPACE -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')
    fi
    
    if [ ! -z "$EXTERNAL_IP" ]; then
        echo "Gateway LoadBalancer: http://$EXTERNAL_IP:8081"
    else
        echo "Gateway Service: Use 'kubectl port-forward svc/gateway-service 8081:8081 -n $NAMESPACE'"
    fi
    
    # Ingress information
    INGRESS_HOST=$(kubectl get ingress gateway-ingress -n $NAMESPACE -o jsonpath='{.spec.rules[0].host}' 2>/dev/null || echo "")
    if [ ! -z "$INGRESS_HOST" ]; then
        echo "Ingress Host: http://$INGRESS_HOST"
        echo "Add to /etc/hosts: <ingress-controller-ip> $INGRESS_HOST"
    fi
}

cleanup() {
    log_warning "Cleaning up deployment..."
    
    # Delete with helm if release exists
    if helm list -n $NAMESPACE | grep -q $RELEASE_NAME; then
        log_info "Uninstalling Helm release..."
        helm uninstall $RELEASE_NAME -n $NAMESPACE
    else
        log_info "Deleting Kubernetes resources..."
        kubectl delete -f k8s/ --ignore-not-found=true
    fi
    
    # Delete namespace
    kubectl delete namespace $NAMESPACE --ignore-not-found=true
    
    log_success "Cleanup completed"
}

print_help() {
    echo "Quantum-Safe Service Mesh Deployment Script"
    echo ""
    echo "Usage: $0 [command]"
    echo ""
    echo "Commands:"
    echo "  build       Build Docker images"
    echo "  deploy      Deploy using kubectl"
    echo "  helm        Deploy using Helm"
    echo "  demo        Run demonstration"
    echo "  status      Show deployment status"
    echo "  cleanup     Remove deployment"
    echo "  all         Build, deploy, and run demo"
    echo "  help        Show this help message"
    echo ""
    echo "Environment Variables:"
    echo "  DOCKER_REGISTRY   Docker registry prefix"
    echo "  IMAGE_TAG         Image tag (default: latest)"
    echo ""
    echo "Examples:"
    echo "  $0 all                    # Complete deployment"
    echo "  $0 build                  # Build images only"
    echo "  $0 helm                   # Deploy with Helm"
    echo "  IMAGE_TAG=v1.0 $0 deploy  # Deploy with specific tag"
}

# Main script logic
case "${1:-help}" in
    "build")
        check_prerequisites
        build_images
        load_images_kind
        ;;
    "deploy")
        check_prerequisites
        build_images
        load_images_kind
        deploy_with_kubectl
        show_status
        ;;
    "helm")
        check_prerequisites
        build_images
        load_images_kind
        deploy_with_helm
        show_status
        ;;
    "demo")
        run_demo
        ;;
    "status")
        show_status
        ;;
    "cleanup")
        cleanup
        ;;
    "all")
        check_prerequisites
        build_images
        load_images_kind
        deploy_with_helm
        run_demo
        show_status
        ;;
    "help"|*)
        print_help
        ;;
esac