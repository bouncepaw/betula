// Package auth provides you functions that let you work with auth. All state is stored in-package. The password is stored hashed, so safe enough.
package auth

import (
	"database/sql"
	"git.sr.ht/~bouncepaw/betula/db"
	"golang.org/x/crypto/bcrypt"
	"log"
)

var (
	adminName    string
	passwordHash []byte
	ready        bool
)

// Initialize queries the database for auth information. Call on startup. The module handles all further invocations for you.
func Initialize() {
	var (
		name = db.MetaEntry[sql.NullString]("Admin username")
		pass = db.MetaEntry[sql.NullString]("Admin password hash")
	)
	ready = name.Valid && pass.Valid
	if ready {
		adminName = name.String
		passwordHash = []byte(pass.String)
	}
}

// Ready returns if the admin account is set up. If it is not, Betula should demand it and refuse to work until then.
func Ready() bool {
	if ready {
		return true
	}
	Initialize()
	return ready
}

// CredentialsMatch checks if the credentials match.
func CredentialsMatch(name, pass string) bool {
	if name != adminName {
		log.Println("Matching credentials. Name mismatches.")
		return false
	}
	err := bcrypt.CompareHashAndPassword(db.MetaEntry[[]byte]("Admin password hash"), []byte(pass))
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
	adminName = name

	var err error
	passwordHash, err = bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalln("While hashing:", err)
	}

	db.SetCredentials(adminName, string(passwordHash))
	Initialize()
}
