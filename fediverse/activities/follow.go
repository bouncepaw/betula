package activities

import (
	"encoding/json"
	"fmt"
	"git.sr.ht/~bouncepaw/betula/settings"
)

func NewUndoFollowFromUs(objectID string) ([]byte, error) {
	activity := Dict{
		"@context": atContext,
		"id":       fmt.Sprintf("%s/unfollow?account=%s", settings.SiteURL(), objectID),
		"type":     "Undo",
		"actor":    betulaActor,
		"object": Dict{
			"id":     fmt.Sprintf("%s/follow?account=%s", settings.SiteURL(), objectID),
			"type":   "Follow",
			"actor":  betulaActor,
			"object": objectID,
		},
	}
	return json.Marshal(activity)
}

func NewFollowFromUs(objectID string) ([]byte, error) {
	activity := Dict{
		"@context": atContext,
		"id":       fmt.Sprintf("%s/follow?account=%s", settings.SiteURL(), objectID),
		"type":     "Follow",
		"actor":    betulaActor,
		"object":   objectID,
	}
	return json.Marshal(activity)
}

type FollowReport struct {
	ActorID          string
	ObjectID         string
	OriginalActivity Dict
}

func guessFollow(activity Dict) (any, error) {
	report := FollowReport{
		ActorID:          getIDSomehow(activity, "actor"),
		ObjectID:         getIDSomehow(activity, "object"),
		OriginalActivity: activity,
	}
	if report.ActorID == "" {
		return nil, ErrNoActor
	}
	if report.ObjectID == "" {
		return nil, ErrNoObject
	}
	return report, nil
}
