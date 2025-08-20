output "ec2_instance_info" {
  description = "Information about the EC2 instance"
  value = {
    instance_id = data.aws_instance.existing.id
    public_ip   = data.aws_instance.existing.public_ip
    private_ip  = data.aws_instance.existing.private_ip
    state       = data.aws_instance.existing.instance_state
  }
}

output "application_url" {
  description = "URL where the application is accessible"
  value       = "https://${var.domain_name}"
}

output "application_port" {
  description = "Port where the application is running"
  value       = var.app_port
}

output "dns_record" {
  description = "Cloudflare DNS record information"
  value = {
    name    = cloudflare_record.app_dns.name
    value   = cloudflare_record.app_dns.value
    proxied = cloudflare_record.app_dns.proxied
  }
}

output "deployment_status" {
  description = "Deployment completion status"
  value       = "Application deployed successfully to ${var.ec2_public_ip}:${var.app_port}"
  depends_on  = [null_resource.create_service]
}