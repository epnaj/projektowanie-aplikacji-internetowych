package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// cookieFromSave runs Save and returns the request that would carry the issued
// session cookie back to the server.
func cookieFromSave(t *testing.T, m *SessionManager, userID uint64) *http.Request {
	t.Helper()
	rec := httptest.NewRecorder()
	if err := m.Save(rec, httptest.NewRequest(http.MethodGet, "/", nil), userID); err != nil {
		t.Fatalf("save: %v", err)
	}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	for _, c := range rec.Result().Cookies() {
		req.AddCookie(c)
	}
	return req
}

func TestSessionRoundTrip(t *testing.T) {
	m := NewSessionManager([]byte("super-secret-signing-key"))
	req := cookieFromSave(t, m, 42)

	id, ok := m.UserID(req)
	if !ok || id != 42 {
		t.Fatalf("round trip: want (42,true), got (%d,%v)", id, ok)
	}
}

func TestSessionTamperRejected(t *testing.T) {
	m := NewSessionManager([]byte("super-secret-signing-key"))
	req := cookieFromSave(t, m, 42)

	// Replace the signature with a wrong (but well-formed) one to break the HMAC.
	c, _ := req.Cookie(sessionName)
	payload, _, _ := strings.Cut(c.Value, ".")
	c.Value = payload + "." + encoding.EncodeToString(make([]byte, 32))

	bad := httptest.NewRequest(http.MethodGet, "/", nil)
	bad.AddCookie(c)
	if _, ok := m.UserID(bad); ok {
		t.Fatal("tampered cookie must be rejected")
	}
}

func TestSessionWrongKeyRejected(t *testing.T) {
	signer := NewSessionManager([]byte("key-A"))
	verifier := NewSessionManager([]byte("key-B"))
	req := cookieFromSave(t, signer, 7)

	if _, ok := verifier.UserID(req); ok {
		t.Fatal("cookie signed with a different key must be rejected")
	}
}

func TestSessionNoCookie(t *testing.T) {
	m := NewSessionManager([]byte("k"))
	if _, ok := m.UserID(httptest.NewRequest(http.MethodGet, "/", nil)); ok {
		t.Fatal("missing cookie must yield ok=false")
	}
}

func TestBcryptHashCompare(t *testing.T) {
	h := NewBcryptHasher()
	hash, err := h.Hash("password123")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if hash == "password123" {
		t.Fatal("hash must not equal the plaintext")
	}
	if !h.Compare(hash, "password123") {
		t.Fatal("correct password should compare true")
	}
	if h.Compare(hash, "wrong") {
		t.Fatal("wrong password should compare false")
	}
}
