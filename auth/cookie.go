package auth

import (
	"crypto/rand"
	"encoding/hex"
	"git.sr.ht/~bouncepaw/betula/db"
	"net/http"
	"time"
)

const tokenName = "betula-token"

// AuthorizedFromRequest is true if the user is authorized.
func AuthorizedFromRequest(rq *http.Request) bool {
	cookie, err := rq.Cookie(tokenName)
	if err != nil {
		return false
	}
	return db.SessionExists(cookie.Value)
}

// LogoutFromRequest logs the user in the request out and rewrites the cookie in to an empty one.
func LogoutFromRequest(w http.ResponseWriter, rq *http.Request) {
	cookie, err := rq.Cookie(tokenName)
	if err == nil {
		http.SetCookie(w, newCookie("", time.Unix(0, 0)))
		db.StopSession(cookie.Value)
	}
}

// LogInResponse logs such user in and writes a cookie for them.
func LogInResponse(w http.ResponseWriter) {
	token := randomString(24)
	db.AddSession(token)
	http.SetCookie(w, newCookie(token, time.Now().Add(365*24*time.Hour)))
}

func randomString(n int) string {
	bytes := make([]byte, n)
	_, _ = rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func newCookie(val string, t time.Time) *http.Cookie {
	return &http.Cookie{
		Name:    tokenName,
		Value:   val,
		Expires: t,
		Path:    "/",
	}
}
