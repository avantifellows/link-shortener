package services

import (
	"strings"
	"testing"

	"github.com/avantifellows/link-shortener/internal/database"
	"github.com/avantifellows/link-shortener/internal/models"
)

// newTestService spins up a ShortenerService backed by a throwaway SQLite
// database created with the real schema + migrations, so tests exercise the
// same DDL the production binary does.
func newTestService(t *testing.T) *ShortenerService {
	t.Helper()

	dbPath := t.TempDir() + "/test.db"
	t.Setenv("DATABASE_PATH", dbPath)
	t.Setenv("BASE_URL", "http://test.local")

	db, err := database.Initialize()
	if err != nil {
		t.Fatalf("failed to initialize test database: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	return NewShortenerService(db)
}

// TestCreateAndRedirect is the most important test: a URL gets shortened and the
// generated code resolves back to the exact original URL.
func TestCreateAndRedirect(t *testing.T) {
	s := newTestService(t)

	original := "https://example.com/some/long/path?a=1&b=2"
	resp, err := s.CreateShortURL(models.CreateShortURLRequest{
		OriginalURL: original,
		CreatedBy:   "tester",
	})
	if err != nil {
		t.Fatalf("CreateShortURL failed: %v", err)
	}

	if len(resp.ShortCode) != 4 {
		t.Errorf("expected 4-character short code, got %q (len %d)", resp.ShortCode, len(resp.ShortCode))
	}
	if resp.OriginalURL != original {
		t.Errorf("response OriginalURL = %q, want %q", resp.OriginalURL, original)
	}
	wantShortURL := "http://test.local/" + resp.ShortCode
	if resp.ShortURL != wantShortURL {
		t.Errorf("ShortURL = %q, want %q", resp.ShortURL, wantShortURL)
	}

	// The round trip: the code must resolve back to the original URL.
	got, err := s.GetOriginalURL(resp.ShortCode)
	if err != nil {
		t.Fatalf("GetOriginalURL failed: %v", err)
	}
	if got != original {
		t.Errorf("GetOriginalURL = %q, want %q", got, original)
	}
}

func TestCreateWithCustomCode(t *testing.T) {
	s := newTestService(t)

	resp, err := s.CreateShortURL(models.CreateShortURLRequest{
		OriginalURL: "https://example.com",
		CustomCode:  "my-link",
		CreatedBy:   "tester",
	})
	if err != nil {
		t.Fatalf("CreateShortURL with custom code failed: %v", err)
	}
	if resp.ShortCode != "my-link" {
		t.Errorf("ShortCode = %q, want %q", resp.ShortCode, "my-link")
	}

	got, err := s.GetOriginalURL("my-link")
	if err != nil {
		t.Fatalf("GetOriginalURL failed: %v", err)
	}
	if got != "https://example.com" {
		t.Errorf("GetOriginalURL = %q, want %q", got, "https://example.com")
	}
}

func TestCreateDuplicateCustomCodeFails(t *testing.T) {
	s := newTestService(t)

	req := models.CreateShortURLRequest{OriginalURL: "https://example.com", CustomCode: "dupe"}
	if _, err := s.CreateShortURL(req); err != nil {
		t.Fatalf("first create failed: %v", err)
	}

	_, err := s.CreateShortURL(models.CreateShortURLRequest{OriginalURL: "https://other.com", CustomCode: "dupe"})
	if err == nil {
		t.Fatal("expected error creating duplicate custom code, got nil")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error = %q, want it to mention 'already exists'", err.Error())
	}
}

func TestCreateInvalidURL(t *testing.T) {
	s := newTestService(t)

	cases := map[string]string{
		"empty":       "",
		"no scheme":   "example.com/path",
		"scheme only": "https://",
		"garbage":     "not a url at all",
	}
	for name, badURL := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := s.CreateShortURL(models.CreateShortURLRequest{OriginalURL: badURL}); err == nil {
				t.Errorf("expected error for invalid URL %q, got nil", badURL)
			}
		})
	}
}

func TestCreateInvalidCustomCode(t *testing.T) {
	s := newTestService(t)

	cases := map[string]string{
		"too short":     "ab",
		"too long":      strings.Repeat("a", 21),
		"space":         "ab cd",
		"slash":         "ab/cd",
		"special chars": "ab@cd",
	}
	for name, code := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := s.CreateShortURL(models.CreateShortURLRequest{
				OriginalURL: "https://example.com",
				CustomCode:  code,
			})
			if err == nil {
				t.Errorf("expected error for invalid custom code %q, got nil", code)
			}
		})
	}
}

func TestGetOriginalURLNotFound(t *testing.T) {
	s := newTestService(t)

	_, err := s.GetOriginalURL("nope")
	if err == nil {
		t.Fatal("expected error for unknown short code, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error = %q, want it to mention 'not found'", err.Error())
	}
}

// TestGeneratedCodesAreUnique creates many links and verifies every generated
// code is distinct and resolves to the URL it was created with.
func TestGeneratedCodesAreUnique(t *testing.T) {
	s := newTestService(t)

	const n = 200
	seen := make(map[string]string, n)
	for i := 0; i < n; i++ {
		url := "https://example.com/page/" + string(rune('A'+i%26)) + strings.Repeat("x", i)
		resp, err := s.CreateShortURL(models.CreateShortURLRequest{OriginalURL: url})
		if err != nil {
			t.Fatalf("create #%d failed: %v", i, err)
		}
		if prev, ok := seen[resp.ShortCode]; ok {
			t.Fatalf("duplicate short code %q generated (previously for %q)", resp.ShortCode, prev)
		}
		seen[resp.ShortCode] = url

		got, err := s.GetOriginalURL(resp.ShortCode)
		if err != nil || got != url {
			t.Fatalf("round trip failed for %q: got %q, err %v", resp.ShortCode, got, err)
		}
	}
}

func TestTrackClickIncrementsCount(t *testing.T) {
	s := newTestService(t)

	resp, err := s.CreateShortURL(models.CreateShortURLRequest{OriginalURL: "https://example.com"})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	for i := 0; i < 3; i++ {
		if err := s.TrackClick(resp.ShortCode, "agent", "1.2.3.4", "ref"); err != nil {
			t.Fatalf("TrackClick #%d failed: %v", i, err)
		}
	}

	link, err := s.GetLinkWithAnalytics(resp.ShortCode)
	if err != nil {
		t.Fatalf("GetLinkWithAnalytics failed: %v", err)
	}
	if link.ClickCount != 3 {
		t.Errorf("ClickCount = %d, want 3", link.ClickCount)
	}
	if link.LastAccessed == nil {
		t.Error("LastAccessed = nil, want a timestamp after clicks")
	}
}

func TestUpdateLinkURLOnly(t *testing.T) {
	s := newTestService(t)

	resp, _ := s.CreateShortURL(models.CreateShortURLRequest{OriginalURL: "https://old.com", CustomCode: "keep"})

	updated, err := s.UpdateLink("keep", models.UpdateLinkRequest{NewURL: "https://new.com"})
	if err != nil {
		t.Fatalf("UpdateLink failed: %v", err)
	}
	if updated.NewShortCode != "keep" {
		t.Errorf("NewShortCode = %q, want unchanged 'keep'", updated.NewShortCode)
	}

	got, _ := s.GetOriginalURL(resp.ShortCode)
	if got != "https://new.com" {
		t.Errorf("after update GetOriginalURL = %q, want %q", got, "https://new.com")
	}
}

func TestUpdateLinkChangesShortCode(t *testing.T) {
	s := newTestService(t)

	s.CreateShortURL(models.CreateShortURLRequest{OriginalURL: "https://example.com", CustomCode: "old-code"})

	updated, err := s.UpdateLink("old-code", models.UpdateLinkRequest{NewShortCode: "new-code"})
	if err != nil {
		t.Fatalf("UpdateLink failed: %v", err)
	}
	if updated.NewShortCode != "new-code" {
		t.Errorf("NewShortCode = %q, want %q", updated.NewShortCode, "new-code")
	}

	// New code resolves to the original URL...
	got, err := s.GetOriginalURL("new-code")
	if err != nil || got != "https://example.com" {
		t.Errorf("new code resolves to %q (err %v), want %q", got, err, "https://example.com")
	}

	// ...and the new row records its parent for traceability.
	link, err := s.GetLinkWithAnalytics("new-code")
	if err != nil {
		t.Fatalf("GetLinkWithAnalytics failed: %v", err)
	}
	if link.ParentShortCode == nil || *link.ParentShortCode != "old-code" {
		t.Errorf("ParentShortCode = %v, want 'old-code'", link.ParentShortCode)
	}
}

func TestUpdateLinkConflict(t *testing.T) {
	s := newTestService(t)

	s.CreateShortURL(models.CreateShortURLRequest{OriginalURL: "https://a.com", CustomCode: "taken"})
	s.CreateShortURL(models.CreateShortURLRequest{OriginalURL: "https://b.com", CustomCode: "source"})

	_, err := s.UpdateLink("source", models.UpdateLinkRequest{NewShortCode: "taken"})
	if err == nil {
		t.Fatal("expected error updating to an existing short code, got nil")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error = %q, want it to mention 'already exists'", err.Error())
	}
}

func TestUpdateLinkNotFound(t *testing.T) {
	s := newTestService(t)

	if _, err := s.UpdateLink("ghost", models.UpdateLinkRequest{NewURL: "https://x.com"}); err == nil {
		t.Fatal("expected error updating nonexistent link, got nil")
	}
}

func TestDeleteLinkRemovesLinkAndAnalytics(t *testing.T) {
	s := newTestService(t)

	resp, _ := s.CreateShortURL(models.CreateShortURLRequest{OriginalURL: "https://example.com", CustomCode: "del"})
	if err := s.TrackClick("del", "agent", "1.2.3.4", "ref"); err != nil {
		t.Fatalf("TrackClick failed: %v", err)
	}

	if err := s.DeleteLink("del"); err != nil {
		t.Fatalf("DeleteLink failed: %v", err)
	}

	if _, err := s.GetOriginalURL(resp.ShortCode); err == nil {
		t.Error("expected link to be gone after delete, but GetOriginalURL succeeded")
	}

	// Analytics for the deleted code should be gone too.
	analytics, err := s.GetAnalyticsPaginated(1, 50, "")
	if err != nil {
		t.Fatalf("GetAnalyticsPaginated failed: %v", err)
	}
	for _, c := range analytics.RecentClicks {
		if c.ShortCode == "del" {
			t.Error("found orphaned click analytics for deleted code 'del'")
		}
	}
}

func TestDeleteLinkNotFound(t *testing.T) {
	s := newTestService(t)

	if err := s.DeleteLink("ghost"); err == nil {
		t.Fatal("expected error deleting nonexistent link, got nil")
	}
}

func TestGetAnalyticsPaginatedAndSearch(t *testing.T) {
	s := newTestService(t)

	// 3 "alpha" links and 2 "beta" links.
	for i := 0; i < 3; i++ {
		s.CreateShortURL(models.CreateShortURLRequest{OriginalURL: "https://alpha.com/" + strings.Repeat("a", i+1)})
	}
	for i := 0; i < 2; i++ {
		s.CreateShortURL(models.CreateShortURLRequest{OriginalURL: "https://beta.com/" + strings.Repeat("b", i+1)})
	}

	// Totals across everything.
	all, err := s.GetAnalyticsPaginated(1, 50, "")
	if err != nil {
		t.Fatalf("GetAnalyticsPaginated failed: %v", err)
	}
	if all.TotalLinks != 5 {
		t.Errorf("TotalLinks = %d, want 5", all.TotalLinks)
	}

	// Search narrows results.
	betas, err := s.GetAnalyticsPaginated(1, 50, "beta.com")
	if err != nil {
		t.Fatalf("search query failed: %v", err)
	}
	if betas.TotalLinks != 2 {
		t.Errorf("search 'beta.com' TotalLinks = %d, want 2", betas.TotalLinks)
	}

	// Pagination math: 5 items, page size 2 => 3 pages.
	paged, err := s.GetAnalyticsPaginated(1, 2, "")
	if err != nil {
		t.Fatalf("paginated query failed: %v", err)
	}
	if len(paged.Links) != 2 {
		t.Errorf("page 1 returned %d links, want 2", len(paged.Links))
	}
	if paged.Pagination.TotalPages != 3 {
		t.Errorf("TotalPages = %d, want 3", paged.Pagination.TotalPages)
	}
	if !paged.Pagination.HasNext || paged.Pagination.HasPrev {
		t.Errorf("page 1: HasNext=%v HasPrev=%v, want true/false", paged.Pagination.HasNext, paged.Pagination.HasPrev)
	}
}
