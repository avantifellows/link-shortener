# Minimal deployment - just app on specific port
# No Nginx, no SSL, just the application running

# Same as main.tf but remove:
# - null_resource.install_ssl
# - null_resource.configure_nginx

# Just deploy app and configure security group for the port
resource "aws_security_group_rule" "app_port" {
  type              = "ingress"
  from_port         = var.app_port
  to_port           = var.app_port
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = data.aws_instance.existing.security_groups[0]
}

# Update Cloudflare record to point to specific port (if needed)
resource "cloudflare_record" "app_dns_port" {
  zone_id = var.cloudflare_zone_id
  name    = var.domain_name
  value   = var.ec2_public_ip
  type    = "A"
  ttl     = 1
  proxied = false  # Direct connection to port
  comment = "Link shortener service (port ${var.app_port})"
}