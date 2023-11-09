package activities

import (
	"database/sql"
	"git.sr.ht/~bouncepaw/betula/db"
	"git.sr.ht/~bouncepaw/betula/settings"
)

type actor struct {
	Type              string `json:"type"`
	Id                string `json:"id"`
	Inbox             string `json:"inbox"`
	Name              string `json:"name"`
	PreferredUsername string `json:"preferredUsername"`
}

var betulaActor actor

// GenerateBetulaActor updates what actor to use for outgoing activities.
// It makes no validation.
func GenerateBetulaActor() {
	username := "betula"
	usernameMaybe := db.MetaEntry[sql.NullString](db.BetulaMetaAdminUsername)
	if usernameMaybe.Valid {
		username = usernameMaybe.String
	}
	betulaActor = actor{
		Type:              "Person",
		Id:                settings.SiteURL(),
		Inbox:             settings.SiteURL() + "/inbox",
		Name:              settings.SiteName(),
		PreferredUsername: username,
	}
}
