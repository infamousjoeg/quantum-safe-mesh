# AWS Deployment Guide

This guide covers deploying the Quantum-Safe Service Mesh on AWS using Terraform and a kind cluster running on EC2.

## 🌟 Architecture Overview

```
Internet → [Application Load Balancer] → EC2 Instance (kind cluster)
           ↓
[Security Groups] → [VPC] → [Private/Public Subnets]
           ↓
Kind Cluster → [Ingress Controller] → [PQC Services]
           ↓
[ECR] ← [Docker Images] ← [Build Process]
           ↓
[CloudWatch] ← [Monitoring & Logs] ← [Services]
```

### Components

- **EC2 Instance**: t3.large instance running Ubuntu 22.04
- **kind Cluster**: Multi-node Kubernetes cluster (1 control-plane, 2 workers)
- **ECR Repositories**: Container registry for service images
- **VPC**: Isolated network with public/private subnets
- **Security Groups**: Firewall rules for secure access
- **CloudWatch**: Monitoring, logging, and alerting
- **Elastic IP**: Static IP address for external access

## 🚀 Quick Start

### 1-Command Deployment
```bash
# Complete AWS deployment
./scripts/aws-deploy.sh all
```

## 📋 Prerequisites

### Required Tools
- **Terraform** (>= 1.0)
- **AWS CLI** (>= 2.0)
- **kubectl** (optional, for local management)
- **jq** (optional, for JSON processing)

### AWS Requirements
- AWS Account with programmatic access
- AWS CLI configured with appropriate credentials
- Permissions for EC2, VPC, ECR, CloudWatch, IAM

### AWS Permissions Required
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ec2:*",
        "vpc:*",
        "ecr:*",
        "iam:*",
        "cloudwatch:*",
        "logs:*",
        "sns:*",
        "ssm:*"
      ],
      "Resource": "*"
    }
  ]
}
```

## 🔧 Configuration

### Environment Variables
```bash
export AWS_REGION="us-west-2"        # AWS region
export ENVIRONMENT="dev"             # Environment name
export PROJECT_NAME="quantum-safe-mesh"  # Project name
export ALERT_EMAIL="you@example.com"     # CloudWatch alerts
```

### Terraform Variables

Create `terraform/terraform.tfvars`:
```hcl
aws_region = "us-west-2"
environment = "dev"
project_name = "quantum-safe-mesh"
owner = "your-team"

# Instance configuration
instance_type = "t3.large"
root_volume_size = 50
allowed_cidr_blocks = ["0.0.0.0/0"]  # Restrict in production

# Monitoring
alert_email = "alerts@example.com"
enable_monitoring = true

# Versions
kind_version = "v0.20.0"
kubernetes_version = "v1.28.0"
kubectl_version = "v1.28.0"
helm_version = "v3.13.0"
```

## 🛠️ Deployment Steps

### Option 1: Automated Deployment
```bash
# Clone repository
git clone <your-repo-url>
cd quantum-safe-mesh

# Make script executable
chmod +x scripts/aws-deploy.sh

# Complete deployment
./scripts/aws-deploy.sh all
```

### Option 2: Step-by-Step Deployment
```bash
# 1. Initialize Terraform
./scripts/aws-deploy.sh init

# 2. Create execution plan
./scripts/aws-deploy.sh plan

# 3. Deploy infrastructure
./scripts/aws-deploy.sh apply

# 4. Deploy services
./scripts/aws-deploy.sh deploy

# 5. Setup monitoring
./scripts/aws-deploy.sh monitor

# 6. Run tests
./scripts/aws-deploy.sh test

# 7. Show access info
./scripts/aws-deploy.sh info
```

### Option 3: Manual Terraform
```bash
cd terraform

# Initialize
terraform init

# Plan
terraform plan \
  -var="aws_region=us-west-2" \
  -var="environment=dev" \
  -out=tfplan

# Apply
terraform apply tfplan

# Get outputs
terraform output
```

## 📊 Monitoring & Observability

### CloudWatch Integration
- **Dashboard**: AWS CloudWatch dashboard with EC2 metrics
- **Alarms**: CPU utilization and status check alarms
- **Logs**: Application and system logs
- **SNS Alerts**: Email notifications for critical issues

### Application Monitoring
- **Prometheus**: http://INSTANCE_IP:30090
- **Grafana**: http://INSTANCE_IP:30300 (admin/admin123)
- **PQC Metrics**: Custom metrics for quantum-safe operations

### Log Sources
- EC2 instance logs
- kind cluster logs
- Application service logs
- Kubernetes system logs

## 🌐 Access & URLs

After deployment, you'll have access to:

```bash
# Service URLs
Gateway Service: http://INSTANCE_IP:8081
Kubernetes API: https://INSTANCE_IP:6443

# Monitoring
Prometheus: http://INSTANCE_IP:30090
Grafana: http://INSTANCE_IP:30300
CloudWatch: AWS Console → CloudWatch

# SSH Access
ssh -i quantum-safe-mesh-key.pem ubuntu@INSTANCE_IP
```

## 🔍 Testing & Validation

### Health Checks
```bash
# Check instance status
curl http://INSTANCE_IP:8081/health

# Check all services
./scripts/aws-deploy.sh test
```

### PQC Tests
```bash
# Echo test
curl -X POST http://INSTANCE_IP:8081/echo \
  -H "Content-Type: application/json" \
  -d '{"message": "AWS PQC test"}'

# Processing test
curl -X POST http://INSTANCE_IP:8081/process \
  -H "Content-Type: application/json" \
  -d '{"data": "quantum processing on AWS"}'

# Status check
curl http://INSTANCE_IP:8081/status
```

### Kubernetes Operations
```bash
# SSH into instance
ssh -i quantum-safe-mesh-key.pem ubuntu@INSTANCE_IP

# Check cluster
kubectl get nodes
kubectl get pods -A

# Check services
kubectl get svc -n quantum-safe-mesh
kubectl get ing -n quantum-safe-mesh

# View logs
kubectl logs deployment/gateway-service -n quantum-safe-mesh
```

## 🔒 Security Considerations

### Network Security
- **VPC**: Isolated network environment
- **Security Groups**: Restrictive firewall rules
- **Private Subnets**: Database and internal services
- **Public Subnets**: Load balancers and bastion hosts only

### Instance Security
- **Encrypted EBS**: Root volume encryption enabled
- **IAM Roles**: Minimal required permissions
- **SSH Keys**: Generated and managed by Terraform
- **Security Updates**: Automated via user-data script

### Application Security
- **Network Policies**: Zero-trust networking in Kubernetes
- **Pod Security**: Non-root containers with dropped capabilities
- **Secrets**: Kubernetes secrets for sensitive data
- **TLS**: HTTPS/TLS for all external communication

## 💰 Cost Optimization

### Estimated Monthly Costs (us-west-2)
- **EC2 t3.large**: ~$60/month
- **EBS gp3 50GB**: ~$5/month
- **Elastic IP**: ~$4/month
- **NAT Gateway**: ~$45/month
- **CloudWatch**: ~$10/month
- **ECR**: ~$1/month
- **Total**: ~$125/month

### Cost Reduction Options
```hcl
# Use smaller instance
instance_type = "t3.medium"  # ~$30/month

# Single AZ deployment
availability_zones = ["us-west-2a"]  # Reduces NAT costs

# Disable NAT Gateway for private subnets
# (requires instance in public subnet)
```

## 🔧 Customization

### Instance Types
```hcl
# Development
instance_type = "t3.medium"   # 2 vCPU, 4 GB RAM

# Production
instance_type = "t3.large"    # 2 vCPU, 8 GB RAM
instance_type = "t3.xlarge"   # 4 vCPU, 16 GB RAM
```

### Multi-Environment Setup
```bash
# Development
ENVIRONMENT=dev ./scripts/aws-deploy.sh all

# Staging
ENVIRONMENT=staging ./scripts/aws-deploy.sh all

# Production
ENVIRONMENT=prod ./scripts/aws-deploy.sh all
```

### Custom Domain
```hcl
# In terraform/terraform.tfvars
domain_name = "quantum-mesh.example.com"
certificate_arn = "arn:aws:acm:us-west-2:123456789012:certificate/..."
```

## 🚨 Troubleshooting

### Common Issues

#### Terraform Errors
```bash
# State lock issues
terraform force-unlock LOCK_ID

# Provider version conflicts
terraform init -upgrade

# Permission errors
aws sts get-caller-identity
```

#### Instance Connection Issues
```bash
# Security group rules
aws ec2 describe-security-groups --group-ids sg-xxx

# Instance status
aws ec2 describe-instances --instance-ids i-xxx

# System logs
aws ec2 get-console-output --instance-id i-xxx
```

#### Service Deployment Issues
```bash
# SSH into instance
ssh -i quantum-safe-mesh-key.pem ubuntu@INSTANCE_IP

# Check user-data logs
sudo tail -f /var/log/user-data.log

# Check kind cluster
kind get clusters
kubectl cluster-info

# Check service logs
kubectl logs -f deployment/auth-service -n quantum-safe-mesh
```

### Debug Commands
```bash
# Infrastructure status
./scripts/aws-deploy.sh info

# Service health
curl -f http://INSTANCE_IP:8081/health

# Cluster status
ssh -i key.pem ubuntu@INSTANCE_IP 'kubectl get pods -A'

# Application logs
ssh -i key.pem ubuntu@INSTANCE_IP 'kubectl logs job/quantum-safe-demo -n quantum-safe-mesh'
```

## 🗑️ Cleanup

### Complete Removal
```bash
# Destroy all AWS resources
./scripts/aws-deploy.sh destroy

# Or manually
cd terraform
terraform destroy -auto-approve
```

### Selective Cleanup
```bash
# Remove specific resources
terraform destroy -target=aws_instance.quantum_safe_mesh
```

## 📈 Scaling & Production

### Horizontal Scaling
- Deploy multiple instances across AZs
- Use Application Load Balancer
- Implement Auto Scaling Groups
- Use RDS for shared state

### Production Checklist
- [ ] Use private ECR repositories
- [ ] Implement backup strategies
- [ ] Set up log aggregation
- [ ] Configure SSL/TLS certificates
- [ ] Implement proper secrets management
- [ ] Set up disaster recovery
- [ ] Configure monitoring alerts
- [ ] Implement security scanning
- [ ] Set up CI/CD pipelines
- [ ] Document runbooks

## 🔗 Integration

### CI/CD Pipeline
```yaml
# GitHub Actions example
name: Deploy to AWS
on:
  push:
    branches: [main]
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - name: Deploy to AWS
      run: ./scripts/aws-deploy.sh all
      env:
        AWS_ACCESS_KEY_ID: ${{ secrets.AWS_ACCESS_KEY_ID }}
        AWS_SECRET_ACCESS_KEY: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
```

### Monitoring Integration
- **Datadog**: Custom metrics and dashboards
- **New Relic**: Application performance monitoring
- **PagerDuty**: Alert routing and escalation
- **Slack**: Notification integration

## 📚 Additional Resources

- [AWS EC2 Documentation](https://docs.aws.amazon.com/ec2/)
- [Terraform AWS Provider](https://registry.terraform.io/providers/hashicorp/aws/latest/docs)
- [kind Documentation](https://kind.sigs.k8s.io/)
- [AWS CloudWatch](https://docs.aws.amazon.com/cloudwatch/)

---

**🌟 This deployment showcases enterprise-ready quantum-safe cryptography in the cloud. Scale, monitor, and secure your post-quantum future on AWS!**