package router

import (
	"net/http"
	"time"

	"github.com/avantifellows/link-shortener/internal/handlers"
	authmiddleware "github.com/avantifellows/link-shortener/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// New builds the application's HTTP router. It is shared by the server entrypoint
// (cmd/server) and the handler tests so route wiring stays in a single place.
func New(h *handlers.Handlers) http.Handler {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(middleware.Timeout(60 * time.Second))

	// Public routes (no authentication required)
	r.Get("/health", h.Health)
	r.Get("/", h.Dashboard)         // Dashboard web interface
	r.Get("/{code}", h.RedirectURL) // URL redirects - MUST be last to avoid conflicts

	// API routes (with authentication)
	r.Route("/api", func(r chi.Router) {
		r.Use(authmiddleware.AuthMiddleware)

		// Links API
		r.Get("/links", h.Analytics)            // List all links (GET /api/links)
		r.Post("/links", h.CreateShortURL)      // Create link (POST /api/links)
		r.Get("/links/{code}", h.GetLink)       // Get link details (GET /api/links/{code})
		r.Put("/links/{code}", h.UpdateLink)    // Update link (PUT /api/links/{code})
		r.Delete("/links/{code}", h.DeleteLink) // Delete link (DELETE /api/links/{code})
	})

	// Legacy routes for backward compatibility (can be removed later)
	r.Group(func(r chi.Router) {
		r.Use(authmiddleware.AuthMiddleware)
		r.Post("/shorten", h.CreateShortURL)  // Legacy: redirect to /api/links
		r.Get("/analytics", h.Analytics)      // Legacy: redirect to /api/links
		r.Put("/edit/{code}", h.UpdateLink)   // Legacy: redirect to /api/links/{code}
		r.Patch("/edit/{code}", h.UpdateLink) // Legacy: redirect to /api/links/{code}
	})

	// Serve static files
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	return r
}
