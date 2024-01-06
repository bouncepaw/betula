// Package activities provides generation of JSON activities and activity data extraction from JSON.
//
// JSON activities are made with New* functions. They all have the same actor. Call GenerateBetulaActor to regenerate the actor.
package activities

import (
	"errors"

	"git.sr.ht/~bouncepaw/betula/db"
	"git.sr.ht/~bouncepaw/betula/settings"
	"git.sr.ht/~bouncepaw/betula/types"
)

func getIDSomehow(activity map[string]any, field string) string {
	m := activity[field]
	switch v := m.(type) {
	case string:
		return v
	}
	for k, v := range m.(map[string]any) {
		if k != "id" {
			continue
		}
		switch v := v.(type) {
		case string:
			return v
		default:
			return ""
		}
	}
	return ""
}

const atContext = "https://www.w3.org/ns/activitystreams"

type dict = map[string]any

var (
	ErrNoType          = errors.New("activities: type absent or invalid")
	ErrNoActor         = errors.New("activities: actor absent or invalid")
	ErrNoActorUsername = errors.New("activities: actor with absent or invalid username")
	ErrUnknownType     = errors.New("activities: unknown activity type")
	ErrNoId            = errors.New("activities: id absent or invalid")
	ErrNoObject        = errors.New("activities: object absent or invalid")
)

var betulaActor types.Actor

// GenerateBetulaActor updates what actor to use for outgoing activities.
// It makes no validation.
func GenerateBetulaActor() {
	username := db.AdminUsername()
	if username == "" {
		username = "betula"
	}
	betulaActor = types.Actor{
		ID:                settings.SiteURL() + "/@" + db.AdminUsername(),
		Type:              "Person",
		Inbox:             settings.SiteURL() + "/inbox",
		PreferredUsername: username,
		DisplayedName:     settings.SiteName(),
		Summary:           "", // TODO:
		ServerSoftware:    "", // TODO:
	}
}
