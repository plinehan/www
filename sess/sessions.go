package sess

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"strings"
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
	slog.Warn("SESSION_SECRET unset; using insecure dev session key")
	return []byte("dev-insecure-sudo-session")
}

type cookiePayload struct {
	Email     string `json:"email"`
	ExpiresAt int64  `json:"expires_at"`
}

func sign(key, data []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	return mac.Sum(nil)
}

func signedValue(email string, expiresAt time.Time) (string, error) {
	payload, err := json.Marshal(cookiePayload{Email: email, ExpiresAt: expiresAt.Unix()})
	if err != nil {
		return "", err
	}
	b64 := base64.RawURLEncoding.EncodeToString(payload)
	sig := base64.RawURLEncoding.EncodeToString(sign(SessionSigningKey(), []byte(b64)))
	return b64 + "." + sig, nil
}

func parseCookie(value string) (email string, ok bool) {
	dot := strings.LastIndex(value, ".")
	if dot < 0 {
		return "", false
	}
	b64payload, b64sig := value[:dot], value[dot+1:]

	sig, err := base64.RawURLEncoding.DecodeString(b64sig)
	if err != nil {
		return "", false
	}
	if !hmac.Equal(sign(SessionSigningKey(), []byte(b64payload)), sig) {
		return "", false
	}
	payload, err := base64.RawURLEncoding.DecodeString(b64payload)
	if err != nil {
		return "", false
	}
	var p cookiePayload
	if err = json.Unmarshal(payload, &p); err != nil {
		return "", false
	}
	if time.Now().Unix() > p.ExpiresAt {
		return "", false
	}
	if p.Email != allowedSudoEmail {
		return "", false
	}
	return p.Email, true
}

func fromRequest(r *http.Request) (email string, ok bool) {
	c, err := r.Cookie(sudoSessionCookieName)
	if err != nil || c.Value == "" {
		return "", false
	}
	return parseCookie(c.Value)
}

func setCookie(w http.ResponseWriter, r *http.Request, email string) error {
	val, err := signedValue(email, time.Now().Add(sessionMaxAge))
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
