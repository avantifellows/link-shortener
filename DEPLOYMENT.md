# Deployment Guide

This guide explains how to deploy the Link Shortener service to your own server.

## Prerequisites

- Linux server with SSH access
- Go 1.21+ installed locally (for building the binary)
- nginx installed on the server
- Domain name pointed to your server IP

## Setup Deployment Configuration

1. **Copy the deployment template:**
   ```bash
   cp .env.deploy .env.deploy.local
   ```

2. **Edit `.env.deploy.local` with your server details:**
   ```bash
   # Server SSH connection
   SERVER=ubuntu@YOUR_SERVER_IP
   KEY_PATH=~/path/to/your/private/key.pem
   
   # Deployment paths (usually don't need to change these)
   DEPLOY_DIR=/opt/link-shortener
   SERVICE_NAME=link-shortener
   ```

3. **Update `.env.production` with your domain:**
   ```bash
   BASE_URL=https://your-domain.com
   PORT=8080
   DATABASE_PATH=/var/lib/link-shortener/database.db
   AUTH_TOKEN=your-secure-random-token
   ```

## Build and Deploy

1. **Build the application:**
   ```bash
   CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o link-shortener-linux cmd/server/main.go
   ```

2. **Deploy to server:**
   ```bash
   ./deploy.sh
   ```

The deployment script will:
- Upload the binary and static files
- Create systemd service
- Configure nginx with SSL support
- Start the service
- Verify deployment

## Database Setup

If you have existing data to import:

```bash
./dump_db.sh
```

This will upload your local database to the server.

## Cloudflare SSL Configuration

The nginx configuration is set up for **Cloudflare Flexible SSL**:
- Browser → Cloudflare: HTTPS
- Cloudflare → Server: HTTP
- No SSL certificates needed on server

If using a different SSL setup, modify the nginx configuration in `deploy.sh`.

## Troubleshooting

**"No deployment configuration found"**
- Make sure you created `.env.deploy.local` with your server details

**"Permission denied (publickey)"**
- Check that your SSH key path is correct in `.env.deploy.local`
- Ensure the private key has correct permissions: `chmod 600 /path/to/key.pem`

**"Too many redirects"**
- This usually means SSL configuration mismatch
- For Cloudflare Flexible SSL, the current config should work
- For other SSL setups, you may need to modify nginx config

**Service not starting**
- Check logs: `ssh user@server "sudo journalctl -u link-shortener -f"`
- Verify database directory permissions
- Check that all required environment variables are set

## Manual Service Management

Connect to your server and manage the service:

```bash
# Check status
sudo systemctl status link-shortener

# View logs
sudo journalctl -u link-shortener -f

# Restart service
sudo systemctl restart link-shortener

# Stop service
sudo systemctl stop link-shortener
```