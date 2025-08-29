# Kubernetes Deployment Guide

## 🚀 Quick Start

### 1-Minute Deployment
```bash
# Clone repository
git clone <repo-url>
cd quantum-safe-mesh

# Make script executable  
chmod +x scripts/deploy.sh

# Deploy everything (builds images, deploys to k8s, runs demo)
./scripts/deploy.sh all
```

## 📋 Prerequisites

### Required Tools
- **Docker**: For building container images
- **kubectl**: Configured to access your Kubernetes cluster
- **Kubernetes Cluster**: One of the following:
  - Local: kind, minikube, Docker Desktop
  - Cloud: EKS, GKE, AKS
  - On-premise: Kubeadm, OpenShift

### Optional Tools
- **Helm 3.x**: For chart-based deployment (recommended)
- **jq**: For JSON processing in demo scripts

### Cluster Requirements
- **Kubernetes Version**: 1.20+
- **Resources**: Minimum 2 CPU cores, 4GB RAM
- **Features**: NetworkPolicy support (optional)

## 🛠️ Deployment Options

### Option 1: Automated Script (Recommended)
```bash
# Complete deployment
./scripts/deploy.sh all

# Individual steps
./scripts/deploy.sh build    # Build images
./scripts/deploy.sh helm     # Deploy with Helm  
./scripts/deploy.sh demo     # Run demo
./scripts/deploy.sh status   # Check status
./scripts/deploy.sh cleanup  # Remove everything
```

### Option 2: Helm Chart
```bash
# Install with default values
helm install quantum-safe-demo helm/quantum-safe-mesh \
  --namespace quantum-safe-mesh \
  --create-namespace \
  --wait

# Custom values
helm install quantum-safe-demo helm/quantum-safe-mesh \
  --namespace quantum-safe-mesh \
  --create-namespace \
  --values custom-values.yaml \
  --wait

# Upgrade
helm upgrade quantum-safe-demo helm/quantum-safe-mesh \
  --namespace quantum-safe-mesh \
  --wait
```

### Option 3: Raw Kubernetes Manifests
```bash
# Apply manifests
kubectl apply -f k8s/

# Wait for deployment
kubectl wait --for=condition=available --timeout=300s \
  deployment --all -n quantum-safe-mesh

# Run demo
kubectl apply -f k8s/demo-job.yaml
```

## 🔧 Configuration

### Helm Values Customization

Create a `custom-values.yaml`:
```yaml
# Scaling
authService:
  replicaCount: 3
gatewayService:
  replicaCount: 5
backendService:
  replicaCount: 4

# Ingress
gatewayService:
  ingress:
    enabled: true
    host: "quantum-mesh.example.com"
    className: "nginx"

# Monitoring
monitoring:
  enabled: true
  prometheus:
    enabled: true
  grafana:
    enabled: true

# Persistent storage
storage:
  persistent: true
  storageClass: "fast-ssd"
  size: "5Gi"

# Resources
authService:
  resources:
    requests:
      memory: "128Mi"
      cpu: "100m"
    limits:
      memory: "256Mi" 
      cpu: "500m"
```

### Environment Variables
```bash
# Image configuration
export DOCKER_REGISTRY="your-registry.com"
export IMAGE_TAG="v1.0.0"

# Deploy with custom images
./scripts/deploy.sh helm
```

## 📊 Monitoring Setup

### Prometheus Integration
```yaml
# ServiceMonitor for Prometheus Operator
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: quantum-safe-mesh
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: quantum-safe-mesh
  endpoints:
  - port: http
    path: /metrics
```

### Grafana Dashboard
The Helm chart includes a pre-configured Grafana dashboard showing:
- Service health status
- Request rates and latency
- PQC signature operations
- Error rates
- Key generation metrics

## 🔐 Security Configuration

### Network Policies
```yaml
# Enable zero-trust networking
networkPolicy:
  enabled: true
  zeroTrust: true
```

### Pod Security
```yaml
# Security contexts are enabled by default
securityContext:
  allowPrivilegeEscalation: false
  runAsNonRoot: true
  runAsUser: 1000
  capabilities:
    drop: [ALL]
```

### RBAC
```yaml
# Service account with minimal permissions
rbac:
  create: true
serviceAccount:
  create: true
  name: "quantum-safe-mesh"
```

## 🌐 External Access

### LoadBalancer
```yaml
gatewayService:
  service:
    type: LoadBalancer
    annotations:
      # AWS
      service.beta.kubernetes.io/aws-load-balancer-type: "nlb"
      
      # GCP
      cloud.google.com/load-balancer-type: "External"
      
      # Azure
      service.beta.kubernetes.io/azure-load-balancer-internal: "false"
```

### Ingress
```yaml
gatewayService:
  ingress:
    enabled: true
    className: "nginx"  # or "traefik", "istio"
    host: "quantum-safe-mesh.local"
    annotations:
      cert-manager.io/cluster-issuer: "letsencrypt-prod"
    tls:
    - secretName: quantum-mesh-tls
      hosts:
      - quantum-safe-mesh.local
```

### Port Forwarding (Development)
```bash
# Gateway service
kubectl port-forward svc/gateway-service 8081:8081 -n quantum-safe-mesh

# Access demo
curl http://localhost:8081/health
```

## 📈 Scaling

### Horizontal Pod Autoscaling
```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: gateway-service-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: gateway-service
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
```

### Manual Scaling
```bash
# Scale gateway service
kubectl scale deployment gateway-service --replicas=5 -n quantum-safe-mesh

# Scale all services
kubectl scale deployment --all --replicas=3 -n quantum-safe-mesh
```

## 🧪 Testing & Validation

### Health Checks
```bash
# All services health
kubectl get pods -n quantum-safe-mesh

# Detailed status
./scripts/deploy.sh status
```

### Demo Execution
```bash
# Run full demo
./scripts/deploy.sh demo

# Check demo logs
kubectl logs job/quantum-safe-demo -n quantum-safe-mesh
```

### Load Testing
```bash
# Port forward gateway
kubectl port-forward svc/gateway-service 8081:8081 -n quantum-safe-mesh &

# Run load test
for i in {1..100}; do
  curl -X POST http://localhost:8081/echo \
    -H "Content-Type: application/json" \
    -d '{"test":"load-'$i'"}' &
done
wait
```

## 🔍 Troubleshooting

### Common Issues

#### ImagePullBackOff
```bash
# Check image availability
docker images | grep quantum-safe-mesh

# For kind clusters, load images
kind load docker-image quantum-safe-mesh/auth:latest
```

#### Service Connection Issues
```bash
# Check service DNS resolution
kubectl run -it --rm debug --image=busybox --restart=Never -- \
  nslookup auth-service.quantum-safe-mesh.svc.cluster.local

# Test service connectivity
kubectl run -it --rm debug --image=curlimages/curl --restart=Never -- \
  curl http://auth-service.quantum-safe-mesh.svc.cluster.local:8080/health
```

#### PQC Key Generation Issues
```bash
# Check container logs
kubectl logs deployment/auth-service -n quantum-safe-mesh

# Check for sufficient entropy
kubectl exec deployment/auth-service -n quantum-safe-mesh -- \
  cat /proc/sys/kernel/random/entropy_avail
```

### Debug Commands
```bash
# Describe resources
kubectl describe pods -n quantum-safe-mesh
kubectl describe services -n quantum-safe-mesh
kubectl describe ingress -n quantum-safe-mesh

# Check events
kubectl get events --sort-by=.metadata.creationTimestamp -n quantum-safe-mesh

# Resource usage
kubectl top pods -n quantum-safe-mesh
kubectl top nodes
```

## 🗑️ Cleanup

### Complete Removal
```bash
# Using script
./scripts/deploy.sh cleanup

# Manual cleanup
helm uninstall quantum-safe-demo -n quantum-safe-mesh
kubectl delete namespace quantum-safe-mesh

# Clean images
docker rmi quantum-safe-mesh/auth:latest
docker rmi quantum-safe-mesh/gateway:latest
docker rmi quantum-safe-mesh/backend:latest
docker rmi quantum-safe-mesh/demo:latest
```

## 📚 Further Reading

- [Kubernetes Official Documentation](https://kubernetes.io/docs/)
- [Helm Documentation](https://helm.sh/docs/)
- [NIST Post-Quantum Cryptography](https://csrc.nist.gov/projects/post-quantum-cryptography)
- [Service Mesh Patterns](https://www.oreilly.com/library/view/service-mesh-patterns/9781492086444/)

---

**Need help?** Check the main [README.md](./README.md) or open an issue!