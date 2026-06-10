// SPDX-FileCopyrightText: 2023 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2024 Danila Gorelko
// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"time"

	"git.sr.ht/~bouncepaw/betula/db"
	"git.sr.ht/~bouncepaw/betula/pkg/bxslices"
	"git.sr.ht/~bouncepaw/betula/types"
)

const tokenName = "betula-token"

var sessionsRepo = db.NewSessionsRepo()

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
			bxslices.MoveElement(sessions, i, 0)
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
	exists, err := sessionsRepo.SessionExists(context.Background(), cookie.Value)
	if err != nil {
		slog.Error("Failed to check session existence", "err", err)
		return false
	}
	return exists
}

// LogoutFromRequest logs the user in the request out and rewrites the cookie in to an empty one.
func LogoutFromRequest(w http.ResponseWriter, rq *http.Request) {
	cookie, err := rq.Cookie(tokenName)
	if err == nil {
		http.SetCookie(w, newCookie("", time.Unix(0, 0)))
		StopSession(cookie.Value)
	}
}

// LogInResponse logs such user in and writes a cookie for them.
func LogInResponse(userAgent string, w http.ResponseWriter) {
	token := randomString(24)
	if err := sessionsRepo.AddSession(context.Background(), token, userAgent); err != nil {
		slog.Error("Failed to add session", "err", err)
		return
	}
	http.SetCookie(w, newCookie(token, time.Now().Add(365*24*time.Hour)))
}

// StopSession ends the session with the given token.
func StopSession(token string) {
	if err := sessionsRepo.StopSession(context.Background(), token); err != nil {
		slog.Error("Failed to stop session", "err", err)
	}
}

// StopAllSessions ends every session except the one with excludeToken.
func StopAllSessions(excludeToken string) {
	if err := sessionsRepo.StopAllSessions(context.Background(), excludeToken); err != nil {
		slog.Error("Failed to stop sessions", "err", err)
	}
}

func Sessions() []types.Session {
	sessions, err := sessionsRepo.Sessions(context.Background())
	if err != nil {
		slog.Error("Failed to load sessions", "err", err)
		return nil
	}
	return sessions
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
