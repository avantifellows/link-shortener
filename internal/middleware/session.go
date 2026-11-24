package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// SessionCookieName is the cookie that keeps a person logged in to the dashboard.
const SessionCookieName = "lnk_session"

// SessionTTL is how long a dashboard login lasts.
const SessionTTL = 12 * time.Hour

// TokenMatches compares a given token with the expected one in constant time.
func TokenMatches(given, expected string) bool {
	if given == "" || expected == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(given), []byte(expected)) == 1
}

// sessionMAC signs an expiry time with the auth token. The cookie never holds the
// token itself, and rotating AUTH_TOKEN logs everybody out.
func sessionMAC(token string, expiry int64) string {
	mac := hmac.New(sha256.New, []byte(token))
	mac.Write([]byte("lnk-dashboard-session:" + strconv.FormatInt(expiry, 10)))
	return hex.EncodeToString(mac.Sum(nil))
}

// NewSessionValue returns a signed cookie value that expires SessionTTL after now.
func NewSessionValue(token string, now time.Time) string {
	expiry := now.Add(SessionTTL).Unix()
	return strconv.FormatInt(expiry, 10) + "." + sessionMAC(token, expiry)
}

// ValidSessionValue checks a cookie value's signature and expiry.
func ValidSessionValue(value, token string, now time.Time) bool {
	if token == "" {
		return false
	}
	expiryText, mac, ok := strings.Cut(value, ".")
	if !ok {
		return false
	}
	expiry, err := strconv.ParseInt(expiryText, 10, 64)
	if err != nil || now.Unix() >= expiry {
		return false
	}
	return hmac.Equal([]byte(mac), []byte(sessionMAC(token, expiry)))
}

// HasValidSession reports whether the request carries a valid dashboard login cookie.
func HasValidSession(r *http.Request) bool {
	c, err := r.Cookie(SessionCookieName)
	if err != nil {
		return false
	}
	return ValidSessionValue(c.Value, os.Getenv("AUTH_TOKEN"), time.Now())
}

// isHTMXRequest is true for requests sent by htmx from the dashboard. Browsers do not
// let other sites add this header to a cross-site request without a CORS preflight,
// which this server never grants, so it guards cookie-authenticated calls against CSRF
// (on top of the SameSite=Strict cookie).
func isHTMXRequest(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}
