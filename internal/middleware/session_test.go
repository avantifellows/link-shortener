package middleware

import (
	"strings"
	"testing"
	"time"
)

func TestSessionValueRoundTrip(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	v := NewSessionValue("tok", now)

	if strings.Contains(v, "tok") {
		t.Fatal("session value must not contain the token")
	}
	if !ValidSessionValue(v, "tok", now.Add(time.Hour)) {
		t.Error("fresh session should be valid")
	}
	if ValidSessionValue(v, "tok", now.Add(SessionTTL+time.Second)) {
		t.Error("expired session should be invalid")
	}
	if ValidSessionValue(v, "rotated-token", now) {
		t.Error("session signed with an old token should be invalid")
	}
	if ValidSessionValue(v, "", now) {
		t.Error("no token configured means no valid session")
	}
	for _, bad := range []string{"", "abc", "123.", ".abc", "notanumber.abc"} {
		if ValidSessionValue(bad, "tok", now) {
			t.Errorf("malformed value %q should be invalid", bad)
		}
	}
}

func TestTokenMatches(t *testing.T) {
	if !TokenMatches("a", "a") || TokenMatches("a", "b") || TokenMatches("", "") || TokenMatches("a", "") {
		t.Error("TokenMatches gave a wrong answer")
	}
}
