// Package activities provides generation of JSON activities and activity data extraction from JSON.
//
// JSON activities are made with New* functions. They all have the same actor. Call GenerateBetulaActor to regenerate the actor.
package activities

import (
	"errors"
	"git.sr.ht/~bouncepaw/betula/settings"
	"git.sr.ht/~bouncepaw/betula/stricks"
)

func getIDSomehow(activity Dict, field string) string {
	m := activity[field]
	switch v := m.(type) {
	case string:
		if stricks.ValidURL(v) {
			return v
		}
		return ""
	}
	for k, v := range m.(Dict) {
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

func getString(activity Dict, field string) string {
	m := activity[field]
	switch v := m.(type) {
	case string:
		return v
	}
	return ""
}

const atContext = "https://www.w3.org/ns/activitystreams"
const publicAudience = "https://www.w3.org/ns/activitystreams#Public"

type Dict = map[string]any

var (
	ErrNoType          = errors.New("activities: type absent or invalid")
	ErrNoActor         = errors.New("activities: actor absent or invalid")
	ErrNoActorUsername = errors.New("activities: actor with absent or invalid username")
	ErrUnknownType     = errors.New("activities: unknown activity type")
	ErrNoId            = errors.New("activities: id absent or invalid")
	ErrNoObject        = errors.New("activities: object absent or invalid")
	ErrEmptyField      = errors.New("activities: empty field")
	ErrNotNote         = errors.New("activities: not a Note")
	ErrHostMismatch    = errors.New("activities: host mismatch")
)

var betulaActor string

// GenerateBetulaActor updates what actor to use for outgoing activities.
// It makes no validation.
func GenerateBetulaActor() {
	username := settings.AdminUsername()
	if username == "" {
		username = "betula"
	}
	betulaActor = settings.SiteURL() + "/@" + username
}
