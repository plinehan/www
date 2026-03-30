package sess

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"time"
)

const (
	allowedSudoEmail      = "plinehan@gmail.com"
	sudoSessionCookieName = "sudo_session"
	sessionMaxAge         = 7 * 24 * time.Hour
)

func SessionSigningKey() []byte {
	if s := os.Getenv("SESSION_SECRET"); s != "" {
		return []byte(s)
	}
	if cs := os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET"); cs != "" {
		h := sha256.Sum256([]byte("sudo-session:" + cs))
		return h[:]
	}
	slog.Warn("SESSION_SECRET and GOOGLE_OAUTH_CLIENT_SECRET unset; using insecure dev session key")
	return []byte("dev-insecure-sudo-session")
}

type cookieValue struct {
	Email     string    `json:"email"`
	ExpiresAt time.Time `json:"expires_at"`
}

type signedCookieValue struct {
	Value []byte
	Sig   []byte
}

func signedValue(email string, expiresAt time.Time) (string, error) {
	cookieValue, err := json.Marshal(cookieValue{
		Email:     email,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return "", err
	}
	mac := hmac.New(sha256.New, SessionSigningKey())
	mac.Write([]byte(cookieValue))
	sig := mac.Sum(nil)
	outer, err := json.Marshal(&signedCookieValue{
		Value: cookieValue,
		Sig:   sig,
	})
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(outer), nil
}

func parseCookie(value string) (email string, ok bool) {
	raw, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return "", false
	}
	signedCookieValue := &signedCookieValue{}
	err = json.Unmarshal(raw, signedCookieValue)
	if err != nil {
		return "", false
	}
	cookieValue := &cookieValue{}
	err = json.Unmarshal(signedCookieValue.Value, cookieValue)
	if err != nil {
		return "", false
	}
	h := hmac.New(sha256.New, SessionSigningKey())
	h.Write([]byte(signedCookieValue.Value))
	if !hmac.Equal(h.Sum(nil), signedCookieValue.Sig) {
		return "", false
	}
	if time.Now().Unix() > cookieValue.ExpiresAt.Unix() {
		return "", false
	}
	if cookieValue.Email != allowedSudoEmail {
		return "", false
	}
	return cookieValue.Email, true
}

func fromRequest(r *http.Request) (email string, ok bool) {
	c, err := r.Cookie(sudoSessionCookieName)
	if err != nil || c.Value == "" {
		return "", false
	}
	return parseCookie(c.Value)
}

func setCookie(w http.ResponseWriter, r *http.Request, email string) error {
	expiresAt := time.Now().Add(sessionMaxAge)
	val, err := signedValue(email, expiresAt)
	if err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sudoSessionCookieName,
		Value:    val,
		Path:     "/sudo",
		MaxAge:   int(sessionMaxAge.Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https",
	})
	return nil
}
