# Alternative main.tf that skips Nginx configuration
# Use this if you have existing Nginx setup

# All the same resources as main.tf EXCEPT remove this block:
# resource "null_resource" "configure_nginx" { ... }

# Instead, just deploy the application and let you configure Nginx manually
resource "null_resource" "nginx_instructions" {
  depends_on = [null_resource.create_service]

  provisioner "local-exec" {
    command = <<-EOT
      echo "================================================"
      echo "APPLICATION DEPLOYED SUCCESSFULLY!"
      echo "================================================"
      echo "Service running on: http://localhost:${var.app_port}"
      echo "Domain: ${var.domain_name}"
      echo ""
      echo "MANUAL NGINX CONFIGURATION NEEDED:"
      echo "1. Add this to your existing nginx config:"
      echo ""
      echo "server {"
      echo "    listen 443 ssl;"
      echo "    server_name ${var.domain_name};"
      echo "    location / {"
      echo "        proxy_pass http://127.0.0.1:${var.app_port};"
      echo "        proxy_set_header Host \$host;"
      echo "        proxy_set_header X-Real-IP \$remote_addr;"
      echo "        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;"
      echo "        proxy_set_header X-Forwarded-Proto \$scheme;"
      echo "    }"
      echo "}"
      echo ""
      echo "2. Test and reload nginx:"
      echo "   sudo nginx -t && sudo systemctl reload nginx"
      echo "================================================"
    EOT
  }
}