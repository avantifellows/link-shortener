package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/avantifellows/link-shortener/internal/models"
	"github.com/avantifellows/link-shortener/internal/services"
	"github.com/go-chi/chi/v5"
)

type Handlers struct {
	shortenerService *services.ShortenerService
	templates        *template.Template
}

func New(db *sql.DB) *Handlers {
	// Create template functions
	funcMap := template.FuncMap{
		"divf": func(a, b int) float64 {
			if b == 0 {
				return 0
			}
			return float64(a) / float64(b)
		},
		"add": func(a, b int) int {
			return a + b
		},
		"sub": func(a, b int) int {
			return a - b
		},
		"pagination_range": func(current, total int) []int {
			var pages []int
			start := current - 2
			end := current + 2
			
			if start < 1 {
				start = 1
			}
			if end > total {
				end = total
			}
			
			for i := start; i <= end; i++ {
				pages = append(pages, i)
			}
			return pages
		},
		"build_url": func(page int, pageSize int, searchTerm string) string {
			params := fmt.Sprintf("page=%d&size=%d", page, pageSize)
			if searchTerm != "" {
				params += fmt.Sprintf("&search=%s", strings.ReplaceAll(searchTerm, " ", "+"))
			}
			return "/?" + params
		},
	}
	
	// Load templates with functions
	templates := template.Must(template.New("").Funcs(funcMap).ParseGlob("templates/*.html"))
	
	return &Handlers{
		shortenerService: services.NewShortenerService(db),
		templates:        templates,
	}
}

func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

func (h *Handlers) Dashboard(w http.ResponseWriter, r *http.Request) {
	// Get pagination and search parameters
	page := getIntParam(r, "page", 1)
	pageSize := getIntParam(r, "size", 50)
	searchTerm := strings.TrimSpace(r.URL.Query().Get("search"))

	analytics, err := h.shortenerService.GetAnalyticsPaginated(page, pageSize, searchTerm)
	if err != nil {
		log.Printf("Error getting analytics: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data := struct {
		Title      string
		Analytics  *models.AnalyticsResponse
		BaseURL    string
		AuthToken  string
		SearchTerm string
	}{
		Title:      "Link Shortener Dashboard",
		Analytics:  analytics,
		BaseURL:    getBaseURL(),
		AuthToken:  getAuthToken(),
		SearchTerm: searchTerm,
	}

	w.Header().Set("Content-Type", "text/html")
	if err := h.templates.ExecuteTemplate(w, "dashboard.html", data); err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
}

func (h *Handlers) CreateShortURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	req := models.CreateShortURLRequest{
		OriginalURL: strings.TrimSpace(r.FormValue("original_url")),
		CustomCode:  strings.TrimSpace(r.FormValue("custom_code")),
		CreatedBy:   strings.TrimSpace(r.FormValue("created_by")),
	}

	if req.OriginalURL == "" {
		http.Error(w, "Original URL is required", http.StatusBadRequest)
		return
	}

	// Create short URL
	response, err := h.shortenerService.CreateShortURL(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if request accepts JSON (API call) or HTML (htmx/form)
	if strings.Contains(r.Header.Get("Accept"), "application/json") {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	} else {
		// Return htmx partial template with success message
		data := struct {
			Success  bool
			ShortURL string
			Error    string
		}{
			Success:  true,
			ShortURL: response.ShortURL,
		}

		w.Header().Set("Content-Type", "text/html")
		if err := h.templates.ExecuteTemplate(w, "success-message.html", data); err != nil {
			http.Error(w, "Template error", http.StatusInternalServerError)
			return
		}
	}
}

func (h *Handlers) RedirectURL(w http.ResponseWriter, r *http.Request) {
	shortCode := chi.URLParam(r, "code")
	if shortCode == "" {
		http.NotFound(w, r)
		return
	}

	// Get original URL
	originalURL, err := h.shortenerService.GetOriginalURL(shortCode)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Track click analytics
	userAgent := r.Header.Get("User-Agent")
	ipAddress := getClientIP(r)
	referrer := r.Header.Get("Referer")

	// Track click in background (don't block redirect)
	go func() {
		h.shortenerService.TrackClick(shortCode, userAgent, ipAddress, referrer)
	}()

	// Redirect to original URL
	http.Redirect(w, r, originalURL, http.StatusFound)
}

func (h *Handlers) Analytics(w http.ResponseWriter, r *http.Request) {
	// Get pagination and search parameters
	page := getIntParam(r, "page", 1)
	pageSize := getIntParam(r, "size", 50)
	searchTerm := strings.TrimSpace(r.URL.Query().Get("search"))

	analytics, err := h.shortenerService.GetAnalyticsPaginated(page, pageSize, searchTerm)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Check if request wants JSON (API) or HTML (htmx partial)
	if strings.Contains(r.Header.Get("Accept"), "application/json") {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(analytics)
	} else {
		// Return htmx partial with analytics table
		data := struct {
			Analytics  *models.AnalyticsResponse
			BaseURL    string
			SearchTerm string
		}{
			Analytics:  analytics,
			BaseURL:    getBaseURL(),
			SearchTerm: searchTerm,
		}

		w.Header().Set("Content-Type", "text/html")
		if err := h.templates.ExecuteTemplate(w, "analytics-table.html", data); err != nil {
			http.Error(w, "Template error", http.StatusInternalServerError)
			return
		}
	}
}

func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxies)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// Take the first IP if multiple are present
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	return strings.Split(r.RemoteAddr, ":")[0]
}

func getBaseURL() string {
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	return baseURL
}

func getAuthToken() string {
	return os.Getenv("AUTH_TOKEN")
}

func getIntParam(r *http.Request, paramName string, defaultValue int) int {
	param := r.URL.Query().Get(paramName)
	if param == "" {
		return defaultValue
	}
	
	value, err := strconv.Atoi(param)
	if err != nil {
		return defaultValue
	}
	
	return value
}