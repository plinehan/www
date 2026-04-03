package sess

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const testKey = "test-signing-key"

// token signs a cookie value using testKey
func token(t *testing.T, email string, d time.Duration) string {
	t.Helper()
	t.Setenv("SESSION_SECRET", testKey)
	val, err := signedValue(email, time.Now().Add(d))
	require.NoError(t, err)
	return val
}

func TestSessionSigningKey_sessionSecret(t *testing.T) {
	t.Setenv("SESSION_SECRET", "my-secret")
	require.Equal(t, []byte("my-secret"), SessionSigningKey())
}

func TestSessionSigningKey_insecureDefault(t *testing.T) {
	t.Setenv("SESSION_SECRET", "")
	t.Setenv("GOOGLE_OAUTH_CLIENT_SECRET", "")
	require.Equal(t, []byte("dev-insecure-sudo-session"), SessionSigningKey())
}

func TestParseCookie_roundTrip(t *testing.T) {
	val := token(t, allowedSudoEmail, time.Hour)
	got, ok := parseCookie(val)
	require.True(t, ok)
	require.Equal(t, allowedSudoEmail, got)
}

func TestParseCookie_wrongKey(t *testing.T) {
	val := token(t, allowedSudoEmail, time.Hour)
	t.Setenv("SESSION_SECRET", "different-key")
	_, ok := parseCookie(val)
	require.False(t, ok)
}

func TestParseCookie_expired(t *testing.T) {
	val := token(t, allowedSudoEmail, -time.Hour)
	_, ok := parseCookie(val)
	require.False(t, ok)
}

func TestParseCookie_disallowedEmail(t *testing.T) {
	val := token(t, "someone@example.com", time.Hour)
	_, ok := parseCookie(val)
	require.False(t, ok)
}

func TestParseCookie_malformed(t *testing.T) {
	t.Setenv("SESSION_SECRET", testKey)
	for _, raw := range []string{"", ".", "nodot", "not-base64!.sig", "payload.not-base64!"} {
		_, ok := parseCookie(raw)
		require.False(t, ok, "parseCookie(%q) should fail", raw)
	}
}

func TestFromRequest_missingCookie(t *testing.T) {
	t.Setenv("SESSION_SECRET", testKey)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	_, ok := fromRequest(req)
	require.False(t, ok)
}

func TestFromRequest_validCookie(t *testing.T) {
	val := token(t, allowedSudoEmail, time.Hour)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: sudoSessionCookieName, Value: val})
	email, ok := fromRequest(req)
	require.True(t, ok)
	require.Equal(t, allowedSudoEmail, email)
}

func TestSetCookie_roundTrip(t *testing.T) {
	t.Setenv("SESSION_SECRET", testKey)
	rec := httptest.NewRecorder()
	require.NoError(t, setCookie(rec, httptest.NewRequest(http.MethodGet, "/", nil), allowedSudoEmail))

	req2 := httptest.NewRequest(http.MethodGet, "/sudo", nil)
	for _, c := range rec.Result().Cookies() {
		req2.AddCookie(c)
	}
	email, ok := fromRequest(req2)
	require.True(t, ok)
	require.Equal(t, allowedSudoEmail, email)
	require.Contains(t, rec.Header().Get("Set-Cookie"), "HttpOnly")
}

func TestSetCookie_secureBehindProxy(t *testing.T) {
	t.Setenv("SESSION_SECRET", testKey)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	require.NoError(t, setCookie(rec, req, allowedSudoEmail))
	require.Contains(t, rec.Header().Get("Set-Cookie"), "Secure")
}
