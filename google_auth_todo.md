# Google Auth Implementation TODO

Implementation plan for adding Google OAuth to the web UI while keeping Bearer token authentication for API endpoints.

## Phase 1: Dependencies & Configuration

- [ ] Add OAuth2 dependencies to go.mod
  ```bash
  go get golang.org/x/oauth2
  go get google.golang.org/api/oauth2/v2
  ```

- [ ] Add Google OAuth environment variables to `.env.example`
  ```bash
  GOOGLE_CLIENT_ID=your-google-client-id
  GOOGLE_CLIENT_SECRET=your-google-client-secret
  GOOGLE_REDIRECT_URL=http://localhost:8080/auth/callback
  ```

- [ ] Create Google OAuth application in Google Cloud Console
  - Enable Google+ API or Google Identity API
  - Set authorized redirect URIs
  - Get client ID and secret

## Phase 2: Authentication Infrastructure

- [ ] Create `internal/auth/google.go`
  - OAuth2 config setup
  - State parameter generation for CSRF protection
  - User info extraction from Google response

- [ ] Create `internal/middleware/google_auth.go`
  - Session validation middleware for UI routes
  - Redirect to login if not authenticated
  - Extract user info and add to request context

- [ ] Add session management
  - Choose: JWT tokens vs server-side sessions
  - Implement session storage and validation
  - Handle session expiration

## Phase 3: Route Updates

- [ ] Add new authentication routes in `cmd/server/main.go`
  ```go
  r.Get("/auth/google", h.GoogleLogin)
  r.Get("/auth/callback", h.GoogleCallback)
  r.Post("/auth/logout", h.Logout)
  ```

- [ ] Protect UI routes with Google auth middleware
  ```go
  r.Group(func(r chi.Router) {
      r.Use(googleAuthMiddleware)
      r.Get("/", h.Dashboard)
      // Other UI routes that need protection
  })
  ```

- [ ] Keep API routes unchanged (Bearer token auth)
  - Verify `/api/*` routes still use Bearer token
  - Verify legacy routes still work for API clients

## Phase 4: Handler Implementation

- [ ] Implement `GoogleLogin` handler
  - Generate OAuth URL with state parameter
  - Redirect user to Google auth

- [ ] Implement `GoogleCallback` handler
  - Validate state parameter
  - Exchange code for token
  - Get user info from Google
  - Create session/JWT
  - Redirect to dashboard

- [ ] Implement `Logout` handler
  - Clear session/JWT
  - Redirect to login page

- [ ] Update existing handlers to use Google user info
  - Extract user email from session for `created_by` field
  - Pass user info to templates

## Phase 5: Template Updates

- [ ] Update `templates/base.html`
  - Add user info display (name, email)
  - Add login/logout buttons
  - Style user interface elements

- [ ] Update `templates/dashboard.html`
  - Show user-specific messaging
  - Update form submission to use authenticated user

- [ ] Create login page template (optional)
  - Simple page with "Sign in with Google" button
  - Or handle redirect directly from protected routes

## Phase 6: Testing & Validation

- [ ] Update `test_api.sh`
  - Verify API endpoints still work with Bearer token
  - Add note about UI requiring browser testing

- [ ] Manual testing checklist
  - [ ] Google login flow works
  - [ ] User info displays correctly
  - [ ] Form submissions include user email as `created_by`
  - [ ] Logout clears session properly
  - [ ] API endpoints unchanged (Bearer token still works)
  - [ ] Redirects still public
  - [ ] Protected UI routes redirect to login

- [ ] Test error scenarios
  - [ ] Invalid OAuth state parameter
  - [ ] Google API errors
  - [ ] Session expiration handling
  - [ ] Non-Avanti Fellows email (if restricting domain)

## Phase 7: Documentation & Deployment

- [ ] Update `README.md`
  - Add Google OAuth setup instructions
  - Update environment variable documentation
  - Add deployment notes for Google credentials

- [ ] Update `CLAUDE.md`
  - Document new authentication patterns
  - Update route structure documentation
  - Add Google OAuth troubleshooting

- [ ] Production deployment considerations
  - [ ] Add Google OAuth secrets to deployment environment
  - [ ] Update redirect URLs for production domain
  - [ ] Test OAuth flow on production domain

## Technical Decisions to Make

- [ ] **Session Management**: JWT tokens vs server-side sessions?
  - JWT: Stateless, simpler deployment
  - Sessions: More secure, easier to revoke

- [ ] **Domain Restriction**: Limit to @avantifellows.org emails?
  - Check `hd` parameter in Google OAuth response
  - Reject non-Avanti Fellows users

- [ ] **Session Duration**: How long should sessions last?
  - Consider user experience vs security
  - Implement refresh logic if needed

- [ ] **Error Handling**: How to handle OAuth failures?
  - Graceful error pages
  - Logging for debugging

## Rollback Plan

- [ ] Document rollback steps
  - Remove Google auth middleware from UI routes
  - Revert to public dashboard access
  - Keep all API functionality unchanged

## Estimated Timeline

- **Phase 1-2**: 1 day (setup and infrastructure)
- **Phase 3-4**: 1 day (routes and handlers)
- **Phase 5**: 0.5 days (templates)
- **Phase 6-7**: 0.5 days (testing and docs)

**Total: ~3 days** with buffer for testing and refinement.