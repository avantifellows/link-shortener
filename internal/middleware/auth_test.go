package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// okHandler is a trivial downstream handler that records whether it was reached.
func okHandler(reached *bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*reached = true
		w.WriteHeader(http.StatusOK)
	})
}

func TestAuthMiddleware(t *testing.T) {
	const token = "secret-token"

	cases := []struct {
		name        string
		authHeader  string
		wantStatus  int
		wantReached bool
	}{
		{"valid token", "Bearer " + token, http.StatusOK, true},
		{"missing header", "", http.StatusUnauthorized, false},
		{"wrong token", "Bearer nope", http.StatusUnauthorized, false},
		{"not bearer scheme", "Basic " + token, http.StatusUnauthorized, false},
		{"token without bearer prefix", token, http.StatusUnauthorized, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("AUTH_TOKEN", token)

			reached := false
			h := AuthMiddleware(okHandler(&reached))

			req := httptest.NewRequest(http.MethodGet, "/api/links", nil)
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tc.wantStatus)
			}
			if reached != tc.wantReached {
				t.Errorf("downstream reached = %v, want %v", reached, tc.wantReached)
			}
		})
	}
}

func TestAuthMiddlewareMisconfigured(t *testing.T) {
	t.Setenv("AUTH_TOKEN", "")

	reached := false
	h := AuthMiddleware(okHandler(&reached))

	req := httptest.NewRequest(http.MethodGet, "/api/links", nil)
	req.Header.Set("Authorization", "Bearer anything")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d when AUTH_TOKEN unset", rec.Code, http.StatusInternalServerError)
	}
	if reached {
		t.Error("downstream handler should not be reached when AUTH_TOKEN is unset")
	}
}
