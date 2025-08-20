# Link Shortener

A minimalist link shortener service for Avanti Fellows, built with Go, htmx, and SQLite.

## Features

- **URL Shortening**: Create short URLs from long ones with optional custom codes
- **Fast Redirects**: High-performance redirects using Go's compiled binary
- **Analytics Dashboard**: Track clicks and view link statistics with public access
- **Real-time Updates**: htmx-powered interface with auto-refresh
- **Bearer Token Authentication**: Secure API access for link creation
- **SQLite Storage**: Lightweight database for URL mappings and analytics
- **Responsive UI**: Tailwind CSS for modern, mobile-friendly design

## Quick Start

1. **Clone and run locally**:
   ```bash
   git clone <repo-url>
   cd link_shortening
   go mod tidy
   
   # Copy environment template and configure
   cp .env.example .env.local
   # Edit .env.local and set your AUTH_TOKEN (or use the default)
   
   go run cmd/server/main.go
   ```

2. **Open your browser** to `http://localhost:8080`

3. **View analytics** and existing links on the public dashboard

4. **Create a short link** via [API](API.md) (requires authentication) or browser form

5. **Test the redirect** by visiting the generated short URL

## API Documentation

For complete API documentation including endpoints, authentication, examples, and SDKs, see **[API.md](API.md)**

## Project Structure

```
link_shortening/
├── cmd/server/main.go           # Application entry point
├── internal/
│   ├── handlers/handlers.go     # HTTP request handlers
│   ├── middleware/auth.go       # Bearer token authentication
│   ├── models/link.go          # Data structures
│   ├── database/database.go    # SQLite setup and migrations
│   └── services/shortener.go   # Business logic
├── templates/                  # HTML templates with htmx
│   ├── base.html
│   ├── dashboard.html
│   ├── analytics-table.html
│   └── success-message.html
├── .env.example                # Environment template
├── .env.local                  # Local environment (gitignored)
├── test_api.sh                 # Comprehensive API test suite
├── go.mod                      # Go dependencies
└── README.md
```

## Database Schema

### link_mappings
- `short_code` (TEXT, PRIMARY KEY) - The shortened code
- `original_url` (TEXT) - Original long URL
- `created_at` (INTEGER) - Unix timestamp
- `created_by` (TEXT) - Creator identifier
- `click_count` (INTEGER) - Number of clicks
- `last_accessed` (INTEGER) - Last click timestamp

### click_analytics
- `id` (INTEGER, AUTOINCREMENT) - Unique click ID
- `short_code` (TEXT) - Reference to link
- `timestamp` (INTEGER) - Click timestamp
- `user_agent` (TEXT) - Browser user agent
- `ip_address` (TEXT) - Client IP address
- `referrer` (TEXT) - HTTP referrer header

## Environment Variables

Create `.env.local` from `.env.example` and configure:

- `AUTH_TOKEN` - Bearer token for API authentication (UUID format)
- `PORT` - Server port (default: 8080)
- `DATABASE_PATH` - SQLite database path (default: link_shortener.db)
- `BASE_URL` - Base URL for short links (default: http://localhost:8080)
- `DEBUG` - Enable debug logging (default: false)
- `LOG_LEVEL` - Logging level (default: INFO)

## Dependencies

- **Go 1.21+** - Programming language
- **Chi Router** - HTTP router and middleware
- **modernc.org/sqlite** - Pure Go SQLite driver
- **godotenv** - Environment variable management
- **htmx** - Frontend interactivity (via CDN)
- **Tailwind CSS** - Styling (via CDN)

## Testing

### Comprehensive Test Suite
Run the complete API test suite:
```bash
# Start server
go run cmd/server/main.go

# Run all tests
./test_api.sh
```

### Manual Testing
```bash
# Set authentication token (get from your .env.local)
export AUTH_TOKEN="your-uuid-token-here"

# Create short URL (requires auth)
curl -X POST http://localhost:8080/shorten \
  -H "Authorization: Bearer $AUTH_TOKEN" \
  -H "Accept: application/json" \
  -d "original_url=https://www.avantifellows.org&created_by=test"

# Test redirect (public, no auth needed)
curl -L http://localhost:8080/{short_code}

# Check analytics (public, no auth needed)
curl -H "Accept: application/json" http://localhost:8080/analytics
```

### Test Coverage
The test suite validates:
- ✅ Public endpoints (dashboard, analytics, redirects, health)
- ✅ Protected endpoints (link creation with bearer token)
- ✅ Authentication rejection (invalid/missing tokens)
- ✅ JSON API responses
- ✅ Error handling (404s, malformed requests)

## Authentication Setup

### Current Authentication Model
- 🌐 **Dashboard**: Public access (browser-friendly)
- 🌐 **Analytics**: Public access (for monitoring) 
- 🔒 **Link Creation**: Bearer token required (API security)
- 🌐 **Redirects**: Public access (end-user friendly)

### For Production Deployment
Add the `AUTH_TOKEN` to your deployment environment:
- **GitHub Secrets**: `AUTH_TOKEN=your-generated-uuid-token`
- **Docker**: `-e AUTH_TOKEN=your-generated-uuid-token`
- **Systemd**: Add to environment file

Generate a secure UUID token:
```bash
# On macOS/Linux
uuidgen

# Or online UUID generator
# Use this token in your .env.local and deployment secrets
```

## Next Steps (Deployment)

- [x] ✅ Bearer token authentication
- [x] ✅ Public dashboard access
- [x] ✅ API security for link creation
- [ ] Rate limiting and security middleware
- [ ] Docker containerization  
- [ ] Production deployment with Terraform
- [ ] SSL certificate setup
- [ ] Custom domain configuration (lnk.avantifellows.org)
- [ ] Monitoring and logging

## License

Internal use for Avanti Fellows.