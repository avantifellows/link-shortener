# Add this to main.tf to check existing nginx configuration

# Check existing nginx setup before deployment
resource "null_resource" "check_nginx" {
  connection {
    type        = "ssh"
    user        = var.ssh_user
    private_key = file(var.ssh_private_key_path)
    host        = var.ec2_public_ip
    timeout     = "5m"
  }

  provisioner "remote-exec" {
    inline = [
      "echo '=== EXISTING NGINX CHECK ==='",
      "if systemctl is-active --quiet nginx; then",
      "  echo 'Nginx is running'",
      "  echo 'Existing sites:'",
      "  ls -la /etc/nginx/sites-enabled/ || echo 'No sites-enabled directory'",
      "  echo 'Listening ports:'",
      "  sudo netstat -tlnp | grep nginx || echo 'No nginx processes found'",
      "else",
      "  echo 'Nginx is not running'",
      "fi",
      "echo '=== END CHECK ==='",
      "",
      "read -p 'Continue with deployment? (y/N) ' -n 1 -r",
      "echo",
      "if [[ ! $REPLY =~ ^[Yy]$ ]]; then",
      "  echo 'Deployment cancelled'",
      "  exit 1",
      "fi"
    ]
  }
}