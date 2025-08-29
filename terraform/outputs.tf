# Outputs for Quantum-Safe Service Mesh AWS Infrastructure

output "instance_public_ip" {
  description = "Public IP address of the EC2 instance"
  value       = aws_eip.quantum_safe_mesh.public_ip
}

output "instance_public_dns" {
  description = "Public DNS name of the EC2 instance"
  value       = aws_instance.quantum_safe_mesh.public_dns
}

output "instance_id" {
  description = "ID of the EC2 instance"
  value       = aws_instance.quantum_safe_mesh.id
}

output "vpc_id" {
  description = "ID of the VPC"
  value       = module.vpc.vpc_id
}

output "security_group_id" {
  description = "ID of the security group"
  value       = aws_security_group.quantum_safe_mesh.id
}

output "ecr_repositories" {
  description = "ECR repository URLs"
  value = {
    for k, v in aws_ecr_repository.services : k => v.repository_url
  }
}

output "private_key_path" {
  description = "Path to the private key file"
  value       = local_file.private_key.filename
  sensitive   = true
}

output "ssh_command" {
  description = "SSH command to connect to the instance"
  value       = "ssh -i ${local_file.private_key.filename} ubuntu@${aws_eip.quantum_safe_mesh.public_ip}"
}

output "kubectl_command" {
  description = "Command to configure kubectl for the kind cluster"
  value       = "ssh -i ${local_file.private_key.filename} ubuntu@${aws_eip.quantum_safe_mesh.public_ip} 'kind get kubeconfig --name ${var.project_name}-${var.environment}' > ~/.kube/config-${var.environment}"
}

output "gateway_service_url" {
  description = "URL to access the gateway service"
  value       = "http://${aws_eip.quantum_safe_mesh.public_ip}:8081"
}

output "kubernetes_api_url" {
  description = "URL to access the Kubernetes API server"
  value       = "https://${aws_eip.quantum_safe_mesh.public_ip}:6443"
}

output "deployment_commands" {
  description = "Commands to deploy the quantum-safe service mesh"
  value = {
    ssh_connect = "ssh -i ${local_file.private_key.filename} ubuntu@${aws_eip.quantum_safe_mesh.public_ip}"
    deploy_services = "ssh -i ${local_file.private_key.filename} ubuntu@${aws_eip.quantum_safe_mesh.public_ip} 'cd /home/ubuntu/quantum-safe-mesh && ./scripts/deploy.sh all'"
    check_status = "ssh -i ${local_file.private_key.filename} ubuntu@${aws_eip.quantum_safe_mesh.public_ip} 'cd /home/ubuntu/quantum-safe-mesh && ./scripts/deploy.sh status'"
    view_logs = "ssh -i ${local_file.private_key.filename} ubuntu@${aws_eip.quantum_safe_mesh.public_ip} 'kubectl logs job/quantum-safe-demo -n quantum-safe-mesh'"
  }
}

output "monitoring_urls" {
  description = "URLs for monitoring and observability"
  value = {
    prometheus = "http://${aws_eip.quantum_safe_mesh.public_ip}:30090"
    grafana = "http://${aws_eip.quantum_safe_mesh.public_ip}:30300"
    cloudwatch_dashboard = "https://${var.aws_region}.console.aws.amazon.com/cloudwatch/home?region=${var.aws_region}#dashboards:name=${aws_cloudwatch_dashboard.quantum_safe_mesh.dashboard_name}"
  }
}

# Output environment information
output "environment_info" {
  description = "Environment information"
  value = {
    environment = var.environment
    region = var.aws_region
    instance_type = var.instance_type
    kubernetes_version = var.kubernetes_version
    kind_version = var.kind_version
  }
}