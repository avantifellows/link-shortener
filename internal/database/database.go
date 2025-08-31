package database

import (
	"database/sql"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

const schema = `
CREATE TABLE IF NOT EXISTS link_mappings (
    short_code TEXT PRIMARY KEY,
    original_url TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    created_by TEXT,
    click_count INTEGER DEFAULT 0,
    last_accessed INTEGER
);

CREATE INDEX IF NOT EXISTS idx_created_at ON link_mappings(created_at);
CREATE INDEX IF NOT EXISTS idx_click_count ON link_mappings(click_count);

CREATE TABLE IF NOT EXISTS click_analytics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    short_code TEXT NOT NULL,
    timestamp INTEGER NOT NULL,
    user_agent TEXT,
    ip_address TEXT,
    referrer TEXT,
    FOREIGN KEY (short_code) REFERENCES link_mappings(short_code)
);

CREATE INDEX IF NOT EXISTS idx_short_code_timestamp ON click_analytics(short_code, timestamp);
`

func Initialize() (*sql.DB, error) {
	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "link_shortener.db"
	}

	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	// Add SQLite connection parameters for better concurrency
	connectionString := dbPath + "?_journal_mode=WAL&_synchronous=NORMAL&_cache_size=1000&_foreign_keys=1&_busy_timeout=5000"
	
	db, err := sql.Open("sqlite", connectionString)
	if err != nil {
		return nil, err
	}

	// Configure connection pool for high concurrency
	db.SetMaxOpenConns(25)    // Limit concurrent connections to prevent resource exhaustion
	db.SetMaxIdleConns(25)    // Keep connections alive for reuse
	db.SetConnMaxLifetime(5 * time.Minute) // Rotate connections periodically

	if err := db.Ping(); err != nil {
		return nil, err
	}

	// Create schema
	if _, err := db.Exec(schema); err != nil {
		return nil, err
	}

	return db, nil
}