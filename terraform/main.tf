# Quantum-Safe Service Mesh AWS EC2 Infrastructure with Kind
# This Terraform configuration deploys an EC2 instance with kind cluster for the quantum-safe service mesh

terraform {
  required_version = ">= 1.0"
  
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "~> 4.0"
    }
  }

  # Uncomment and configure for remote state
  # backend "s3" {
  #   bucket = "quantum-safe-mesh-terraform-state"
  #   key    = "infrastructure/terraform.tfstate"
  #   region = "us-west-2"
  #   encrypt = true
  #   dynamodb_table = "quantum-safe-mesh-terraform-locks"
  # }
}

# Configure AWS Provider
provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Project     = "quantum-safe-mesh"
      Environment = var.environment
      ManagedBy   = "terraform"
      Owner       = var.owner
    }
  }
}

# Data sources
data "aws_availability_zones" "available" {
  state = "available"
}

data "aws_ami" "ubuntu" {
  most_recent = true
  owners      = ["099720109477"] # Canonical

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-22.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}

# VPC Module
module "vpc" {
  source = "./modules/vpc"

  environment         = var.environment
  vpc_cidr           = var.vpc_cidr
  availability_zones = slice(data.aws_availability_zones.available.names, 0, 2)
  
  tags = local.common_tags
}

# Security Group for EC2 instance
resource "aws_security_group" "quantum_safe_mesh" {
  name_prefix = "${var.project_name}-${var.environment}-"
  vpc_id      = module.vpc.vpc_id
  description = "Security group for Quantum Safe Mesh EC2 instance"

  # SSH access
  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = var.allowed_cidr_blocks
    description = "SSH access"
  }

  # HTTP access for the gateway service
  ingress {
    from_port   = 8081
    to_port     = 8081
    protocol    = "tcp"
    cidr_blocks = var.allowed_cidr_blocks
    description = "Gateway service HTTP"
  }

  # HTTPS access
  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = var.allowed_cidr_blocks
    description = "HTTPS access"
  }

  # HTTP access for ingress
  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = var.allowed_cidr_blocks
    description = "HTTP access"
  }

  # Kind API server access (for kubectl)
  ingress {
    from_port   = 6443
    to_port     = 6443
    protocol    = "tcp"
    cidr_blocks = var.allowed_cidr_blocks
    description = "Kubernetes API server"
  }

  # NodePort range for services
  ingress {
    from_port   = 30000
    to_port     = 32767
    protocol    = "tcp"
    cidr_blocks = var.allowed_cidr_blocks
    description = "Kubernetes NodePort range"
  }

  # All outbound traffic
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
    description = "All outbound traffic"
  }

  tags = merge(local.common_tags, {
    Name = "${var.project_name}-${var.environment}-sg"
  })
}

# Key pair for EC2 access
resource "tls_private_key" "quantum_safe_mesh" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "aws_key_pair" "quantum_safe_mesh" {
  key_name   = "${var.project_name}-${var.environment}-key"
  public_key = tls_private_key.quantum_safe_mesh.public_key_openssh

  tags = local.common_tags
}

# Save private key locally
resource "local_file" "private_key" {
  content  = tls_private_key.quantum_safe_mesh.private_key_pem
  filename = "${path.module}/quantum-safe-mesh-key.pem"
  file_permission = "0600"
}

# IAM role for EC2 instance
resource "aws_iam_role" "quantum_safe_mesh_ec2" {
  name = "${var.project_name}-${var.environment}-ec2-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      }
    ]
  })

  tags = local.common_tags
}

# IAM policy for ECR access
resource "aws_iam_role_policy" "quantum_safe_mesh_ecr" {
  name = "ECRAccess"
  role = aws_iam_role.quantum_safe_mesh_ec2.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "ecr:GetAuthorizationToken",
          "ecr:BatchCheckLayerAvailability",
          "ecr:GetDownloadUrlForLayer",
          "ecr:BatchGetImage"
        ]
        Resource = "*"
      }
    ]
  })
}

# IAM instance profile
resource "aws_iam_instance_profile" "quantum_safe_mesh" {
  name = "${var.project_name}-${var.environment}-profile"
  role = aws_iam_role.quantum_safe_mesh_ec2.name

  tags = local.common_tags
}

# ECR repositories
resource "aws_ecr_repository" "services" {
  for_each = toset(var.services)
  
  name                 = "${var.project_name}/${each.key}"
  image_tag_mutability = "MUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }

  encryption_configuration {
    encryption_type = "AES256"
  }

  tags = merge(local.common_tags, {
    Service = each.key
  })
}

# ECR lifecycle policy
resource "aws_ecr_lifecycle_policy" "services" {
  for_each = aws_ecr_repository.services
  
  repository = each.value.name

  policy = jsonencode({
    rules = [
      {
        rulePriority = 1
        description  = "Keep last 10 images"
        selection = {
          tagStatus   = "any"
          countType   = "imageCountMoreThan"
          countNumber = 10
        }
        action = {
          type = "expire"
        }
      }
    ]
  })
}

# User data script to set up the EC2 instance
locals {
  user_data = templatefile("${path.module}/user-data.sh", {
    aws_region     = var.aws_region
    cluster_name   = "${var.project_name}-${var.environment}"
    ecr_repository = "${data.aws_caller_identity.current.account_id}.dkr.ecr.${var.aws_region}.amazonaws.com"
    services       = jsonencode(var.services)
  })
}

data "aws_caller_identity" "current" {}

# EC2 instance
resource "aws_instance" "quantum_safe_mesh" {
  ami                    = data.aws_ami.ubuntu.id
  instance_type         = var.instance_type
  key_name              = aws_key_pair.quantum_safe_mesh.key_name
  vpc_security_group_ids = [aws_security_group.quantum_safe_mesh.id]
  subnet_id             = module.vpc.public_subnet_ids[0]
  iam_instance_profile  = aws_iam_instance_profile.quantum_safe_mesh.name

  user_data = base64encode(local.user_data)

  root_block_device {
    volume_type = "gp3"
    volume_size = var.root_volume_size
    encrypted   = true
  }

  tags = merge(local.common_tags, {
    Name = "${var.project_name}-${var.environment}-instance"
  })

  # Wait for instance to be ready
  provisioner "remote-exec" {
    inline = [
      "while [ ! -f /var/log/user-data-complete ]; do sleep 10; done",
      "echo 'User data setup complete'"
    ]

    connection {
      type        = "ssh"
      host        = self.public_ip
      user        = "ubuntu"
      private_key = tls_private_key.quantum_safe_mesh.private_key_pem
      timeout     = "10m"
    }
  }
}

# Elastic IP for the instance
resource "aws_eip" "quantum_safe_mesh" {
  instance = aws_instance.quantum_safe_mesh.id
  domain   = "vpc"

  tags = merge(local.common_tags, {
    Name = "${var.project_name}-${var.environment}-eip"
  })

  depends_on = [module.vpc.internet_gateway]
}