# Data source to get information about existing EC2 instance
data "aws_instance" "existing" {
  instance_id = var.ec2_instance_id
}

# Cloudflare DNS record pointing to EC2 instance
resource "cloudflare_record" "app_dns" {
  zone_id = var.cloudflare_zone_id
  name    = var.domain_name
  value   = var.ec2_public_ip
  type    = "A"
  ttl     = 1
  proxied = true
  comment = "Link shortener service"
}

# Build the Go application locally
resource "null_resource" "build_app" {
  triggers = {
    # Rebuild when source code changes
    source_hash = filemd5("../cmd/server/main.go")
  }

  provisioner "local-exec" {
    command = <<-EOT
      cd ..
      GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o terraform/link-shortener cmd/server/main.go
    EOT
  }
}

# Deploy application files to EC2
resource "null_resource" "deploy_app" {
  depends_on = [null_resource.build_app]

  triggers = {
    # Redeploy when binary changes
    binary_hash = filemd5("link-shortener")
    config_hash = md5(jsonencode({
      auth_token = var.auth_token
      app_port   = var.app_port
      domain     = var.domain_name
    }))
  }

  connection {
    type        = "ssh"
    user        = var.ssh_user
    private_key = file(var.ssh_private_key_path)
    host        = var.ec2_public_ip
    timeout     = "5m"
  }

  # Create application directory
  provisioner "remote-exec" {
    inline = [
      "sudo mkdir -p /opt/link-shortener",
      "sudo mkdir -p /var/lib/link-shortener",
      "sudo mkdir -p /etc/ssl/cloudflare",
      "sudo chown -R ${var.ssh_user}:${var.ssh_user} /opt/link-shortener",
      "sudo chown -R ${var.ssh_user}:${var.ssh_user} /var/lib/link-shortener"
    ]
  }

  # Upload application binary
  provisioner "file" {
    source      = "link-shortener"
    destination = "/opt/link-shortener/link-shortener"
  }

  # Upload templates
  provisioner "file" {
    source      = "../templates/"
    destination = "/opt/link-shortener/"
  }

  # Upload static files if they exist
  provisioner "file" {
    source      = "../static/"
    destination = "/opt/link-shortener/"
  }

  # Create environment file
  provisioner "file" {
    content = templatefile("${path.module}/templates/env.tpl", {
      auth_token    = var.auth_token
      database_path = "/var/lib/link-shortener/database.db"
      port          = var.app_port
      base_url      = "https://${var.domain_name}"
    })
    destination = "/opt/link-shortener/.env"
  }

  # Set permissions
  provisioner "remote-exec" {
    inline = [
      "chmod +x /opt/link-shortener/link-shortener",
      "chmod 600 /opt/link-shortener/.env"
    ]
  }
}

# Install SSL certificates
resource "null_resource" "install_ssl" {
  connection {
    type        = "ssh"
    user        = var.ssh_user
    private_key = file(var.ssh_private_key_path)
    host        = var.ec2_public_ip
    timeout     = "5m"
  }

  # Upload SSL certificates
  provisioner "file" {
    content     = var.cloudflare_origin_cert
    destination = "/tmp/origin-cert.pem"
  }

  provisioner "file" {
    content     = var.cloudflare_origin_key
    destination = "/tmp/origin-key.pem"
  }

  # Install certificates
  provisioner "remote-exec" {
    inline = [
      "sudo mv /tmp/origin-cert.pem /etc/ssl/cloudflare/",
      "sudo mv /tmp/origin-key.pem /etc/ssl/cloudflare/",
      "sudo chmod 600 /etc/ssl/cloudflare/*",
      "sudo chown root:root /etc/ssl/cloudflare/*"
    ]
  }
}

# Create systemd service
resource "null_resource" "create_service" {
  depends_on = [null_resource.deploy_app]

  connection {
    type        = "ssh"
    user        = var.ssh_user
    private_key = file(var.ssh_private_key_path)
    host        = var.ec2_public_ip
    timeout     = "5m"
  }

  # Create systemd service file
  provisioner "file" {
    content = templatefile("${path.module}/templates/link-shortener.service.tpl", {
      app_port = var.app_port
      ssh_user = var.ssh_user
    })
    destination = "/tmp/link-shortener.service"
  }

  # Install and start service
  provisioner "remote-exec" {
    inline = [
      "sudo mv /tmp/link-shortener.service /etc/systemd/system/",
      "sudo systemctl daemon-reload",
      "sudo systemctl enable link-shortener",
      "sudo systemctl restart link-shortener",
      "sleep 5",
      "sudo systemctl status link-shortener"
    ]
  }
}

# Configure Nginx (if needed)
resource "null_resource" "configure_nginx" {
  depends_on = [null_resource.create_service, null_resource.install_ssl]

  connection {
    type        = "ssh"
    user        = var.ssh_user
    private_key = file(var.ssh_private_key_path)
    host        = var.ec2_public_ip
    timeout     = "5m"
  }

  # Create nginx configuration
  provisioner "file" {
    content = templatefile("${path.module}/templates/nginx.conf.tpl", {
      domain_name = var.domain_name
      app_port    = var.app_port
    })
    destination = "/tmp/link-shortener.conf"
  }

  # Install nginx and configure
  provisioner "remote-exec" {
    inline = [
      "sudo apt-get update",
      "sudo apt-get install -y nginx",
      "sudo mv /tmp/link-shortener.conf /etc/nginx/sites-available/",
      "sudo ln -sf /etc/nginx/sites-available/link-shortener.conf /etc/nginx/sites-enabled/",
      "sudo rm -f /etc/nginx/sites-enabled/default",
      "sudo nginx -t",
      "sudo systemctl restart nginx",
      "sudo systemctl enable nginx"
    ]
  }
}