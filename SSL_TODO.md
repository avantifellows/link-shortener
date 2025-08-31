# SSL Configuration TODO

## Current Issue
- Intermittent empty responses from `/shorten` API endpoint
- Root cause: Cloudflare Flexible SSL causing connection issues between Cloudflare proxy and nginx
- Direct server access (bypassing Cloudflare) works consistently

## Solution: Switch to Direct SSL with Let's Encrypt

### Steps to Fix

1. **Install SSL Certificate on Server**
   ```bash
   # SSH to server
   ssh -i ~/certs/AvantiFellows.pem ubuntu@65.0.246.88
   
   # Install certbot
   sudo apt install certbot python3-certbot-nginx
   
   # Get Let's Encrypt certificate (auto-configures nginx)
   sudo certbot --nginx -d lnk.avantifellows.org
   ```

2. **Turn Off Cloudflare Proxy**
   - Go to Cloudflare DNS settings
   - Click orange cloud next to `lnk` A record → should turn gray
   - This makes Cloudflare DNS-only (no proxy)

3. **Verify Configuration**
   ```bash
   # Test HTTPS works
   curl -I https://lnk.avantifellows.org/health
   
   # Test API endpoint consistency
   for i in {1..10}; do curl -s -X POST https://lnk.avantifellows.org/shorten -H "Accept: 
   ```

### Let's Encrypt Auto-Renewal

**Automatic Setup:**
- Certbot automatically creates systemd timer for renewal
- Checks twice daily, renews when 30 days left (certs last 90 days)
- Auto-reloads nginx after successful renewal

**Verification Commands:**
```bash
# Check auto-renewal status
sudo systemctl status certbot.timer

# Test renewal process (dry run)
sudo certbot renew --dry-run

# View renewal logs
sudo journalctl -u certbot.timer
```

**Maintenance:** Zero maintenance required in normal operation.

## Benefits of This Approach

✅ **Fixes intermittent failures** - eliminates Cloudflare proxy connection issues  
✅ **Automatic certificate management** - Let's Encrypt handles renewals  
✅ **Direct connection** - simpler troubleshooting  
✅ **Free SSL certificates** - no cost  
✅ **Standard setup** - widely used and supported  

## Trade-offs

❌ **No DDoS protection** - server handles all traffic directly  
❌ **No CDN benefits** - slower for global users  
❌ **No Cloudflare analytics** - lose traffic insights  
❌ **Certificate management responsibility** - though automated  

## Future: Re-enable Cloudflare DDoS Protection

**Yes, you can re-enable Cloudflare proxy later!**

When you want DDoS protection back:

1. **Keep Let's Encrypt certificate** on server
2. **Turn Cloudflare proxy back ON** (gray cloud → orange cloud)
3. **Change Cloudflare SSL mode** from "Flexible" to "Full (strict)"

**How "Full (strict)" works:**
- Client → HTTPS → Cloudflare → HTTPS → Your server
- Cloudflare validates your Let's Encrypt certificate
- Both connections encrypted end-to-end
- No more connection issues since both sides use HTTPS

**Best of both worlds:**
- ✅ Cloudflare DDoS protection + CDN
- ✅ Let's Encrypt certificate (your control)
- ✅ End-to-end encryption
- ✅ No proxy connection issues

## Recommendation

1. **Start with direct SSL** (fix current issues)
2. **Monitor traffic and attacks** 
3. **Re-enable Cloudflare proxy with Full SSL** if you need DDoS protection
4. **Keep Let's Encrypt** - it works with both approaches

## Current Server Details

- **Server IP:** 65.0.246.88
- **SSH:** `ssh -i ~/certs/AvantiFellows.pem ubuntu@65.0.246.88`
- **Go App:** Running on port 8080
- **Domain:** lnk.avantifellows.org
- **Current Issue:** Flexible SSL proxy causing intermittent API failures

## Notes

- Let's Encrypt certificates work perfectly with Cloudflare's "Full (strict)" mode
- This gives you flexibility to switch between direct SSL and Cloudflare proxy as needed
- Current nginx config only needs certbot to add SSL configuration
- No code changes needed in Go application
