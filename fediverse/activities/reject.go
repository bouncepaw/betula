package activities

import (
	"encoding/json"
	"fmt"
	"git.sr.ht/~bouncepaw/betula/settings"
	"git.sr.ht/~bouncepaw/betula/stricks"
)

func NewReject(rejectedActivity dict) ([]byte, error) {
	delete(rejectedActivity, "@context")
	activity := dict{
		"@context": atContext,
		"id":       fmt.Sprintf("%s/temp/%s", settings.SiteURL(), stricks.RandomWhatever()),
		"type":     "Reject",
		"actor":    betulaActor,
		"object":   rejectedActivity,
	}
	return json.Marshal(activity)
}

type RejectReport struct {
	ActorID  string
	ObjectID string
	Object   dict
}

func guessReject(activity dict) (any, error) {
	report := RejectReport{
		ActorID:  getIDSomehow(activity, "actor"),
		ObjectID: getIDSomehow(activity, "object"),
	}
	if report.ActorID == "" {
		return nil, ErrNoActor
	}
	if report.ObjectID == "" {
		return nil, ErrNoActor
	}
	if obj, ok := activity["object"]; ok {
		switch v := obj.(type) {
		case dict:
			report.Object = v
		}
	}

	return report, nil
}
