// SPDX-FileCopyrightText: 2023 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2024 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

// Package auth provides you functions that let you work with auth. All state is stored in-package. The password is stored hashed, so safe enough.
package auth

import (
	"database/sql"
	"log/slog"
	"os"
	"sync/atomic"

	"golang.org/x/crypto/bcrypt"

	"git.sr.ht/~bouncepaw/betula/db"
	"git.sr.ht/~bouncepaw/betula/settings"
)

var (
	ready atomic.Bool
)

// Initialize queries the database for auth information. Call on startup. The module handles all further invocations for you.
func Initialize() {
	ready.Store(false)
	var (
		name = db.MetaEntry[sql.NullString](db.BetulaMetaAdminUsername)
		pass = db.MetaEntry[sql.NullString](db.BetulaMetaAdminPasswordHash)
	)
	ready.Store(name.Valid && pass.Valid)
}

// Ready returns if the admin account is set up. If it is not, Betula should demand it and refuse to work until then.
func Ready() bool {
	ready := ready.Load()
	if ready {
		return true
	}
	Initialize()
	return ready
}

// CredentialsMatch checks if the credentials match.
func CredentialsMatch(name, pass string) bool {
	if name != settings.AdminUsername() {
		slog.Info("Matching credentials. Name mismatch")
		return false
	}
	err := bcrypt.CompareHashAndPassword(db.MetaEntry[[]byte](db.BetulaMetaAdminPasswordHash), []byte(pass))
	if err != nil {
		slog.Info("Matching credentials. Password mismatch")
		return false
	}
	slog.Info("Credentials match")
	return true
}

// SetCredentials sets new credentials.
func SetCredentials(name, pass string) {
	slog.Info("Setting new credentials")

	hash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		slog.Error("Failed to hash password", "err", err)
		os.Exit(1)
	}

	db.SetCredentials(name, string(hash))
	Initialize()
	settings.Index()
}
