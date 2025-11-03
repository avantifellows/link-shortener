# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A minimalist link shortener service for Avanti Fellows (~200 internal users) built with Go, htmx, and SQLite. The service provides URL shortening with analytics, featuring a hybrid authentication model: public dashboard/redirects and authenticated API endpoints.

## Tech Stack & Architecture

### Backend Stack
- **Go 1.24+** with Chi Router for HTTP routing and middleware
- **SQLite** with modernc.org/sqlite (pure Go driver) for data persistence
- **html/template** for server-side rendering with custom template functions
- **godotenv** for environment variable management

### Frontend Stack
- **Server-side rendered HTML** with Go templates
- **htmx** (via CDN) for dynamic interactions and real-time updates
- **Tailwind CSS** (via CDN) for styling
- **Minimal JavaScript** - htmx handles most dynamic behavior

### Key Architecture Patterns
- **Hybrid Authentication**: Public dashboard/redirects, Bearer token protected API
- **Async Click Tracking**: Background goroutine processes click events via channels
- **Template Function Extensions**: Custom template functions for pagination and calculations
- **Graceful Shutdown**: 30-second timeout with signal handling

## Project Structure & Components

```
internal/
├── handlers/handlers.go     # HTTP handlers with template rendering and async click tracking
├── services/shortener.go    # Core business logic (URL validation, code generation)
├── database/database.go     # SQLite setup, migrations, connection pooling
├── models/link.go          # Data structures for LinkMapping and ClickAnalytics
├── middleware/auth.go      # Bearer token authentication middleware
└── logger/logger.go        # Structured logging utilities

cmd/
├── server/main.go          # Application entry point with Chi router setup
└── import/main.go          # Data import utility (if exists)

templates/                   # Go html/templates with htmx integration
├── base.html               # Base template with common layout
├── dashboard.html          # Main interface with search and pagination
├── analytics-table.html    # htmx partial for analytics data
└── success-message.html    # htmx success notifications
```

## Development Commands

### Local Development Setup
```bash
# Environment setup
cp .env.example .env.local
# Edit .env.local with your AUTH_TOKEN (generate with: uuidgen)

# Install dependencies and run
go mod tidy
go run cmd/server/main.go

# Alternative quick start
make dev  # if Makefile exists
```

### Testing
```bash
# Comprehensive API test suite (requires running server)
./test_api.sh

# Manual API testing
export AUTH_TOKEN="your-uuid-token"
curl -X POST http://localhost:8080/shorten \
  -H "Authorization: Bearer $AUTH_TOKEN" \
  -H "Accept: application/json" \
  -d "original_url=https://example.com&created_by=test"

# Go testing (when test files exist)
go test ./...
go test -race ./...
go test -bench=. ./...
```

### Building & Deployment
```bash
# Local build
go build -o link-shortener cmd/server/main.go

# Production build (optimized)
CGO_ENABLED=0 go build -ldflags="-s -w" -o link-shortener cmd/server/main.go

# Deploy to server (uses .env.deploy.local config)
./deploy.sh
```

## Database Schema & Models

### Core Tables
- **link_mappings**: Primary URL storage with short_code (PK), original_url, click tracking
- **click_analytics**: Detailed click tracking with user_agent, ip_address, referrer

### Important Go Models
- **LinkMapping**: Core struct with db and json tags, handles nullable LastAccessed
- **ClickAnalytics**: Click tracking with foreign key to short_code
- **CreateShortURLRequest/Response**: API request/response structures with form and json tags

## API Architecture & Routes

### Authentication Model
- **Public Routes**: `/`, `/{code}`, `/health` - No auth required
- **Protected Routes**: `/api/*` and legacy endpoints - Bearer token required
- **Hybrid Access**: Dashboard public for viewing, API protected for creation/modification

### Route Structure
```go
// Public routes
r.Get("/health", h.Health)
r.Get("/", h.Dashboard)
r.Get("/{code}", h.RedirectURL)

// Authenticated API routes (/api/*)
r.Route("/api", func(r chi.Router) {
    r.Use(authmiddleware.AuthMiddleware)
    r.Get("/links", h.Analytics)
    r.Post("/links", h.CreateShortURL)
    r.Get("/links/{code}", h.GetLink)
    r.Put("/links/{code}", h.UpdateLink)
    r.Delete("/links/{code}", h.DeleteLink)
})

// Legacy routes (backward compatibility)
r.Post("/shorten", h.CreateShortURL)  // Protected
r.Get("/analytics", h.Analytics)       // Protected
```

## Environment Configuration

### Required Variables (.env.local)
```bash
AUTH_TOKEN=your-uuid-token-here  # Bearer token for API authentication
DATABASE_PATH=./link_shortener.db
PORT=8080
BASE_URL=http://localhost:8080
DEBUG=true
LOG_LEVEL=INFO
```

### Production Deployment Variables
Additional variables in deployment configs for server details, SSL certificates, and service management.

## Development Patterns

### Template System
- **Custom Functions**: `divf`, `add`, `sub`, `pagination_range` for calculations
- **htmx Integration**: Partial template updates for real-time analytics
- **Form Handling**: Supports both form and JSON content types

### Async Click Tracking
```go
// Handlers use background channel for non-blocking click recording
clickQueue := make(chan ClickEvent, 1000)
go h.processClickQueue()  // Background goroutine processes clicks
```

### Database Patterns
- **Connection Pooling**: Configured for SQLite with proper timeout settings
- **Migration on Startup**: Schema creation with IF NOT EXISTS patterns
- **Prepared Statements**: Used throughout for performance and security

### Error Handling
- **Structured Logging**: Custom logger package for consistent output
- **Graceful Degradation**: Click tracking failures don't affect redirects
- **Content Negotiation**: Returns JSON or HTML based on Accept header

## Testing Strategy

### API Test Suite (test_api.sh)
Validates all endpoints including:
- Public access (dashboard, redirects, health)
- Authentication (valid/invalid tokens)
- JSON/HTML content negotiation
- Error scenarios (404s, malformed requests)

### Manual Testing Patterns
- Use environment AUTH_TOKEN for authenticated requests
- Test both legacy and new API endpoints
- Verify htmx partial updates work correctly

## Common Development Tasks

### Adding New Endpoints
1. Add route in cmd/server/main.go
2. Implement handler in internal/handlers/handlers.go
3. Update models if needed
4. Add tests to test_api.sh
5. Update templates if UI changes needed

### Database Changes
1. Update schema in internal/database/database.go
2. Add new models to internal/models/link.go
3. Test migrations work with existing data
4. Update handlers and services as needed

### Template Updates
1. Edit templates/*.html files
2. Test htmx functionality locally
3. Ensure responsive design with Tailwind classes
4. Verify both desktop and mobile layouts