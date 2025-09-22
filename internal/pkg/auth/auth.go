package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
)

func NewUserID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b[:]), nil
}

func signUserID(uid string, secret []byte) string {
	h := hmac.New(sha256.New, secret)
	h.Write([]byte(uid))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}

func buildToken(uid string, secret []byte) string {
	return uid + "." + signUserID(uid, secret)
}

func ParseAndVerifyToken(token string, secret []byte) (string, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return "", errors.New("bad token format")
	}
	uid, sig := parts[0], parts[1]
	want := signUserID(uid, secret)
	if !hmac.Equal([]byte(sig), []byte(want)) {
		return "", errors.New("bad token signature")
	}
	return uid, nil
}

func UserIDFromRequest(r *http.Request, cookieName string, secret []byte) (string, bool) {
	c, err := r.Cookie(cookieName)
	if err != nil {
		return "", false
	}
	uid, err := ParseAndVerifyToken(c.Value, secret)
	if err != nil {
		return "", false
	}
	return uid, true
}

func IssueUserCookie(w http.ResponseWriter, cookieName string, secret []byte) (string, error) {
	uid, err := NewUserID()
	if err != nil {
		return "", err
	}
	token := buildToken(uid, secret)
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	return uid, nil
}
