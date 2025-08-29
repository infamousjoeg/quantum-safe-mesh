# CloudWatch Monitoring for Quantum-Safe Service Mesh

# CloudWatch Log Group for EC2 instance
resource "aws_cloudwatch_log_group" "quantum_safe_mesh" {
  name              = "/aws/ec2/${var.project_name}-${var.environment}"
  retention_in_days = 7

  tags = local.common_tags
}

# CloudWatch Dashboard
resource "aws_cloudwatch_dashboard" "quantum_safe_mesh" {
  dashboard_name = "${var.project_name}-${var.environment}-dashboard"

  dashboard_body = jsonencode({
    widgets = [
      {
        type   = "metric"
        x      = 0
        y      = 0
        width  = 12
        height = 6

        properties = {
          metrics = [
            ["AWS/EC2", "CPUUtilization", "InstanceId", aws_instance.quantum_safe_mesh.id],
            [".", "NetworkIn", ".", "."],
            [".", "NetworkOut", ".", "."]
          ]
          view    = "timeSeries"
          stacked = false
          region  = var.aws_region
          title   = "EC2 Instance Metrics"
          period  = 300
        }
      },
      {
        type   = "metric"
        x      = 12
        y      = 0
        width  = 12
        height = 6

        properties = {
          metrics = [
            ["AWS/EC2", "StatusCheckFailed", "InstanceId", aws_instance.quantum_safe_mesh.id],
            [".", "StatusCheckFailed_Instance", ".", "."],
            [".", "StatusCheckFailed_System", ".", "."]
          ]
          view   = "timeSeries"
          region = var.aws_region
          title  = "Instance Status Checks"
          period = 300
        }
      },
      {
        type   = "log"
        x      = 0
        y      = 6
        width  = 24
        height = 6

        properties = {
          query   = "SOURCE '${aws_cloudwatch_log_group.quantum_safe_mesh.name}' | fields @timestamp, @message | filter @message like /quantum-safe/ | sort @timestamp desc | limit 100"
          region  = var.aws_region
          title   = "Quantum-Safe Service Mesh Logs"
        }
      }
    ]
  })
}

# CloudWatch Alarms
resource "aws_cloudwatch_metric_alarm" "high_cpu" {
  alarm_name          = "${var.project_name}-${var.environment}-high-cpu"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "CPUUtilization"
  namespace           = "AWS/EC2"
  period              = "300"
  statistic           = "Average"
  threshold           = "80"
  alarm_description   = "This metric monitors ec2 cpu utilization"
  alarm_actions       = [aws_sns_topic.alerts.arn]

  dimensions = {
    InstanceId = aws_instance.quantum_safe_mesh.id
  }

  tags = local.common_tags
}

resource "aws_cloudwatch_metric_alarm" "instance_status_check" {
  alarm_name          = "${var.project_name}-${var.environment}-instance-status"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "StatusCheckFailed_Instance"
  namespace           = "AWS/EC2"
  period              = "300"
  statistic           = "Maximum"
  threshold           = "0"
  alarm_description   = "This metric monitors instance status check"
  alarm_actions       = [aws_sns_topic.alerts.arn]

  dimensions = {
    InstanceId = aws_instance.quantum_safe_mesh.id
  }

  tags = local.common_tags
}

# SNS Topic for alerts
resource "aws_sns_topic" "alerts" {
  name = "${var.project_name}-${var.environment}-alerts"

  tags = local.common_tags
}

# SNS Topic Subscription (email)
resource "aws_sns_topic_subscription" "email_alerts" {
  count = var.alert_email != "" ? 1 : 0
  
  topic_arn = aws_sns_topic.alerts.arn
  protocol  = "email"
  endpoint  = var.alert_email
}

# CloudWatch Agent IAM Role
resource "aws_iam_role" "cloudwatch_agent" {
  name = "${var.project_name}-${var.environment}-cloudwatch-agent"

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

# Attach CloudWatch Agent policy
resource "aws_iam_role_policy_attachment" "cloudwatch_agent" {
  role       = aws_iam_role.cloudwatch_agent.name
  policy_arn = "arn:aws:iam::aws:policy/CloudWatchAgentServerPolicy"
}

# Custom metric policy for application metrics
resource "aws_iam_role_policy" "custom_metrics" {
  name = "CustomMetricsPolicy"
  role = aws_iam_role.quantum_safe_mesh_ec2.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "cloudwatch:PutMetricData",
          "logs:PutLogEvents",
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:DescribeLogStreams"
        ]
        Resource = "*"
      }
    ]
  })
}

# CloudWatch Agent configuration
resource "aws_ssm_parameter" "cloudwatch_agent_config" {
  name  = "/${var.project_name}/${var.environment}/cloudwatch-agent-config"
  type  = "String"
  value = jsonencode({
    agent = {
      metrics_collection_interval = 60
      run_as_user = "ubuntu"
    }
    logs = {
      logs_collected = {
        files = {
          collect_list = [
            {
              file_path = "/var/log/user-data.log"
              log_group_name = aws_cloudwatch_log_group.quantum_safe_mesh.name
              log_stream_name = "{instance_id}/user-data"
              timezone = "UTC"
            },
            {
              file_path = "/home/ubuntu/quantum-safe-mesh/logs/*.log"
              log_group_name = aws_cloudwatch_log_group.quantum_safe_mesh.name
              log_stream_name = "{instance_id}/application"
              timezone = "UTC"
            }
          ]
        }
      }
    }
    metrics = {
      namespace = "${var.project_name}/${var.environment}"
      metrics_collected = {
        cpu = {
          measurement = [
            "cpu_usage_idle",
            "cpu_usage_iowait",
            "cpu_usage_user",
            "cpu_usage_system"
          ]
          metrics_collection_interval = 60
        }
        disk = {
          measurement = [
            "used_percent"
          ]
          metrics_collection_interval = 60
          resources = ["*"]
        }
        diskio = {
          measurement = [
            "io_time",
            "read_bytes",
            "write_bytes",
            "reads",
            "writes"
          ]
          metrics_collection_interval = 60
          resources = ["*"]
        }
        mem = {
          measurement = [
            "mem_used_percent"
          ]
          metrics_collection_interval = 60
        }
        netstat = {
          measurement = [
            "tcp_established",
            "tcp_time_wait"
          ]
          metrics_collection_interval = 60
        }
        swap = {
          measurement = [
            "swap_used_percent"
          ]
          metrics_collection_interval = 60
        }
      }
    }
  })

  tags = local.common_tags
}