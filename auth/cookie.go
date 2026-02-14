// SPDX-FileCopyrightText: 2023 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2024 Danila Gorelko
//
// SPDX-License-Identifier: AGPL-3.0-only

package auth

import (
	"crypto/rand"
	"encoding/hex"
	"git.sr.ht/~bouncepaw/betula/db"
	"git.sr.ht/~bouncepaw/betula/tools"
	"git.sr.ht/~bouncepaw/betula/types"
	"net/http"
	"time"
)

const tokenName = "betula-token"

func Token(rq *http.Request) (string, error) {
	cookie, err := rq.Cookie(tokenName)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

func MarkCurrentSession(currentToken string, sessions []types.Session) []types.Session {
	for i, session := range sessions {
		if session.Token == currentToken {
			sessions[i].Current = true
			tools.MoveElement(sessions, i, 0)
			return sessions
		}
	}
	return sessions
}

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
func LogInResponse(userAgent string, w http.ResponseWriter) {
	token := randomString(24)
	db.AddSession(token, userAgent)
	http.SetCookie(w, newCookie(token, time.Now().Add(365*24*time.Hour)))
}

func Sessions() []types.Session {
	return db.Sessions()
}

func randomString(n int) string {
	bytes := make([]byte, n)
	_, _ = rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func newCookie(val string, t time.Time) *http.Cookie {
	return &http.Cookie{
		Name:     tokenName,
		Value:    val,
		Expires:  t,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	}
}
