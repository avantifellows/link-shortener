package models

import (
	"time"
)

type LinkMapping struct {
	ShortCode    string     `json:"short_code" db:"short_code"`
	OriginalURL  string     `json:"original_url" db:"original_url"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	CreatedBy    string     `json:"created_by" db:"created_by"`
	ClickCount   int        `json:"click_count" db:"click_count"`
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

type CreateShortURLRequest struct {
	OriginalURL string `json:"original_url" form:"original_url"`
	CustomCode  string `json:"custom_code" form:"custom_code"`
	CreatedBy   string `json:"created_by" form:"created_by"`
}

type CreateShortURLResponse struct {
	ShortCode   string `json:"short_code"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type AnalyticsResponse struct {
	Links        []LinkMapping    `json:"links"`
	TotalLinks   int              `json:"total_links"`
	TotalClicks  int              `json:"total_clicks"`
	RecentClicks []ClickAnalytics `json:"recent_clicks"`
	Pagination   *Pagination      `json:"pagination,omitempty"`
}

type Pagination struct {
	CurrentPage int `json:"current_page"`
	TotalPages  int `json:"total_pages"`
	PageSize    int `json:"page_size"`
	TotalItems  int `json:"total_items"`
	HasNext     bool `json:"has_next"`
	HasPrev     bool `json:"has_prev"`
}