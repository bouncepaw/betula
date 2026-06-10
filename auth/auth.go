// SPDX-FileCopyrightText: 2023 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2024 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

// Package auth provides you functions that let you work with auth. All state is stored in-package. The password is stored hashed, so safe enough.
package auth

import (
	"context"
	"log/slog"
	"os"
	"sync/atomic"

	"golang.org/x/crypto/bcrypt"

	"git.sr.ht/~bouncepaw/betula/db"
	"git.sr.ht/~bouncepaw/betula/ports/settings"
	"git.sr.ht/~bouncepaw/betula/settings"
)

var (
	ready        atomic.Bool
	settingsRepo = &db.SettingsRepo{}
)

// Initialize queries the database for auth information. Call on startup. The module handles all further invocations for you.
func Initialize() {
	ready.Store(false)
	ctx := context.Background()
	name, err := settingsRepo.MetaEntryNullString(ctx, settingsports.BetulaMetaAdminUsername)
	if err != nil {
		slog.Error("Failed to read admin username", "err", err)
		os.Exit(1)
	}
	pass, err := settingsRepo.MetaEntryNullString(ctx, settingsports.BetulaMetaAdminPasswordHash)
	if err != nil {
		slog.Error("Failed to read admin password hash", "err", err)
		os.Exit(1)
	}
	ready.Store(name.Valid && pass.Valid)
}

// Ready returns if the admin account is set up. If it is not, Betula should demand it and refuse to work until then.
func Ready() bool {
	if ready.Load() {
		return true
	}
	Initialize()
	return ready.Load()
}

// CredentialsMatch checks if the credentials match.
func CredentialsMatch(name, pass string) bool {
	if name != settings.AdminUsername() {
		slog.Info("Matching credentials. Name mismatch")
		return false
	}
	hash, err := settingsRepo.MetaEntryBytes(context.Background(), settingsports.BetulaMetaAdminPasswordHash)
	if err != nil {
		slog.Error("Failed to read admin password hash", "err", err)
		return false
	}
	if err := bcrypt.CompareHashAndPassword(hash, []byte(pass)); err != nil {
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

	if err := settingsRepo.SetCredentials(context.Background(), name, string(hash)); err != nil {
		slog.Error("Failed to set credentials", "err", err)
		os.Exit(1)
	}
	Initialize()
	settings.Index()
}
