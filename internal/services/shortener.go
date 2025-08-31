package services

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/avantifellows/link-shortener/internal/models"
)

type ShortenerService struct {
	db *sql.DB
}

func NewShortenerService(db *sql.DB) *ShortenerService {
	return &ShortenerService{db: db}
}

func (s *ShortenerService) CreateShortURL(req models.CreateShortURLRequest) (*models.CreateShortURLResponse, error) {
	// Validate URL
	if !isValidURL(req.OriginalURL) {
		return nil, fmt.Errorf("invalid URL format")
	}

	var shortCode string
	var err error

	// Use custom code if provided and available
	if req.CustomCode != "" {
		if !isValidShortCode(req.CustomCode) {
			return nil, fmt.Errorf("invalid custom code format")
		}
		
		shortCode = req.CustomCode
		// For custom codes, we'll handle conflicts in the database insert
	} else {
		// For generated codes, use retry logic with database insert
		shortCode, err = s.generateUniqueShortCode(req.OriginalURL, req.CreatedBy)
		if err != nil {
			return nil, fmt.Errorf("failed to create short code: %w", err)
		}
		
		// Code was successfully inserted, return response
		baseURL := getBaseURL()
		return &models.CreateShortURLResponse{
			ShortCode:   shortCode,
			ShortURL:    fmt.Sprintf("%s/%s", baseURL, shortCode),
			OriginalURL: req.OriginalURL,
		}, nil
	}

	// Handle custom code insertion (with potential conflict)
	_, err = s.db.Exec(`
		INSERT INTO link_mappings (short_code, original_url, created_at, created_by, click_count)
		VALUES (?, ?, ?, ?, 0)
	`, shortCode, req.OriginalURL, time.Now().Unix(), req.CreatedBy)

	if err != nil {
		// Check if this is a constraint violation (code already exists)
		if strings.Contains(err.Error(), "UNIQUE constraint failed") || strings.Contains(err.Error(), "PRIMARY KEY constraint failed") {
			return nil, fmt.Errorf("custom code already exists")
		}
		return nil, fmt.Errorf("failed to store URL mapping: %w", err)
	}

	baseURL := getBaseURL()
	return &models.CreateShortURLResponse{
		ShortCode:   shortCode,
		ShortURL:    fmt.Sprintf("%s/%s", baseURL, shortCode),
		OriginalURL: req.OriginalURL,
	}, nil
}

func (s *ShortenerService) GetOriginalURL(shortCode string) (string, error) {
	var originalURL string
	err := s.db.QueryRow(`
		SELECT original_url FROM link_mappings WHERE short_code = ?
	`, shortCode).Scan(&originalURL)

	if err == sql.ErrNoRows {
		return "", fmt.Errorf("short code not found")
	}
	if err != nil {
		return "", fmt.Errorf("database error: %w", err)
	}

	return originalURL, nil
}

func (s *ShortenerService) TrackClick(shortCode, userAgent, ipAddress, referrer string) error {
	// Record click analytics
	_, err := s.db.Exec(`
		INSERT INTO click_analytics (short_code, timestamp, user_agent, ip_address, referrer)
		VALUES (?, ?, ?, ?, ?)
	`, shortCode, time.Now().Unix(), userAgent, ipAddress, referrer)

	if err != nil {
		return fmt.Errorf("failed to record click analytics: %w", err)
	}

	// Update click count and last accessed
	_, err = s.db.Exec(`
		UPDATE link_mappings 
		SET click_count = click_count + 1, last_accessed = ?
		WHERE short_code = ?
	`, time.Now().Unix(), shortCode)

	if err != nil {
		return fmt.Errorf("failed to update click count: %w", err)
	}

	return nil
}

func (s *ShortenerService) BeginTransaction() (*sql.Tx, error) {
	return s.db.Begin()
}

func (s *ShortenerService) TrackClickInTransaction(tx *sql.Tx, shortCode, userAgent, ipAddress, referrer string, timestamp time.Time) error {
	// Record click analytics
	_, err := tx.Exec(`
		INSERT INTO click_analytics (short_code, timestamp, user_agent, ip_address, referrer)
		VALUES (?, ?, ?, ?, ?)
	`, shortCode, timestamp.Unix(), userAgent, ipAddress, referrer)

	if err != nil {
		return fmt.Errorf("failed to record click analytics: %w", err)
	}

	// Update click count and last accessed
	_, err = tx.Exec(`
		UPDATE link_mappings 
		SET click_count = click_count + 1, last_accessed = ?
		WHERE short_code = ?
	`, timestamp.Unix(), shortCode)

	if err != nil {
		return fmt.Errorf("failed to update click count: %w", err)
	}

	return nil
}

func (s *ShortenerService) GetAnalytics() (*models.AnalyticsResponse, error) {
	return s.GetAnalyticsPaginated(1, 50, "")
}

func (s *ShortenerService) GetAnalyticsPaginated(page, pageSize int, searchTerm string) (*models.AnalyticsResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 1000 {
		pageSize = 50
	}

	// Build WHERE clause for search
	var whereClause string
	var queryArgs []interface{}
	var countArgs []interface{}
	
	if searchTerm != "" {
		searchPattern := "%" + searchTerm + "%"
		whereClause = "WHERE (short_code LIKE ? OR original_url LIKE ?)"
		queryArgs = []interface{}{searchPattern, searchPattern}
		countArgs = []interface{}{searchPattern, searchPattern}
	}

	// Get total count first
	var totalLinks int
	var totalClicks int
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*), COALESCE(SUM(click_count), 0)
		FROM link_mappings %s
	`, whereClause)
	
	err := s.db.QueryRow(countQuery, countArgs...).Scan(&totalLinks, &totalClicks)
	if err != nil {
		return nil, fmt.Errorf("failed to get totals: %w", err)
	}

	// Calculate pagination
	offset := (page - 1) * pageSize
	totalPages := (totalLinks + pageSize - 1) / pageSize

	// Get paginated links
	linkQuery := fmt.Sprintf(`
		SELECT short_code, original_url, created_at, created_by, click_count, last_accessed
		FROM link_mappings %s
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, whereClause)
	
	// Add LIMIT and OFFSET to query args
	queryArgs = append(queryArgs, pageSize, offset)
	
	rows, err := s.db.Query(linkQuery, queryArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch links: %w", err)
	}
	defer rows.Close()

	var links []models.LinkMapping

	for rows.Next() {
		var link models.LinkMapping
		var createdAt int64
		var lastAccessed sql.NullInt64

		err := rows.Scan(&link.ShortCode, &link.OriginalURL, &createdAt, &link.CreatedBy, &link.ClickCount, &lastAccessed)
		if err != nil {
			return nil, fmt.Errorf("failed to scan link: %w", err)
		}

		link.CreatedAt = time.Unix(createdAt, 0)
		if lastAccessed.Valid {
			t := time.Unix(lastAccessed.Int64, 0)
			link.LastAccessed = &t
		}

		links = append(links, link)
	}

	// Get recent clicks
	clickRows, err := s.db.Query(`
		SELECT id, short_code, timestamp, user_agent, ip_address, referrer
		FROM click_analytics 
		ORDER BY timestamp DESC 
		LIMIT 50
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch recent clicks: %w", err)
	}
	defer clickRows.Close()

	var recentClicks []models.ClickAnalytics
	for clickRows.Next() {
		var click models.ClickAnalytics
		var timestamp int64

		err := clickRows.Scan(&click.ID, &click.ShortCode, &timestamp, &click.UserAgent, &click.IPAddress, &click.Referrer)
		if err != nil {
			return nil, fmt.Errorf("failed to scan click: %w", err)
		}

		click.Timestamp = time.Unix(timestamp, 0)
		recentClicks = append(recentClicks, click)
	}

	return &models.AnalyticsResponse{
		Links:        links,
		TotalLinks:   totalLinks,
		TotalClicks:  totalClicks,
		RecentClicks: recentClicks,
		Pagination: &models.Pagination{
			CurrentPage: page,
			TotalPages:  totalPages,
			PageSize:    pageSize,
			TotalItems:  totalLinks,
			HasNext:     page < totalPages,
			HasPrev:     page > 1,
		},
	}, nil
}

func (s *ShortenerService) generateUniqueShortCode(originalURL, createdBy string) (string, error) {
	const maxAttempts = 10
	
	for i := 0; i < maxAttempts; i++ {
		// Generate 4 random bytes for 4-character code (like Firebase Dynamic Links)
		bytes := make([]byte, 3)
		if _, err := rand.Read(bytes); err != nil {
			return "", err
		}
		
		// Use URL-safe base64 encoding and remove padding
		code := strings.TrimRight(base64.URLEncoding.EncodeToString(bytes), "=")
		
		// Take exactly 4 characters (base64 of 3 bytes = 4 chars)
		if len(code) >= 4 {
			code = code[:4]
		}
		
		// Attempt to insert directly into database - this is atomic
		_, err := s.db.Exec(`
			INSERT INTO link_mappings (short_code, original_url, created_at, created_by, click_count)
			VALUES (?, ?, ?, ?, 0)
		`, code, originalURL, time.Now().Unix(), createdBy)
		
		if err == nil {
			// Success! Code was unique and inserted
			return code, nil
		}
		
		// If it's a constraint violation, try again with a new code
		if strings.Contains(err.Error(), "UNIQUE constraint failed") || strings.Contains(err.Error(), "PRIMARY KEY constraint failed") {
			continue
		}
		
		// Other database error, return it
		return "", fmt.Errorf("database error: %w", err)
	}
	
	return "", fmt.Errorf("failed to generate unique short code after %d attempts", maxAttempts)
}

func (s *ShortenerService) shortCodeExists(code string) bool {
	var exists bool
	err := s.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM link_mappings WHERE short_code = ?)`, code).Scan(&exists)
	return err == nil && exists
}

func isValidURL(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func isValidShortCode(code string) bool {
	if len(code) < 3 || len(code) > 20 {
		return false
	}
	
	// Allow alphanumeric characters, hyphens, and underscores
	for _, r := range code {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_') {
			return false
		}
	}
	
	return true
}

func getBaseURL() string {
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	return baseURL
}