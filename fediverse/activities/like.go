// SPDX-FileCopyrightText: 2025 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package activities

import (
	"encoding/json"
	"git.sr.ht/~bouncepaw/betula/settings"
	"git.sr.ht/~bouncepaw/betula/stricks"
	"path"
)

func NewLike(likedObjectID string) (json.RawMessage, error) {
	activity := Dict{
		"@context": atContext,
		"id":       path.Join(settings.SiteURL(), "temp", stricks.RandomWhatever()),
		"type":     "Like",
		"actor":    betulaActor,
		"object":   likedObjectID,
	}
	return json.Marshal(activity)
}

// LikeReport reports that actor with ActorID liked the object with ObjectID.
type LikeReport struct {
	ID       string
	ActorID  string
	ObjectID string
	Activity json.RawMessage
}

func (lr LikeReport) Valid() error {
	switch {
	case lr.ID == "":
		return ErrNoId
	case lr.ActorID == "":
		return ErrNoActor
	case lr.ObjectID == "":
		return ErrNoObject
	default:
		return nil
	}
}

func guessLike(activity Dict) (any, error) {
	report := LikeReport{
		ID:       getIDSomehow(activity, "id"),
		ActorID:  getIDSomehow(activity, "actor"),
		ObjectID: getIDSomehow(activity, "object"),
		Activity: json.RawMessage(activity["original activity"].([]byte)),
	}
	if err := report.Valid(); err != nil {
		return nil, err
	}

	return report, nil
}
