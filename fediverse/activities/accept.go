// SPDX-FileCopyrightText: 2022-2025 Betula contributors
//
// SPDX-License-Identifier: AGPL-3.0-only

package activities

import (
	"encoding/json"
	"fmt"
	"git.sr.ht/~bouncepaw/betula/settings"
	"git.sr.ht/~bouncepaw/betula/stricks"
)

// NewAccept wraps the acceptedActivity in an Accept activity.
// The @context of the wrapped activity is deleted.
func NewAccept(acceptedActivity Dict) ([]byte, error) {
	delete(acceptedActivity, "@context")
	return json.Marshal(Dict{
		"@context": atContext,
		"id":       fmt.Sprintf("%s/temp/%s", settings.SiteURL(), stricks.RandomWhatever()),
		"type":     "Accept",
		"actor":    betulaActor,
		"object":   acceptedActivity,
	})
}

type AcceptReport struct {
	ActorID  string
	ObjectID string
	Object   Dict
}

func guessAccept(activity Dict) (any, error) {
	report := AcceptReport{
		ActorID:  getIDSomehow(activity, "actor"),
		ObjectID: getIDSomehow(activity, "object"),
	}
	if report.ActorID == "" {
		return nil, ErrNoActor
	}
	if report.ObjectID == "" {
		return nil, ErrNoObject
	}
	if obj, ok := activity["object"]; ok {
		switch v := obj.(type) {
		case Dict:
			report.Object = v
		}
	}

	return report, nil
}
