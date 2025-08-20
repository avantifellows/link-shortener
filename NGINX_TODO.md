# NGINX Configuration Assessment for Multi-Domain Setup

## 🎯 Goal
Deploy `lnk.avantifellows.org` (link shortener) alongside existing `scripts.avantifellows.org` on the same EC2 instance without conflicts.

## 📋 Server Assessment Checklist

### 1. Current Nginx Status
```bash
# Check if nginx is installed and running
systemctl status nginx
sudo nginx -v

# Test current configuration
sudo nginx -t

# Check nginx process and memory usage
ps aux | grep nginx
```

### 2. Existing Site Configuration
```bash
# List all enabled sites
ls -la /etc/nginx/sites-enabled/

# Show existing site configurations
cat /etc/nginx/sites-enabled/*

# Check main nginx config
head -30 /etc/nginx/nginx.conf
```

### 3. Port Usage Analysis
```bash
# Check all listening ports
sudo netstat -tlnp

# Specifically check web ports
sudo netstat -tlnp | grep ':80\|:443\|:3000\|:8080'

# Check if any process is using our target port 8080
sudo lsof -i :8080
```

### 4. SSL Certificate Setup
```bash
# Check existing SSL certificates
ls -la /etc/ssl/
ls -la /etc/letsencrypt/ 2>/dev/null || echo "No Let's Encrypt found"

# Find nginx SSL configuration
grep -r "ssl_certificate" /etc/nginx/ 2>/dev/null

# Check certificate details (if found)
# openssl x509 -in /path/to/cert.pem -text -noout | head -20
```

### 5. Current Applications
```bash
# See what applications are running
ps aux | grep -E "(node|python|go|java|nginx)"

# Check systemd services
systemctl list-units --type=service --state=running | grep -v "@"

# Check for existing avantifellows services
systemctl list-units | grep -i avanti
```

### 6. Security Groups & Firewall
```bash
# Check local firewall (if any)
sudo ufw status 2>/dev/null || echo "UFW not active"
sudo iptables -L 2>/dev/null | head -10

# Check which user nginx runs as
ps aux | grep nginx | head -2
```

### 7. Directory Permissions
```bash
# Check nginx directories
ls -la /etc/nginx/
ls -la /var/log/nginx/
ls -la /var/www/ 2>/dev/null || echo "No /var/www"

# Check if /opt directory exists (for our app)
ls -la /opt/
```

## 📊 Information to Document

### Current Setup Analysis:
- [ ] **Nginx Version**: ________________
- [ ] **Nginx Status**: Running / Stopped / Not Installed
- [ ] **scripts.avantifellows.org port**: ________________
- [ ] **SSL Certificate type**: Let's Encrypt / Cloudflare Origin / Self-signed / Other
- [ ] **Certificate location**: ________________
- [ ] **Available ports**: ________________

### Current Domain Configuration:
```
# Copy paste the output from sites-enabled here:



```

### Current SSL Configuration:
```
# Copy paste SSL-related nginx config here:



```

### Port Usage Summary:
```
# Copy paste netstat output here:



```

## 🚀 Deployment Strategy Decision Matrix

Based on findings, choose one:

### Option 1: Add to Existing Nginx ✅ (Recommended if)
- [ ] Nginx is already configured for scripts.avantifellows.org
- [ ] SSL certificates are working (Let's Encrypt or Cloudflare Origin)
- [ ] No conflicts with existing configuration
- [ ] Port 8080 is available for link shortener app

**Terraform modifications needed:**
- [ ] Use existing SSL certificates
- [ ] Add only lnk-avantifellows site config
- [ ] Preserve all existing configurations

### Option 2: Port-Only Deployment ✅ (Recommended if)
- [ ] Complex existing nginx setup
- [ ] Don't want to risk breaking scripts.avantifellows.org
- [ ] Prefer isolation between services

**Terraform modifications needed:**
- [ ] Skip nginx configuration entirely
- [ ] Open port 8080 in security group
- [ ] Use Cloudflare proxy for SSL

### Option 3: Separate Nginx Instance ⚠️ (Only if needed)
- [ ] Existing nginx can't be modified
- [ ] Need complete isolation
- [ ] Resources allow multiple nginx instances

## 🔧 Terraform Configuration Tasks

Based on assessment, complete these:

### High Priority:
- [ ] Remove automatic nginx installation from terraform
- [ ] Modify main.tf to add only lnk site configuration
- [ ] Update variables.tf for existing SSL cert paths
- [ ] Create conditional resources based on nginx status

### Medium Priority:  
- [ ] Add pre-deployment validation checks
- [ ] Create backup/rollback procedures
- [ ] Update security group rules appropriately

### Low Priority:
- [ ] Add monitoring for both services
- [ ] Optimize nginx configuration for both domains
- [ ] Set up log rotation and management

## 📝 Notes & Decisions

### Assessment Results:
```
# Document findings here after server check



```

### Chosen Strategy:
- [ ] **Option 1**: Add to existing nginx
- [ ] **Option 2**: Port-only deployment  
- [ ] **Option 3**: Separate nginx instance

### Reasoning:
```
# Document decision reasoning here



```

### Next Steps:
1. [ ] Complete server assessment
2. [ ] Choose deployment strategy
3. [ ] Modify terraform configuration
4. [ ] Test deployment in staging/dev
5. [ ] Deploy to production

---

**Date Completed**: _____________  
**Completed By**: _____________  
**Approved By**: _____________
