package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/epnaj/projektowanie-aplikacji-internetowych/internal/core"
)

const (
	sessionName   = "pixel_session"
	maxAgeSeconds = 7 * 24 * 60 * 60 // one week
)

// encoding is URL-safe base64 without padding, so cookie values stay free of
// reserved characters.
var encoding = base64.RawURLEncoding

// SessionManager issues and verifies stateless, HMAC-signed session cookies.
// The cookie carries the user ID and an expiry, signed with a secret key so it
// cannot be forged or tampered with. It is HttpOnly and SameSite=Lax: the value
// is unreadable from JS (XSS mitigation) and not sent on cross-site requests.
//
// This is a deliberately tiny, dependency-free alternative to a session
// library. There is no server-side session store, so "logout" relies on the
// cookie being cleared (plus the short expiry) rather than server revocation.
type SessionManager struct {
	key []byte
}

// NewSessionManager builds a manager from a signing key (keep it secret and
// stable across restarts, e.g. from an env var; a changing key invalidates all
// existing sessions).
func NewSessionManager(authKey []byte) *SessionManager {
	return &SessionManager{key: authKey}
}

// sign returns the HMAC-SHA256 of msg under the manager's key.
func (m *SessionManager) sign(msg []byte) []byte {
	mac := hmac.New(sha256.New, m.key)
	mac.Write(msg)
	return mac.Sum(nil)
}

// Save writes a signed cookie carrying the user's ID and an expiry timestamp.
func (m *SessionManager) Save(w http.ResponseWriter, r *http.Request, userID core.ID) error {
	exp := time.Now().Add(maxAgeSeconds * time.Second).Unix()
	payload := strconv.FormatUint(uint64(userID), 10) + ":" + strconv.FormatInt(exp, 10)
	value := encoding.EncodeToString([]byte(payload)) + "." + encoding.EncodeToString(m.sign([]byte(payload)))

	http.SetCookie(w, &http.Cookie{
		Name:     sessionName,
		Value:    value,
		Path:     "/",
		MaxAge:   maxAgeSeconds,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}

// UserID verifies the cookie signature and expiry and returns the user ID, or
// ok=false when there is no valid, unexpired, untampered session.
func (m *SessionManager) UserID(r *http.Request) (core.ID, bool) {
	cookie, err := r.Cookie(sessionName)
	if err != nil {
		return 0, false
	}

	payloadB64, sigB64, ok := strings.Cut(cookie.Value, ".")
	if !ok {
		return 0, false
	}
	payload, err := encoding.DecodeString(payloadB64)
	if err != nil {
		return 0, false
	}
	sig, err := encoding.DecodeString(sigB64)
	if err != nil {
		return 0, false
	}
	if !hmac.Equal(sig, m.sign(payload)) {
		return 0, false
	}

	idStr, expStr, ok := strings.Cut(string(payload), ":")
	if !ok {
		return 0, false
	}
	exp, err := strconv.ParseInt(expStr, 10, 64)
	if err != nil || time.Now().Unix() >= exp {
		return 0, false
	}
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return 0, false
	}
	return core.ID(id), true
}

// Clear logs the user out by overwriting the cookie with an expired one.
func (m *SessionManager) Clear(w http.ResponseWriter, r *http.Request) error {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}
