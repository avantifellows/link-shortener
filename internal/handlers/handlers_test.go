package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/avantifellows/link-shortener/internal/database"
	"github.com/avantifellows/link-shortener/internal/handlers"
	"github.com/avantifellows/link-shortener/internal/router"
)

const testAuthToken = "test-token"

// TestMain changes the working directory to the repository root so that
// handlers.New can find templates/*.html (and the static/ dir) via their
// relative paths, exactly as the running server does.
func TestMain(m *testing.M) {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			panic("could not locate repository root (go.mod) from " + dir)
		}
		dir = parent
	}
	if err := os.Chdir(dir); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

// newTestRouter builds a fully wired router backed by a throwaway database.
func newTestRouter(t *testing.T) http.Handler {
	t.Helper()

	t.Setenv("DATABASE_PATH", t.TempDir()+"/test.db")
	t.Setenv("BASE_URL", "http://test.local")
	t.Setenv("AUTH_TOKEN", testAuthToken)

	db, err := database.Initialize()
	if err != nil {
		t.Fatalf("failed to initialize test database: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	return router.New(handlers.New(db))
}

// createLink shortens a URL through the authenticated JSON API and returns the
// generated short code.
func createLink(t *testing.T, r http.Handler, original string) string {
	t.Helper()

	form := url.Values{}
	form.Set("original_url", original)
	form.Set("created_by", "tester")

	req := httptest.NewRequest(http.MethodPost, "/api/links", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+testAuthToken)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("create link: status = %d, body = %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		ShortCode string `json:"short_code"`
		ShortURL  string `json:"short_url"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("create link: failed to decode response %q: %v", rec.Body.String(), err)
	}
	if resp.ShortCode == "" {
		t.Fatalf("create link: empty short code in response %q", rec.Body.String())
	}
	return resp.ShortCode
}

func TestHealth(t *testing.T) {
	r := newTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"healthy"`) {
		t.Errorf("body = %q, want it to contain \"healthy\"", rec.Body.String())
	}
}

// TestShortenAndRedirect is the headline end-to-end test: shorten a URL over
// HTTP, then follow the short code and confirm it redirects to the original.
func TestShortenAndRedirect(t *testing.T) {
	r := newTestRouter(t)

	const original = "https://example.com/landing?utm=abc"
	code := createLink(t, r, original)

	req := httptest.NewRequest(http.MethodGet, "/"+code, nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("redirect status = %d, want %d (302)", rec.Code, http.StatusFound)
	}
	if loc := rec.Header().Get("Location"); loc != original {
		t.Errorf("Location = %q, want %q", loc, original)
	}
}

func TestRedirectUnknownCodeReturns404(t *testing.T) {
	r := newTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/doesnotexist", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestCreateRequiresAuth(t *testing.T) {
	r := newTestRouter(t)

	form := url.Values{}
	form.Set("original_url", "https://example.com")

	cases := []struct {
		name   string
		header string
	}{
		{"no token", ""},
		{"wrong token", "Bearer wrong"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/links", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Set("Accept", "application/json")
			if tc.header != "" {
				req.Header.Set("Authorization", tc.header)
			}
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Errorf("status = %d, want 401", rec.Code)
			}
		})
	}
}

func TestCreateInvalidURLReturns400(t *testing.T) {
	r := newTestRouter(t)

	form := url.Values{}
	form.Set("original_url", "not-a-valid-url")

	req := httptest.NewRequest(http.MethodPost, "/api/links", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+testAuthToken)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestDashboardIsPublic(t *testing.T) {
	r := newTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (dashboard should need no auth)", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Errorf("Content-Type = %q, want text/html", ct)
	}
}

func TestGetLinkAPI(t *testing.T) {
	r := newTestRouter(t)
	code := createLink(t, r, "https://example.com/details")

	req := httptest.NewRequest(http.MethodGet, "/api/links/"+code, nil)
	req.Header.Set("Authorization", "Bearer "+testAuthToken)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body = %s", rec.Code, rec.Body.String())
	}

	var link struct {
		ShortCode   string `json:"short_code"`
		OriginalURL string `json:"original_url"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &link); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if link.ShortCode != code || link.OriginalURL != "https://example.com/details" {
		t.Errorf("got %+v, want code %q and original URL set", link, code)
	}
}

func TestDeleteLinkAPI(t *testing.T) {
	r := newTestRouter(t)
	code := createLink(t, r, "https://example.com/temp")

	req := httptest.NewRequest(http.MethodDelete, "/api/links/"+code, nil)
	req.Header.Set("Authorization", "Bearer "+testAuthToken)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("delete status = %d, want 200, body = %s", rec.Code, rec.Body.String())
	}

	// The code should no longer redirect.
	req = httptest.NewRequest(http.MethodGet, "/"+code, nil)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("after delete, redirect status = %d, want 404", rec.Code)
	}
}

func TestUpdateLinkChangesRedirectTarget(t *testing.T) {
	r := newTestRouter(t)
	code := createLink(t, r, "https://old-target.com")

	form := url.Values{}
	form.Set("new_url", "https://new-target.com")

	req := httptest.NewRequest(http.MethodPut, "/api/links/"+code, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+testAuthToken)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("update status = %d, want 200, body = %s", rec.Code, rec.Body.String())
	}

	// Following the code now lands on the new target.
	req = httptest.NewRequest(http.MethodGet, "/"+code, nil)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if loc := rec.Header().Get("Location"); loc != "https://new-target.com" {
		t.Errorf("Location after update = %q, want %q", loc, "https://new-target.com")
	}
}

func TestAnalyticsJSON(t *testing.T) {
	r := newTestRouter(t)
	createLink(t, r, "https://example.com/one")
	createLink(t, r, "https://example.com/two")

	req := httptest.NewRequest(http.MethodGet, "/api/links", nil)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+testAuthToken)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var resp struct {
		TotalLinks int `json:"total_links"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if resp.TotalLinks != 2 {
		t.Errorf("TotalLinks = %d, want 2", resp.TotalLinks)
	}
}
