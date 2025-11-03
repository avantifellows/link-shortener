#!/bin/bash
set -e

# Load deployment configuration
if [ -f ".env.deploy.local" ]; then
    echo "📋 Loading local deployment config..."
    source .env.deploy.local
elif [ -f ".env.deploy" ]; then
    echo "📋 Loading deployment config..."
    source .env.deploy
else
    echo "❌ Error: No deployment configuration found!"
    echo "   Please create .env.deploy.local with your server details"
    echo "   You can copy .env.deploy as a template"
    exit 1
fi

# Validate required variables
if [ -z "$SERVER" ] || [ -z "$KEY_PATH" ] || [ -z "$DEPLOY_DIR" ] || [ -z "$SERVICE_NAME" ]; then
    echo "❌ Error: Missing required deployment configuration!"
    echo "   Required: SERVER, KEY_PATH, DEPLOY_DIR, SERVICE_NAME"
    exit 1
fi

echo "🎯 Deploying to: $SERVER"

echo "🚀 Deploying Link Shortener to server..."

# Build binary for Linux
echo "🔨 Building binary for Linux..."
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o link-shortener-linux cmd/server/main.go

# Create deployment directory
ssh -i $KEY_PATH $SERVER "sudo mkdir -p $DEPLOY_DIR"

# Upload binary
echo "📦 Uploading binary..."
scp -i $KEY_PATH link-shortener-linux $SERVER:/tmp/
ssh -i $KEY_PATH $SERVER "sudo mv /tmp/link-shortener-linux $DEPLOY_DIR/link-shortener && sudo chmod +x $DEPLOY_DIR/link-shortener"

# Upload templates, static files, and environment config
echo "📁 Uploading templates, static files, and environment config..."
scp -i $KEY_PATH -r templates $SERVER:/tmp/
scp -i $KEY_PATH -r static $SERVER:/tmp/
scp -i $KEY_PATH .env.production $SERVER:/tmp/
ssh -i $KEY_PATH $SERVER "sudo rm -rf $DEPLOY_DIR/templates $DEPLOY_DIR/static && sudo mv /tmp/templates $DEPLOY_DIR/ && sudo mv /tmp/static $DEPLOY_DIR/ && sudo mv /tmp/.env.production $DEPLOY_DIR/.env.local"

# Create systemd service file
echo "⚙️  Creating systemd service..."
cat > /tmp/link-shortener.service << 'EOF'
[Unit]
Description=Link Shortener Service
After=network.target

[Service]
Type=simple
User=ubuntu
WorkingDirectory=/opt/link-shortener
EnvironmentFile=/opt/link-shortener/.env.local
ExecStart=/opt/link-shortener/link-shortener
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target
EOF

scp -i $KEY_PATH /tmp/link-shortener.service $SERVER:/tmp/
ssh -i $KEY_PATH $SERVER "sudo mv /tmp/link-shortener.service /etc/systemd/system/"

# Create database directory
ssh -i $KEY_PATH $SERVER "sudo mkdir -p /var/lib/link-shortener && sudo chown ubuntu:ubuntu /var/lib/link-shortener"

# Set ownership
ssh -i $KEY_PATH $SERVER "sudo chown -R ubuntu:ubuntu $DEPLOY_DIR"

# Enable and start service
echo "🏁 Starting service..."
ssh -i $KEY_PATH $SERVER "sudo systemctl daemon-reload"
ssh -i $KEY_PATH $SERVER "sudo systemctl enable $SERVICE_NAME"
ssh -i $KEY_PATH $SERVER "sudo systemctl restart $SERVICE_NAME"

# Extract domain name from .env.production
DOMAIN_NAME=$(grep "^BASE_URL=" .env.production | cut -d'=' -f2 | sed 's|https://||' | sed 's|http://||')

# Configure nginx
echo "🌐 Configuring nginx for domain: $DOMAIN_NAME..."

# Create proper main nginx.conf
cat > /tmp/nginx.conf << EOF
user www-data;
worker_processes auto;
pid /run/nginx.pid;
include /etc/nginx/modules-enabled/*.conf;

events {
    worker_connections 768;
    # multi_accept on;
}

http {
    sendfile on;
    tcp_nopush on;
    tcp_nodelay on;
    keepalive_timeout 65;
    types_hash_max_size 2048;

    include /etc/nginx/mime.types;
    default_type application/octet-stream;

    ssl_protocols TLSv1 TLSv1.1 TLSv1.2 TLSv1.3;
    ssl_prefer_server_ciphers on;

    access_log /var/log/nginx/access.log;
    error_log /var/log/nginx/error.log;

    gzip on;

    include /etc/nginx/conf.d/*.conf;
    include /etc/nginx/sites-enabled/*;
}
EOF

# Create nginx configuration for Let's Encrypt SSL
cat > /tmp/link-shortener-nginx << EOF
# Redirect HTTP to HTTPS
server {
    listen 80;
    server_name $DOMAIN_NAME;
    return 301 https://\$server_name\$request_uri;
}

# HTTPS server with Let's Encrypt SSL
server {
    listen 443 ssl;
    server_name $DOMAIN_NAME;

    # Let's Encrypt SSL certificates
    ssl_certificate /etc/letsencrypt/live/$DOMAIN_NAME/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/$DOMAIN_NAME/privkey.pem;
    
    # SSL configuration
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-RSA-AES128-SHA256:ECDHE-RSA-AES256-SHA384;
    ssl_prefer_server_ciphers on;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 10m;

    # Security headers
    add_header X-Frame-Options DENY always;
    add_header X-Content-Type-Options nosniff always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;

    # Proxy to Link Shortener application
    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto https;
        proxy_set_header X-Forwarded-Host \$server_name;
        proxy_connect_timeout 30s;
        proxy_send_timeout 30s;
        proxy_read_timeout 30s;
    }

    # Health check endpoint
    location /health {
        proxy_pass http://127.0.0.1:8080/health;
        access_log off;
    }
}
EOF

# Upload nginx configs
scp -i $KEY_PATH /tmp/nginx.conf $SERVER:/tmp/
scp -i $KEY_PATH /tmp/link-shortener-nginx $SERVER:/tmp/

# Backup and replace nginx configuration
ssh -i $KEY_PATH $SERVER "
    # Backup existing configurations
    sudo mkdir -p /etc/nginx/backup-$(date +%Y%m%d_%H%M%S)
    
    # Backup main nginx.conf and site configs
    sudo cp /etc/nginx/nginx.conf /etc/nginx/backup-$(date +%Y%m%d_%H%M%S)/nginx.conf
    if [ -f /etc/nginx/sites-available/link-shortener ]; then
        sudo cp /etc/nginx/sites-available/link-shortener /etc/nginx/backup-$(date +%Y%m%d_%H%M%S)/
    fi
    
    # Remove old temp configurations (common naming patterns)
    sudo rm -f /etc/nginx/sites-enabled/temp-link-shortener
    sudo rm -f /etc/nginx/sites-available/temp-link-shortener
    sudo rm -f /etc/nginx/sites-enabled/link-shortener-temp
    sudo rm -f /etc/nginx/sites-available/link-shortener-temp
    sudo rm -f /etc/nginx/sites-enabled/temp-lnk*
    sudo rm -f /etc/nginx/sites-available/temp-lnk*
    
    # Install fixed main nginx.conf
    sudo mv /tmp/nginx.conf /etc/nginx/nginx.conf
    
    # Install new site configuration
    sudo mv /tmp/link-shortener-nginx /etc/nginx/sites-available/link-shortener
    sudo ln -sf /etc/nginx/sites-available/link-shortener /etc/nginx/sites-enabled/
    
    # Test nginx configuration
    if sudo nginx -t; then
        echo '✅ Nginx configuration test passed'
        sudo systemctl reload nginx
        echo '🔄 Nginx reloaded successfully'
    else
        echo '❌ Nginx configuration test failed!'
        exit 1
    fi
"

# Check service status
echo "✅ Checking service status..."
ssh -i $KEY_PATH $SERVER "sudo systemctl status $SERVICE_NAME --no-pager -l"

# Verify nginx is running
echo "🌐 Checking nginx status..."
ssh -i $KEY_PATH $SERVER "sudo systemctl status nginx --no-pager -l | head -10"

echo ""
echo "🎉 Deployment complete!"
echo "📊 Application: http://65.0.246.88:8080 (direct)"
echo "🌐 Production: https://$DOMAIN_NAME"
echo "📋 Dashboard: https://$DOMAIN_NAME/dashboard"
echo "🔍 Analytics: https://$DOMAIN_NAME/analytics"
echo ""
echo "🧪 Test the deployment:"
echo "curl -s https://$DOMAIN_NAME/health"