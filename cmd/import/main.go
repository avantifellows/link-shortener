package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type LinkRecord struct {
	ShortCode   string
	OriginalURL string
	CreatedAt   time.Time
}

func main() {
	if len(os.Args) != 3 {
		log.Fatal("Usage: go run cmd/import/main.go <csv_file> <database_path>")
	}

	csvFile := os.Args[1]
	dbPath := os.Args[2]

	// Open CSV file
	file, err := os.Open(csvFile)
	if err != nil {
		log.Fatalf("Error opening CSV file: %v", err)
	}
	defer file.Close()

	// Create CSV reader
	reader := csv.NewReader(file)
	
	// Read all records
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("Error reading CSV: %v", err)
	}

	if len(records) == 0 {
		log.Fatal("CSV file is empty")
	}

	// Skip header row
	records = records[1:]
	
	// Open database
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	defer db.Close()

	// Create tables if they don't exist
	if err := createTables(db); err != nil {
		log.Fatalf("Error creating tables: %v", err)
	}

	// Process and import records
	imported := 0
	skipped := 0
	duplicates := 0

	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("Error starting transaction: %v", err)
	}
	defer tx.Rollback()

	// Prepare statements
	checkStmt, err := tx.Prepare("SELECT COUNT(*) FROM link_mappings WHERE short_code = ?")
	if err != nil {
		log.Fatalf("Error preparing check statement: %v", err)
	}
	defer checkStmt.Close()

	insertStmt, err := tx.Prepare(`
		INSERT INTO link_mappings (short_code, original_url, created_at, created_by, click_count, last_accessed)
		VALUES (?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		log.Fatalf("Error preparing insert statement: %v", err)
	}
	defer insertStmt.Close()

	for i, record := range records {
		if len(record) < 29 { // Ensure we have enough columns
			log.Printf("Skipping row %d: insufficient columns", i+2)
			skipped++
			continue
		}

		shortLink := record[0]  // short_link column
		originalURL := record[1] // link column
		timestampStr := record[28] // creation_timestamp column

		// Extract short code from URL
		shortCode := extractShortCode(shortLink)
		if shortCode == "" {
			log.Printf("Skipping row %d: invalid short link format: %s", i+2, shortLink)
			skipped++
			continue
		}

		// Parse timestamp
		createdAt, err := parseTimestamp(timestampStr)
		if err != nil {
			log.Printf("Skipping row %d: invalid timestamp %s: %v", i+2, timestampStr, err)
			skipped++
			continue
		}

		// Check for duplicates
		var count int
		err = checkStmt.QueryRow(shortCode).Scan(&count)
		if err != nil {
			log.Printf("Error checking duplicate for %s: %v", shortCode, err)
			skipped++
			continue
		}

		if count > 0 {
			log.Printf("Skipping duplicate short code: %s", shortCode)
			duplicates++
			continue
		}

		// Insert record
		_, err = insertStmt.Exec(
			shortCode,
			originalURL,
			createdAt.Unix(),
			"imported",
			0,    // click_count
			nil,  // last_accessed
		)
		if err != nil {
			log.Printf("Error inserting record %d (code: %s): %v", i+2, shortCode, err)
			skipped++
			continue
		}

		imported++
		if imported%1000 == 0 {
			log.Printf("Imported %d records...", imported)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		log.Fatalf("Error committing transaction: %v", err)
	}

	log.Printf("Import completed!")
	log.Printf("  Total records processed: %d", len(records))
	log.Printf("  Successfully imported: %d", imported)
	log.Printf("  Duplicates found: %d", duplicates)
	log.Printf("  Skipped (errors): %d", skipped)
}

func extractShortCode(shortLink string) string {
	// Extract code from https://lnk.avantifellows.org/CODE
	prefix := "https://lnk.avantifellows.org/"
	if !strings.HasPrefix(shortLink, prefix) {
		return ""
	}
	return strings.TrimPrefix(shortLink, prefix)
}

func parseTimestamp(timestampStr string) (time.Time, error) {
	// Parse ISO 8601 timestamp like "2023-12-08T05:25:08.000Z"
	layouts := []string{
		"2006-01-02T15:04:05.000Z",
		"2006-01-02T15:04:05Z",
		time.RFC3339,
	}
	
	for _, layout := range layouts {
		if t, err := time.Parse(layout, timestampStr); err == nil {
			return t, nil
		}
	}
	
	return time.Time{}, fmt.Errorf("unable to parse timestamp: %s", timestampStr)
}

func createTables(db *sql.DB) error {
	createLinkMappings := `
	CREATE TABLE IF NOT EXISTS link_mappings (
		short_code TEXT PRIMARY KEY,
		original_url TEXT NOT NULL,
		created_at INTEGER NOT NULL,
		created_by TEXT,
		click_count INTEGER DEFAULT 0,
		last_accessed INTEGER
	);`

	createClickAnalytics := `
	CREATE TABLE IF NOT EXISTS click_analytics (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		short_code TEXT NOT NULL,
		timestamp INTEGER NOT NULL,
		user_agent TEXT,
		ip_address TEXT,
		referrer TEXT,
		FOREIGN KEY (short_code) REFERENCES link_mappings(short_code)
	);`

	createIndexes := `
	CREATE INDEX IF NOT EXISTS idx_created_at ON link_mappings(created_at);
	CREATE INDEX IF NOT EXISTS idx_click_count ON link_mappings(click_count);
	CREATE INDEX IF NOT EXISTS idx_short_code_timestamp ON click_analytics(short_code, timestamp);
	`

	queries := []string{createLinkMappings, createClickAnalytics, createIndexes}
	
	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("error executing query: %v", err)
		}
	}
	
	return nil
}