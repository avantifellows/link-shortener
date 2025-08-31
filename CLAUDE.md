# CLAUDE.md - Link Shortener Project

This file provides guidance to Claude Code when working on the Avanti Fellows Link Shortener service.

## Project Overview

A minimalist link shortener service hosted at `lnk.avantifellows.org/CODE` designed for internal use by Avanti Fellows team (~200 users max).

## Tech Stack

### Backend
- **Go 1.21+** - Compiled language for high performance
- **Chi Router** - Lightweight HTTP router and middleware
- **html/template** - Go's built-in template engine for HTML rendering
- **SQLite** - Lightweight database for URL mappings and analytics
- **database/sql + modernc.org/sqlite** - Pure Go SQLite driver

### Frontend
- **Go html/template** - Server-side rendered HTML
- **htmx** - High power tools for HTML (dynamic interactions)
- **Tailwind CSS** - Utility-first CSS framework (via CDN)
- **Minimal JavaScript** - htmx handles most dynamic behavior

### Infrastructure
- **EC2 t2.large** - Existing server for application hosting
- **Nginx** - Reverse proxy for domain routing to application port
- **SQLite** - Local database file for URL mappings and analytics
- **Systemd** - Service management for application lifecycle
- **Let's Encrypt** - SSL certificate for HTTPS encryption
- **Cloudflare** - DNS management for custom domain

### DevOps
- **Terraform** - Infrastructure as Code for deployment automation
- **GitHub Actions** - CI/CD pipeline with secrets management
- **GitHub Secrets** - Secure storage for certificates and sensitive config
- **Systemctl** - Service management and auto-restart
- **Nginx** - Web server and reverse proxy configuration
- **Let's Encrypt** - Free SSL certificate for HTTPS encryption

## Project Structure

```
link_shortening/
├── cmd/
│   └── server/
│       └── main.go          # Application entry point
├── internal/
│   ├── handlers/            # HTTP request handlers
│   ├── models/              # Data structures and database schemas
│   ├── database/            # Database operations and migrations
│   ├── services/            # Business logic (URL shortening, analytics)
│   └── middleware/          # HTTP middleware (auth, logging, etc.)
├── pkg/
│   └── utils/               # Shared utility functions
├── templates/
│   ├── base.html           # Base template with common layout
│   ├── dashboard.html      # Admin dashboard
│   ├── create.html         # Link creation form
│   └── analytics.html      # Link analytics view
├── static/
│   ├── css/               # Custom CSS files (if any)
│   └── js/                # JavaScript files
├── terraform/
│   ├── main.tf            # EC2 instance data and core resources
│   ├── app-deployment.tf  # Application deployment automation
│   ├── nginx.tf           # Nginx configuration management
│   ├── ssl.tf             # Cloudflare Origin Certificate setup
│   ├── variables.tf       # Configuration variables (GitHub secrets/variables)
│   └── outputs.tf         # Service URLs and deployment status
├── .github/
│   └── workflows/
│       └── deploy.yml     # GitHub Actions deployment workflow
├── internal/
│   ├── handlers/
│   │   └── *_test.go      # Handler unit tests
│   └── services/
│       └── *_test.go      # Service unit tests
├── deployment/
│   ├── templates/         # Terraform template files
│   ├── scripts/           # Deployment and setup scripts
│   └── configs/           # Configuration file templates
├── go.mod                 # Go module dependencies
├── go.sum                 # Go module checksums
├── TODOs.md              # Development task list
└── README.md             # Project documentation
```

## Core Features

### 1. URL Shortening
- Accept long URLs and generate short codes
- Support custom codes (if available)
- Store mappings in SQLite database

### 2. URL Redirection
- Fast redirects from short URLs to original URLs
- Track click analytics (timestamp, user agent, IP)
- Handle 404s for invalid codes

### 3. Analytics Dashboard
- View all created links
- Click statistics per link
- Basic authentication for access control

### 4. HTTP Endpoints
- `POST /shorten` - Create short links (htmx form submission)
- `GET /{code}` - Redirect to original URL
- `GET /dashboard` - Admin dashboard
- `GET /analytics` - Analytics data (htmx partial updates)
- `GET /health` - Health check endpoint

## Database Schema (SQLite)

### Table: `link_mappings`
```sql
CREATE TABLE link_mappings (
    short_code TEXT PRIMARY KEY,
    original_url TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    created_by TEXT,
    click_count INTEGER DEFAULT 0,
    last_accessed INTEGER,
    INDEX idx_created_at (created_at),
    INDEX idx_click_count (click_count)
);
```

### Table: `click_analytics`
```sql
CREATE TABLE click_analytics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    short_code TEXT NOT NULL,
    timestamp INTEGER NOT NULL,
    user_agent TEXT,
    ip_address TEXT,
    referrer TEXT,
    INDEX idx_short_code_timestamp (short_code, timestamp),
    FOREIGN KEY (short_code) REFERENCES link_mappings(short_code)
);
```

### Go Struct Models
```go
type LinkMapping struct {
    ShortCode    string    `json:"short_code" db:"short_code"`
    OriginalURL  string    `json:"original_url" db:"original_url"`
    CreatedAt    time.Time `json:"created_at" db:"created_at"`
    CreatedBy    string    `json:"created_by" db:"created_by"`
    ClickCount   int       `json:"click_count" db:"click_count"`
    LastAccessed *time.Time `json:"last_accessed" db:"last_accessed"`
}

type ClickAnalytics struct {
    ID        int       `json:"id" db:"id"`
    ShortCode string    `json:"short_code" db:"short_code"`
    Timestamp time.Time `json:"timestamp" db:"timestamp"`
    UserAgent string    `json:"user_agent" db:"user_agent"`
    IPAddress string    `json:"ip_address" db:"ip_address"`
    Referrer  string    `json:"referrer" db:"referrer"`
}
```

## Development Commands

```bash
# Development setup
go mod tidy
go run cmd/server/main.go

# Testing
go test ./...
go test -race ./...  # Race condition detection
go test -bench=. ./...  # Benchmark tests

# Build for production
go build -o link-shortener cmd/server/main.go
CGO_ENABLED=0 go build -ldflags="-s -w" -o link-shortener cmd/server/main.go  # Optimized build

# Production deployment (via GitHub Actions + Terraform)
git push origin main  # Triggers automated deployment

# Manual deployment (if needed)
cd terraform/
terraform init
terraform apply  # Uses GitHub secrets as TF_VAR_* environment variables

# Manual service management (if needed)
sudo systemctl status link-shortener
sudo systemctl restart link-shortener
```

## Environment Variables

```bash
# Application Configuration (stored in code)
DATABASE_PATH=/var/lib/link-shortener/database.db
PORT=8080
DEBUG=false
LOG_LEVEL=INFO
SSL_CERT_PATH=/etc/letsencrypt/live/lnk.avantifellows.org/fullchain.pem
SSL_KEY_PATH=/etc/letsencrypt/live/lnk.avantifellows.org/privkey.pem
GRACEFUL_SHUTDOWN_TIMEOUT=30s
```

## GitHub Secrets (Repository Settings → Secrets and Variables → Actions)

### Secrets (Sensitive Data) 🔐
```bash
# Authentication
ADMIN_PASSWORD_HASH=<bcrypt_hash>

# SSH Access
EC2_SSH_PRIVATE_KEY=<ssh_private_key_content>
```

### Variables (Non-sensitive Config) 📝
```bash
# Deployment Configuration
BASE_URL=https://lnk.avantifellows.org
EC2_INSTANCE_ID=i-1234567890abcdef0
EC2_PUBLIC_IP=<instance_public_ip>
ADMIN_USERNAME=admin
```

## Authentication Strategy

- Simple HTTP Basic Auth for dashboard access
- GitHub Secrets for secure admin credential storage
- No user registration - single admin account
- JWT tokens for API access (optional for future)

## Open Source Considerations

- All sensitive configuration externalized to GitHub Secrets
- Contributors can deploy to their own infrastructure with their own secrets
- Clear separation between public code and private configuration
- Fork-safe deployment (forks don't inherit secrets)

## Deployment Strategy (GitHub Actions + Terraform)

1. **Secrets Setup**: Configure GitHub Secrets and Variables (one-time setup)
2. **Code Push**: Push to main branch triggers GitHub Actions workflow
3. **Infrastructure**: Terraform manages EC2 instance configuration using GitHub secrets
4. **Application Deployment**: Terraform uploads code and installs dependencies
5. **Database Setup**: Terraform initializes SQLite database with proper permissions
6. **SSL Certificate**: Terraform installs Cloudflare Origin Certificate from GitHub secrets
7. **Nginx Configuration**: Terraform generates and deploys reverse proxy config with SSL
8. **Systemd Service**: Terraform creates and enables application service
9. **Domain DNS**: Manual Cloudflare DNS update to point to EC2 IP (one-time)
10. **Validation**: Terraform verifies all endpoints and services are running

## Performance Requirements

- **Redirect Response Time**: < 10ms (Go performance)
- **Concurrent Users**: Support 1000+ users per instance
- **Memory Usage**: < 5MB RAM per instance
- **Availability**: 99.9% uptime target
- **Rate Limiting**: 1000 requests/minute per IP
- **Startup Time**: < 100ms (compiled binary)

## Security Considerations

- Input validation for URLs
- Rate limiting on API endpoints
- Basic authentication for admin functions
- HTTPS only (enforced by Cloudflare)
- HTTPS encryption with Let's Encrypt SSL certificate
- GitHub Secrets for sensitive configuration (passwords, SSH keys)
- No sensitive data in code repository or logs
- Automatic SSL certificate renewal via certbot
- SSH key-based authentication for deployment

## Monitoring & Alerting

- Systemd journal logs for application events
- Nginx access logs for request tracking
- SQLite database file monitoring
- Let's Encrypt SSL certificate validity monitoring
- Discord webhook for critical alerts (following AF pattern)
- Logrotate for log file management

## Testing Strategy

- Unit tests for handlers and services (`go test`)
- Integration tests for SQLite operations
- Benchmark tests for performance validation (`go test -bench`)
- Race condition detection (`go test -race`)
- End-to-end tests for full redirect flow
- Load testing with high concurrency (Go's strength)

## Rollback Strategy

- Terraform state management for infrastructure rollbacks
- Git-based application versioning for code rollbacks
- Automated SQLite database backups via Terraform
- Let's Encrypt certificate auto-renewal
- Terraform-managed nginx configuration versioning
- Systemd service management through Terraform
- DNS TTL management for quick domain changes