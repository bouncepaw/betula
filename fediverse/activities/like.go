// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package activities

import (
	"encoding/base64"
	"encoding/json"
	"git.sr.ht/~bouncepaw/betula/settings"
	"git.sr.ht/~bouncepaw/betula/stricks"
	"path"
)

func NewLike(likedObjectID, recipientID string) (json.RawMessage, error) {
	encID := base64.URLEncoding.EncodeToString([]byte(likedObjectID))
	activity := Dict{
		"@context": atContext,
		"id":       path.Join(settings.SiteURL(), "likes", encID),
		"type":     "Like",
		"actor":    betulaActor,
		"object":   likedObjectID,
		"to":       recipientID,
	}
	return json.Marshal(activity)
}

func NewUndoLike(likedObjectID, recipientID string) (json.RawMessage, error) {
	encID := base64.URLEncoding.EncodeToString([]byte(likedObjectID))
	activity := Dict{
		"@context": atContext,
		"id":       path.Join(settings.SiteURL(), "temp", stricks.RandomWhatever()),
		"type":     "Undo",
		"actor":    betulaActor,
		"to":       recipientID,
		"object": Dict{
			"actor":  betulaActor,
			"id":     path.Join(settings.SiteURL(), "likes", encID),
			"object": likedObjectID,
			"to":     recipientID,
			"type":   "Like",
		},
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
	}
	if activity["original activity"] != nil {
		report.Activity = activity["original activity"].([]byte)
	}
	if err := report.Valid(); err != nil {
		return nil, err
	}

	return report, nil
}
