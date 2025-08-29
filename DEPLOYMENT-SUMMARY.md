# Deployment Options Summary

This project provides multiple deployment options for the Quantum-Safe Service Mesh, each suited for different use cases and environments.

## 🏠 Local Development

**Best for**: Learning, development, testing

**Requirements**: Go, Docker, Make

**Commands**:
```bash
make generate-keys
make run-auth &
make run-gateway &
make run-backend &
make demo
```

**Access**: 
- Gateway: http://localhost:8081
- Direct services: 8080, 8081, 8082

## ☸️ Kubernetes (Local)

**Best for**: Container testing, Kubernetes learning

**Requirements**: Docker, kubectl, kind/minikube, Helm

**Commands**:
```bash
./scripts/deploy.sh all
# OR
make k8s-all
```

**Access**:
- Port-forward: `kubectl port-forward svc/gateway-service 8081:8081`
- Gateway: http://localhost:8081

## 🌩️ AWS Cloud Deployment

**Best for**: Production, scalable demos, cloud-native testing

**Requirements**: Terraform, AWS CLI, AWS Account

**Commands**:
```bash
./scripts/aws-deploy.sh all
# OR  
make aws-all
```

**Access**:
- Gateway: http://INSTANCE_IP:8081
- SSH: `ssh -i key.pem ubuntu@INSTANCE_IP`
- Monitoring: CloudWatch Dashboard

## 📊 Comparison Matrix

| Feature | Local | Kubernetes | AWS |
|---------|-------|------------|-----|
| **Setup Time** | 5 min | 10 min | 15 min |
| **Resource Usage** | Low | Medium | Variable |
| **Cost** | Free | Free | ~$125/month |
| **Scalability** | None | Manual | Auto-scaling ready |
| **Monitoring** | Basic logs | Prometheus/Grafana | CloudWatch + P8s/Grafana |
| **Security** | Development | Network policies | VPC + Security groups |
| **Persistence** | File system | ConfigMaps/Secrets | EBS volumes |
| **Load Balancing** | None | Kubernetes Services | AWS ALB |
| **SSL/TLS** | None | Cert-manager | ACM integration |
| **Backup** | Manual | Velero | EBS snapshots |
| **High Availability** | No | Multi-replica | Multi-AZ ready |

## 🎯 Use Case Recommendations

### Development & Learning
```bash
# Quick start for learning PQC
make generate-keys
make run-auth & make run-gateway & make run-backend &
make demo
```

### Container & K8s Testing  
```bash
# Test containerized deployment
./scripts/deploy.sh all
kubectl get pods -n quantum-safe-mesh
```

### Production Demo
```bash
# Full cloud deployment with monitoring
export AWS_REGION="us-west-2"
export ALERT_EMAIL="alerts@company.com"
./scripts/aws-deploy.sh all
```

### CI/CD Integration
```bash
# Automated testing in CI
./scripts/deploy.sh build
./scripts/deploy.sh test
./scripts/deploy.sh cleanup
```

## 🔄 Migration Path

### Local → Kubernetes
1. Test locally: `make demo`
2. Build containers: `make k8s-build`
3. Deploy to K8s: `make k8s-helm`
4. Verify: `make k8s-status`

### Kubernetes → AWS
1. Test K8s deployment: `./scripts/deploy.sh all`
2. Configure AWS: `cp terraform/terraform.tfvars.example terraform/terraform.tfvars`
3. Deploy to AWS: `./scripts/aws-deploy.sh all`
4. Verify: `./scripts/aws-deploy.sh test`

### Development → Production
1. Update configuration files
2. Restrict security groups
3. Enable monitoring & alerting
4. Set up backup procedures
5. Configure SSL certificates
6. Implement CI/CD pipeline

## 🛠️ Troubleshooting by Environment

### Local Issues
- **Port conflicts**: `netstat -tulpn | grep :8081`
- **Go build errors**: `go mod tidy`
- **Key generation**: Check entropy with `cat /proc/sys/kernel/random/entropy_avail`

### Kubernetes Issues  
- **Pod failures**: `kubectl describe pod POD_NAME -n quantum-safe-mesh`
- **Image pulls**: `docker images | grep quantum-safe-mesh`
- **Network**: `kubectl get networkpolicy -n quantum-safe-mesh`

### AWS Issues
- **Terraform errors**: `terraform plan` and check AWS credentials
- **EC2 access**: Security group rules and SSH key permissions
- **Service connectivity**: Check instance logs and kind cluster status

## 📈 Scaling Considerations

### Local Development
- Limited to single machine resources
- No redundancy or load balancing
- Suitable for up to moderate load testing

### Kubernetes
- Scale pods: `kubectl scale deployment gateway-service --replicas=5`
- Add nodes: depends on cluster setup
- Resource limits: Set in Helm values

### AWS Cloud
- Scale instance: Change `instance_type` in Terraform
- Multi-instance: Deploy across multiple AZs
- Auto-scaling: Implement ASG with Terraform
- Load balancing: Add ALB for production

## 🔒 Security Progression

### Local (Development)
- Basic PQC implementation
- No network security
- File-based key storage

### Kubernetes (Testing)
- Network policies for zero-trust
- RBAC and service accounts  
- Kubernetes secrets management
- Pod security contexts

### AWS (Production)
- VPC isolation
- Security groups
- IAM roles and policies
- Encrypted EBS volumes
- CloudWatch monitoring
- SNS alerting

---

**Choose the deployment that matches your needs and scale up as requirements grow!**