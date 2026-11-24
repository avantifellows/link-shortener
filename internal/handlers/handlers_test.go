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

// login posts the token to the dashboard login form and returns the session cookie.
func login(t *testing.T, r http.Handler, token string) *httptest.ResponseRecorder {
	t.Helper()
	form := url.Values{}
	form.Set("token", token)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func sessionCookie(t *testing.T, rec *httptest.ResponseRecorder) *http.Cookie {
	t.Helper()
	for _, c := range rec.Result().Cookies() {
		if c.Name == "lnk_session" {
			return c
		}
	}
	t.Fatalf("no lnk_session cookie set; headers = %v", rec.Header())
	return nil
}

func TestDashboardShowsLoginWithoutSession(t *testing.T) {
	r := newTestRouter(t)
	code := createLink(t, r, "https://example.com/private-list")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (login page)", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `action="/auth/login"`) {
		t.Errorf("expected the login form, got %q", body)
	}
	if strings.Contains(body, testAuthToken) {
		t.Error("login page must not contain the auth token")
	}
	if strings.Contains(body, code) {
		t.Error("login page must not list links")
	}
}

func TestLoginWithWrongToken(t *testing.T) {
	r := newTestRouter(t)
	rec := login(t, r, "wrong-token")

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
	for _, c := range rec.Result().Cookies() {
		if c.Name == "lnk_session" {
			t.Fatal("wrong token must not set a session cookie")
		}
	}
}

func TestLoginThenDashboard(t *testing.T) {
	r := newTestRouter(t)
	code := createLink(t, r, "https://example.com/listed")

	rec := login(t, r, testAuthToken)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("login status = %d, want 303", rec.Code)
	}
	c := sessionCookie(t, rec)
	if !c.HttpOnly || c.SameSite != http.SameSiteStrictMode {
		t.Errorf("cookie must be HttpOnly and SameSite=Strict, got %+v", c)
	}
	if strings.Contains(c.Value, testAuthToken) {
		t.Error("cookie must not contain the auth token")
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(c)
	dash := httptest.NewRecorder()
	r.ServeHTTP(dash, req)

	if dash.Code != http.StatusOK {
		t.Fatalf("dashboard status = %d, want 200", dash.Code)
	}
	if ct := dash.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Errorf("Content-Type = %q, want text/html", ct)
	}
	body := dash.Body.String()
	if !strings.Contains(body, code) {
		t.Error("logged-in dashboard should list the link")
	}
	if strings.Contains(body, testAuthToken) {
		t.Error("dashboard must not contain the auth token")
	}
}

func TestSessionCookieAuthorizesDashboardAPICalls(t *testing.T) {
	r := newTestRouter(t)
	c := sessionCookie(t, login(t, r, testAuthToken))

	post := func(withHTMX bool) int {
		form := url.Values{}
		form.Set("original_url", "https://example.com/from-dashboard")
		req := httptest.NewRequest(http.MethodPost, "/api/links", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		if withHTMX {
			req.Header.Set("HX-Request", "true")
		}
		req.AddCookie(c)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		return rec.Code
	}

	if got := post(true); got != http.StatusOK {
		t.Errorf("htmx call with session: status = %d, want 200", got)
	}
	if got := post(false); got != http.StatusUnauthorized {
		t.Errorf("non-htmx call with only a cookie: status = %d, want 401 (CSRF guard)", got)
	}
}

func TestForgedSessionIsRejected(t *testing.T) {
	r := newTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "lnk_session", Value: "9999999999.deadbeef"})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if !strings.Contains(rec.Body.String(), `action="/auth/login"`) {
		t.Error("a forged cookie must get the login page")
	}
}

func TestLogoutClearsSession(t *testing.T) {
	r := newTestRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want 303", rec.Code)
	}
	c := sessionCookie(t, rec)
	if c.MaxAge >= 0 || c.Value != "" {
		t.Errorf("logout should expire the cookie, got %+v", c)
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
