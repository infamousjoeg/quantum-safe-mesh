# Variables for Quantum-Safe Service Mesh AWS Infrastructure

variable "aws_region" {
  description = "AWS region for deployment"
  type        = string
  default     = "us-west-2"
}

variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
  default     = "dev"
}

variable "project_name" {
  description = "Name of the project"
  type        = string
  default     = "quantum-safe-mesh"
}

variable "owner" {
  description = "Owner of the resources"
  type        = string
  default     = "quantum-safe-team"
}

variable "vpc_cidr" {
  description = "CIDR block for the VPC"
  type        = string
  default     = "10.0.0.0/16"
}

variable "allowed_cidr_blocks" {
  description = "CIDR blocks allowed to access the instance"
  type        = list(string)
  default     = ["0.0.0.0/0"] # Restrict this in production
}

variable "instance_type" {
  description = "EC2 instance type"
  type        = string
  default     = "t3.large"
  
  validation {
    condition = can(regex("^t3\\.(medium|large|xlarge)|^m5\\.(large|xlarge|2xlarge)|^c5\\.(large|xlarge|2xlarge)$", var.instance_type))
    error_message = "Instance type must be suitable for running Docker and Kubernetes (minimum t3.medium)."
  }
}

variable "root_volume_size" {
  description = "Size of the root EBS volume in GB"
  type        = number
  default     = 50
  
  validation {
    condition     = var.root_volume_size >= 30
    error_message = "Root volume size must be at least 30 GB."
  }
}

variable "services" {
  description = "List of services to create ECR repositories for"
  type        = list(string)
  default     = ["auth", "gateway", "backend", "demo"]
}

variable "enable_monitoring" {
  description = "Enable CloudWatch monitoring and logging"
  type        = bool
  default     = true
}

variable "enable_backup" {
  description = "Enable automated backups"
  type        = bool
  default     = false
}

variable "kind_version" {
  description = "Version of kind to install"
  type        = string
  default     = "v0.20.0"
}

variable "kubernetes_version" {
  description = "Version of Kubernetes to run in kind"
  type        = string
  default     = "v1.28.0"
}

variable "kubectl_version" {
  description = "Version of kubectl to install"
  type        = string
  default     = "v1.28.0"
}

variable "helm_version" {
  description = "Version of Helm to install"
  type        = string
  default     = "v3.13.0"
}

variable "ingress_nginx_version" {
  description = "Version of ingress-nginx to install"
  type        = string
  default     = "v1.8.0"
}

variable "alert_email" {
  description = "Email address for CloudWatch alerts"
  type        = string
  default     = ""
}

# Local variables
locals {
  common_tags = {
    Project     = var.project_name
    Environment = var.environment
    ManagedBy   = "terraform"
    Owner       = var.owner
  }
}