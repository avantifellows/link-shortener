# Terraform Deployment Guide

This Terraform configuration deploys the Link Shortener application to an existing EC2 instance with Cloudflare DNS and SSL.

## Prerequisites

1. **Existing EC2 instance** with SSH access
2. **Cloudflare account** with domain management
3. **AWS CLI** configured with appropriate permissions
4. **Terraform** installed (>= 1.0)

## Setup Steps

### 1. Configure Variables
```bash
# Copy the example file
cp terraform.tfvars.example terraform.tfvars

# Edit with your values
vim terraform.tfvars
```

### 2. Get Required Information

#### EC2 Instance Details:
```bash
# Get your instance ID and IP
aws ec2 describe-instances --query 'Reservations[*].Instances[*].[InstanceId,PublicIpAddress,State.Name]' --output table
```

#### Cloudflare Zone ID:
```bash
# List your zones
curl -X GET "https://api.cloudflare.com/client/v4/zones" \
  -H "Authorization: Bearer YOUR_API_TOKEN" \
  -H "Content-Type: application/json"
```

#### Cloudflare Origin Certificate:
1. Go to Cloudflare Dashboard
2. Select your domain
3. Go to SSL/TLS > Origin Server
4. Click "Create Certificate"
5. Copy both certificate and private key

### 3. Deploy

```bash
# Initialize Terraform
terraform init

# Review the plan
terraform plan

# Apply the configuration
terraform apply
```

## What This Deploys

### Application Components:
- ✅ Compiled Go binary
- ✅ Template files
- ✅ Environment configuration
- ✅ Systemd service
- ✅ Auto-start on boot

### Infrastructure Components:
- ✅ Cloudflare DNS A record
- ✅ Nginx reverse proxy
- ✅ SSL certificates (Cloudflare Origin)
- ✅ Security headers and rate limiting

### Security Features:
- ✅ HTTPS-only access
- ✅ Rate limiting (10 req/s)
- ✅ Security headers
- ✅ Restricted file permissions
- ✅ Systemd security settings

## Post-Deployment

### Verify Deployment:
```bash
# Check application status
terraform output

# SSH to instance and verify
ssh -i ~/.ssh/your-key.pem ubuntu@YOUR_IP
sudo systemctl status link-shortener
sudo journalctl -u link-shortener -f
```

### Test Application:
```bash
# Health check
curl https://lnk.avantifellows.org/health

# Create short link (with auth)
curl -X POST https://lnk.avantifellows.org/shorten \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "original_url=https://example.com"
```

## Troubleshooting

### Common Issues:

1. **SSH Connection Failed**
   - Check security group allows SSH (port 22)
   - Verify SSH key permissions: `chmod 600 ~/.ssh/your-key.pem`

2. **Service Won't Start**
   - Check logs: `sudo journalctl -u link-shortener -f`
   - Verify binary permissions: `ls -la /opt/link-shortener/`

3. **Nginx Issues**
   - Test config: `sudo nginx -t`
   - Check error logs: `sudo tail -f /var/log/nginx/error.log`

4. **DNS Not Resolving**
   - Verify Cloudflare zone ID
   - Check DNS propagation: `dig lnk.avantifellows.org`

### Rollback:
```bash
# Destroy resources (keeps EC2 instance)
terraform destroy

# Or just restart service with previous version
sudo systemctl restart link-shortener
```

## File Structure
```
terraform/
├── main.tf                     # Main deployment logic
├── variables.tf               # Input variables
├── outputs.tf                 # Output values
├── providers.tf               # Provider configuration
├── terraform.tfvars.example   # Example configuration
├── templates/
│   ├── env.tpl               # Environment file template
│   ├── link-shortener.service.tpl  # Systemd service
│   └── nginx.conf.tpl        # Nginx configuration
└── README.md                 # This file
```