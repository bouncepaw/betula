// Package auth provides you functions that let you work with auth. All state is stored in-package. The password is stored hashed, so safe enough.
package auth

import (
	"database/sql"
	"git.sr.ht/~bouncepaw/betula/db"
	"git.sr.ht/~bouncepaw/betula/settings"
	"golang.org/x/crypto/bcrypt"
	"log"
	"sync/atomic"
)

var (
	ready atomic.Bool
)

// Initialize queries the database for auth information. Call on startup. The module handles all further invocations for you.
func Initialize() {
	ready.Store(false)
	var (
		name = settings.AdminUsername()
		pass = db.MetaEntry[sql.NullString](db.BetulaMetaAdminPasswordHash)
	)
	ready.Store(name != "" && pass.Valid)
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
		log.Println("Matching credentials. Name mismatches.")
		return false
	}
	err := bcrypt.CompareHashAndPassword(db.MetaEntry[[]byte](db.BetulaMetaAdminPasswordHash), []byte(pass))
	if err != nil {
		log.Println("Matching credentials. Password mismatches.")
		return false
	}
	log.Println("Credentials match.")
	return true
}

// SetCredentials sets new credentials.
func SetCredentials(name, pass string) {
	log.Println("Setting new credentials")

	hash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalln("While hashing:", err)
	}

	db.SetCredentials(name, string(hash))
	Initialize()
}
