#!/bin/bash

# Quantum-Safe Service Mesh EC2 Setup Script
# This script sets up Docker, kind, kubectl, and deploys the quantum-safe service mesh

set -e

# Configuration from Terraform
AWS_REGION="${aws_region}"
CLUSTER_NAME="${cluster_name}"
ECR_REPOSITORY="${ecr_repository}"
SERVICES='${services}'

# Versions
KIND_VERSION="${kind_version}"
KUBECTL_VERSION="${kubectl_version}"
HELM_VERSION="${helm_version}"

# Logging
exec > >(tee -a /var/log/user-data.log)
exec 2>&1

echo "Starting quantum-safe-mesh setup at $(date)"
echo "Region: $AWS_REGION"
echo "Cluster: $CLUSTER_NAME"
echo "ECR: $ECR_REPOSITORY"

# Update system
apt-get update
apt-get upgrade -y

# Install required packages
apt-get install -y \
    curl \
    wget \
    git \
    jq \
    unzip \
    build-essential \
    ca-certificates \
    gnupg \
    lsb-release \
    awscli

# Install Docker
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null
apt-get update
apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

# Add ubuntu user to docker group
usermod -aG docker ubuntu

# Start and enable Docker
systemctl start docker
systemctl enable docker

# Install Go (needed for building the services)
GO_VERSION="1.21.0"
wget https://golang.org/dl/go$${GO_VERSION}.linux-amd64.tar.gz
tar -C /usr/local -xzf go$${GO_VERSION}.linux-amd64.tar.gz
rm go$${GO_VERSION}.linux-amd64.tar.gz

# Add Go to PATH for all users
echo 'export PATH=/usr/local/go/bin:$PATH' >> /etc/profile
echo 'export PATH=/usr/local/go/bin:$PATH' >> /home/ubuntu/.bashrc

# Install kubectl
curl -LO "https://dl.k8s.io/release/$${KUBECTL_VERSION}/bin/linux/amd64/kubectl"
chmod +x kubectl
mv kubectl /usr/local/bin/

# Install kind
curl -Lo ./kind https://kind.sigs.k8s.io/dl/$${KIND_VERSION}/kind-linux-amd64
chmod +x ./kind
mv ./kind /usr/local/bin/kind

# Install Helm
curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3
chmod 700 get_helm.sh
HELM_INSTALL_DIR="/usr/local/bin" ./get_helm.sh --version $${HELM_VERSION}
rm get_helm.sh

# Clone the quantum-safe-mesh repository
cd /home/ubuntu
git clone https://github.com/your-repo/quantum-safe-mesh.git || {
    echo "Repository not found, creating sample structure"
    mkdir -p quantum-safe-mesh
    cd quantum-safe-mesh
    
    # Create a minimal setup script for now
    cat > setup.sh << 'EOF'
#!/bin/bash
echo "Quantum-Safe Service Mesh setup placeholder"
echo "Repository content would be deployed here"
EOF
    chmod +x setup.sh
}

# Change ownership to ubuntu user
chown -R ubuntu:ubuntu /home/ubuntu/quantum-safe-mesh

# Create kind cluster configuration
cat > /home/ubuntu/kind-config.yaml << 'EOF'
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
name: ${cluster_name}
networking:
  apiServerAddress: "0.0.0.0"
  apiServerPort: 6443
nodes:
- role: control-plane
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
  extraPortMappings:
  - containerPort: 80
    hostPort: 80
    protocol: TCP
  - containerPort: 443
    hostPort: 443
    protocol: TCP
  - containerPort: 8081
    hostPort: 8081
    protocol: TCP
  - containerPort: 30090
    hostPort: 30090
    protocol: TCP
  - containerPort: 30300
    hostPort: 30300
    protocol: TCP
- role: worker
- role: worker
EOF

chown ubuntu:ubuntu /home/ubuntu/kind-config.yaml

# Create kind cluster as ubuntu user
sudo -u ubuntu bash << 'EOF'
export PATH=/usr/local/go/bin:/usr/local/bin:$PATH
cd /home/ubuntu

echo "Creating kind cluster..."
kind create cluster --config=kind-config.yaml --wait=300s

echo "Configuring kubectl..."
mkdir -p .kube
kind get kubeconfig --name ${cluster_name} > .kube/config
chmod 600 .kube/config

# Set up kubectl for current session
export KUBECONFIG=/home/ubuntu/.kube/config

echo "Installing ingress-nginx..."
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml

echo "Waiting for ingress-nginx to be ready..."
kubectl wait --namespace ingress-nginx \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/component=controller \
  --timeout=300s

# Configure ingress for external access
echo "Configuring ingress for external access..."
kubectl patch service ingress-nginx-controller -n ingress-nginx -p '{"spec":{"type":"NodePort"}}'

echo "Kind cluster setup complete"
EOF

# Configure ECR authentication
cat > /home/ubuntu/ecr-login.sh << 'EOFECR'
#!/bin/bash
aws ecr get-login-password --region ${aws_region} | docker login --username AWS --password-stdin ${ecr_repository}
EOFECR

chmod +x /home/ubuntu/ecr-login.sh
chown ubuntu:ubuntu /home/ubuntu/ecr-login.sh

# Create deployment script
cat > /home/ubuntu/deploy-quantum-safe-mesh.sh << 'EOFDEPLOY'
#!/bin/bash

set -e

export PATH=/usr/local/go/bin:/usr/local/bin:$PATH
export KUBECONFIG=/home/ubuntu/.kube/config

cd /home/ubuntu/quantum-safe-mesh

echo "🚀 Deploying Quantum-Safe Service Mesh..."

# Authenticate with ECR
/home/ubuntu/ecr-login.sh

# Build and tag images for ECR
echo "Building Docker images..."
if [ -f "scripts/deploy.sh" ]; then
    # Use existing deployment script
    ./scripts/deploy.sh build
    
    # Tag and push to ECR
    for service in auth gateway backend demo; do
        docker tag quantum-safe-mesh/$service:latest ${ecr_repository}/quantum-safe-mesh/$service:latest
        docker push ${ecr_repository}/quantum-safe-mesh/$service:latest
    done
    
    # Deploy to kind cluster
    ./scripts/deploy.sh helm
else
    echo "Deployment scripts not found, creating sample deployment..."
    kubectl create namespace quantum-safe-mesh || true
    
    # Create a simple test deployment
    cat > test-deployment.yaml << 'EOF'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
  namespace: quantum-safe-mesh
spec:
  replicas: 1
  selector:
    matchLabels:
      app: test-app
  template:
    metadata:
      labels:
        app: test-app
    spec:
      containers:
      - name: nginx
        image: nginx:alpine
        ports:
        - containerPort: 80
---
apiVersion: v1
kind: Service
metadata:
  name: test-service
  namespace: quantum-safe-mesh
spec:
  selector:
    app: test-app
  ports:
  - port: 80
    targetPort: 80
  type: NodePort
EOF
    
    kubectl apply -f test-deployment.yaml
fi

echo "✅ Deployment complete!"
echo "Access the service at: http://$(curl -s http://checkip.amazonaws.com):8081"

EOFDEPLOY

chmod +x /home/ubuntu/deploy-quantum-safe-mesh.sh
chown ubuntu:ubuntu /home/ubuntu/deploy-quantum-safe-mesh.sh

# Create monitoring setup script
cat > /home/ubuntu/setup-monitoring.sh << 'EOFMON'
#!/bin/bash

export PATH=/usr/local/go/bin:/usr/local/bin:$PATH
export KUBECONFIG=/home/ubuntu/.kube/config

echo "📊 Setting up monitoring..."

# Install Prometheus
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update

helm install prometheus prometheus-community/kube-prometheus-stack \
    --namespace monitoring \
    --create-namespace \
    --set prometheus.service.type=NodePort \
    --set prometheus.service.nodePort=30090 \
    --set grafana.service.type=NodePort \
    --set grafana.service.nodePort=30300 \
    --set grafana.adminPassword=admin123 \
    --wait

echo "✅ Monitoring setup complete!"
echo "Prometheus: http://$(curl -s http://checkip.amazonaws.com):30090"
echo "Grafana: http://$(curl -s http://checkip.amazonaws.com):30300 (admin/admin123)"

EOFMON

chmod +x /home/ubuntu/setup-monitoring.sh
chown ubuntu:ubuntu /home/ubuntu/setup-monitoring.sh

# Create startup script that runs on boot
cat > /etc/systemd/system/quantum-safe-mesh.service << 'EOFSVC'
[Unit]
Description=Quantum Safe Mesh Service
After=docker.service
Requires=docker.service

[Service]
Type=oneshot
User=ubuntu
WorkingDirectory=/home/ubuntu
ExecStart=/home/ubuntu/deploy-quantum-safe-mesh.sh
RemainAfterExit=yes
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOFSVC

# Enable the service (but don't start it immediately)
systemctl enable quantum-safe-mesh.service

# Create a welcome message
cat > /home/ubuntu/README.md << 'EOFREADME'
# Quantum-Safe Service Mesh on AWS EC2

This EC2 instance has been configured with:
- Docker
- kind (Kubernetes in Docker)
- kubectl
- Helm
- Go 1.21
- AWS CLI

## Available Commands

```bash
# Deploy the quantum-safe service mesh
./deploy-quantum-safe-mesh.sh

# Set up monitoring (Prometheus + Grafana)
./setup-monitoring.sh

# Check cluster status
kubectl get nodes
kubectl get pods --all-namespaces

# Access services
curl http://localhost:8081/health
```

## Service URLs (after deployment)

- Gateway Service: http://PUBLIC_IP:8081
- Prometheus: http://PUBLIC_IP:30090
- Grafana: http://PUBLIC_IP:30300 (admin/admin123)
- Kubernetes API: https://PUBLIC_IP:6443

## Manual Deployment

If you want to deploy manually:

```bash
cd quantum-safe-mesh
./scripts/deploy.sh all
```

EOFREADME

chown ubuntu:ubuntu /home/ubuntu/README.md

# Final system update
apt-get autoremove -y
apt-get autoclean

# Mark setup as complete
touch /var/log/user-data-complete

echo "🎉 Quantum-Safe Service Mesh EC2 setup completed at $(date)"
echo "Instance is ready for deployment!"