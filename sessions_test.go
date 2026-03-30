package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewSessions(t *testing.T) {
	sess := newSessions([]byte("k"))
	require.NotNil(t, sess)
	require.Equal(t, "k", string(sess.key))
}

func TestSessionSigningKey(t *testing.T) {
	tests := []struct {
		name     string
		session  string
		oauth    string
		want     string
		wantHash bool
	}{
		{
			name:    "SESSION_SECRET wins",
			session: "only-session",
			oauth:   "oauth-secret",
			want:    "only-session",
		},
		{
			name:     "GOOGLE_OAUTH_CLIENT_SECRET derived",
			session:  "",
			oauth:    "client-secret",
			wantHash: true,
		},
		{
			name: "insecure default when unset",
			want: "dev-insecure-sudo-session",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SESSION_SECRET", tt.session)
			t.Setenv("GOOGLE_OAUTH_CLIENT_SECRET", tt.oauth)

			got := sessionSigningKey()
			if tt.wantHash {
				require.Len(t, got, 32, "sessionSigningKey() should be sha256")
				require.NotEqual(t, tt.oauth, string(got), "must not equal raw client secret")
			} else {
				require.Equal(t, tt.want, string(got))
			}
		})
	}
}

func TestParseCookie_roundTrip(t *testing.T) {
	sess := newSessions([]byte("test-signing-key-32bytes!!"))
	exp := time.Now().Add(time.Hour)
	val, err := sess.signedValue(allowedSudoEmail, exp)
	require.NoError(t, err)
	got, ok := sess.parseCookie(val)
	require.True(t, ok)
	require.Equal(t, allowedSudoEmail, got)
}

func TestParseCookie_wrongSigningKey(t *testing.T) {
	a := newSessions([]byte("key-a"))
	b := newSessions([]byte("key-b"))
	exp := time.Now().Add(time.Hour)
	val, err := a.signedValue(allowedSudoEmail, exp)
	require.NoError(t, err)
	_, ok := b.parseCookie(val)
	require.False(t, ok, "parseCookie accepted value signed with different key")
}

func TestParseCookie_expired(t *testing.T) {
	sess := newSessions([]byte("key"))
	exp := time.Now().Add(-time.Hour)
	val, err := sess.signedValue(allowedSudoEmail, exp)
	require.NoError(t, err)
	_, ok := sess.parseCookie(val)
	require.False(t, ok, "parseCookie accepted expired session")
}

func TestParseCookie_disallowedEmail(t *testing.T) {
	sess := newSessions([]byte("key"))
	exp := time.Now().Add(time.Hour)
	val, err := sess.signedValue("someone.else@example.com", exp)
	require.NoError(t, err)
	_, ok := sess.parseCookie(val)
	require.False(t, ok, "parseCookie accepted non-allowlisted email")
}

func TestParseCookie_malformed(t *testing.T) {
	sess := newSessions([]byte("key"))
	for _, raw := range []string{
		"",
		"nodot",
		"not-base64!.deadbeef",
		"a.b",
	} {
		_, ok := sess.parseCookie(raw)
		require.False(t, ok, "parseCookie(%q)", raw)
	}
}

func TestFromRequest(t *testing.T) {
	sess := newSessions([]byte("key"))
	t.Run("missing cookie", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		_, ok := sess.fromRequest(req)
		require.False(t, ok, "fromRequest without cookie should fail")
	})
	t.Run("valid cookie", func(t *testing.T) {
		val, err := sess.signedValue(allowedSudoEmail, time.Now().Add(time.Hour))
		require.NoError(t, err)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.AddCookie(&http.Cookie{Name: sudoSessionCookieName, Value: val})
		email, ok := sess.fromRequest(req)
		require.True(t, ok)
		require.Equal(t, allowedSudoEmail, email)
	})
}

func TestSetCookie(t *testing.T) {
	sess := newSessions([]byte("key"))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	require.NoError(t, sess.setCookie(rec, req, allowedSudoEmail))
	res := rec.Result()
	defer res.Body.Close()

	req2 := httptest.NewRequest(http.MethodGet, "/sudo", nil)
	for _, c := range res.Cookies() {
		if c.Name == sudoSessionCookieName {
			req2.AddCookie(c)
		}
	}
	email, ok := sess.fromRequest(req2)
	require.True(t, ok)
	require.Equal(t, allowedSudoEmail, email)
	require.Contains(t, res.Header.Get("Set-Cookie"), "HttpOnly")
}

func TestSetCookie_secureBehindProxy(t *testing.T) {
	sess := newSessions([]byte("key"))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	require.NoError(t, sess.setCookie(rec, req, allowedSudoEmail))
	raw := rec.Header().Get("Set-Cookie")
	require.Contains(t, raw, "Secure", "expected Secure flag when X-Forwarded-Proto=https: %q", raw)
}
