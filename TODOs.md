# Link Shortener TODOs

## Project Overview
Building a minimalist open-source link shortener service at `lnk.avantifellows.org/CODE` using Go, htmx, SQLite, EC2, Cloudflare, and GitHub Actions for deployment.

## Development Tasks

### 1. Project Setup & Core Application
- [x] Initialize Go module and set up project structure with internal/ and cmd/ directories
- [x] Create SQLite database schema and Go struct models for URL mappings and analytics
- [x] Set up Chi router with middleware (logging, recovery, CORS)
- [x] Implement core link shortening endpoints (POST /shorten, GET /{code})
- [x] Build click tracking and analytics collection functionality
- [x] Set up graceful shutdown and health check endpoint

### 2. Frontend & UI (Go + htmx)
- [x] Create Go html/template templates for dashboard UI
- [x] Implement htmx-powered link creation form with real-time feedback
- [x] Build htmx-powered analytics dashboard with live updates
- [ ] Add basic HTTP authentication middleware for dashboard access
- [x] Set up static file serving for Tailwind CSS and htmx assets
- [x] Implement htmx partial page updates for dynamic content

### 3. Infrastructure & Deployment (GitHub Actions + Terraform)
- [ ] Set up GitHub repository secrets for SSL certificates and sensitive config
- [ ] Configure GitHub repository variables for deployment configuration
- [ ] Create GitHub Actions workflow for automated deployment
- [ ] Create Terraform configuration for EC2 instance management
- [ ] Set up Terraform templates for nginx reverse proxy configuration
- [ ] Configure Terraform to use GitHub secrets for SSL certificate deployment
- [ ] Create Terraform resources for systemd service management
- [ ] Set up Terraform null_resource for application deployment automation
- [ ] Configure Terraform local-exec provisioners for dependency installation
- [ ] Create Terraform variables that map to GitHub secrets/variables
- [ ] Update Cloudflare DNS to point to EC2 instance IP (manual one-time step)
- [ ] Set up Terraform outputs for service status and URLs

### 4. Operations & Quality
- [ ] Add error handling, logging, and monitoring (systemd journal, nginx logs)
- [ ] Write unit tests for handlers and services (go test)
- [ ] Add benchmark tests for performance validation
- [ ] Implement race condition testing (go test -race)
- [ ] Create documentation for API usage and deployment
- [ ] Test end-to-end functionality and perform load testing

## Technical Requirements

### HTTP Endpoints
- `POST /shorten` - Create short links (htmx form submission)
- `GET /{code}` - Redirect to original URL (with analytics)
- `GET /dashboard` - Admin dashboard (htmx-powered)
- `GET /analytics` - Get analytics data (htmx partial updates)
- `GET /health` - Health check endpoint for monitoring

### Infrastructure
- **EC2**: Existing t2.large instance for application hosting
- **SQLite**: Local database file for URL mappings and analytics
- **Nginx**: Reverse proxy for domain routing to application port
- **Terraform**: Infrastructure as Code for deployment automation
- **GitHub Actions**: CI/CD pipeline with automated deployment
- **GitHub Secrets**: Secure storage for certificates and sensitive configuration
- **Cloudflare Origin Certificate**: SSL certificate for end-to-end encryption
- **Cloudflare**: DNS management for lnk.avantifellows.org
- **Systemd**: Service management and auto-restart functionality

### Features
- URL shortening with custom/random codes
- Click analytics and tracking
- Simple dashboard UI
- Basic authentication
- Rate limiting and validation

## Success Criteria
- Links redirect correctly with sub-200ms response time
- Dashboard shows accurate click statistics
- Infrastructure deployment is fully automated via GitHub Actions + Terraform
- Nginx reverse proxy correctly routes lnk.avantifellows.org to port 8080
- Cloudflare Origin Certificate is properly installed for end-to-end encryption
- All sensitive configuration is stored in GitHub Secrets (open-source ready)
- Contributors can easily deploy to their own infrastructure
- Service can handle 1000+ concurrent users with <5MB RAM usage
- Clean, maintainable codebase with proper Terraform modules