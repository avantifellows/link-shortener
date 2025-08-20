# AWS Configuration
variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "ec2_instance_id" {
  description = "Existing EC2 instance ID"
  type        = string
}

variable "ec2_public_ip" {
  description = "Public IP of the existing EC2 instance"
  type        = string
}

variable "ssh_private_key_path" {
  description = "Path to SSH private key for EC2 access"
  type        = string
  default     = "~/.ssh/id_rsa"
}

variable "ssh_user" {
  description = "SSH user for EC2 instance"
  type        = string
  default     = "ubuntu"
}

# Application Configuration
variable "app_port" {
  description = "Port where the application will run"
  type        = number
  default     = 8080
}

variable "auth_token" {
  description = "Bearer token for API authentication"
  type        = string
  sensitive   = true
}

# Domain Configuration
variable "domain_name" {
  description = "Domain name for the application"
  type        = string
  default     = "lnk.avantifellows.org"
}

variable "cloudflare_zone_id" {
  description = "Cloudflare Zone ID for the domain"
  type        = string
}

variable "cloudflare_api_token" {
  description = "Cloudflare API token"
  type        = string
  sensitive   = true
}

# SSL Certificates (Cloudflare Origin Certificates)
variable "cloudflare_origin_cert" {
  description = "Cloudflare Origin Certificate (PEM format)"
  type        = string
  sensitive   = true
}

variable "cloudflare_origin_key" {
  description = "Cloudflare Origin Private Key (PEM format)"
  type        = string
  sensitive   = true
}