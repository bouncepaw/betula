package activities

import "encoding/json"

func NewFollowFromUs(objectID string) ([]byte, error) {
	activity := dict{
		"@context": atContext,
		"type":     "Follow",
		"actor":    betulaActor,
		"object":   objectID,
	}
	return json.Marshal(activity)
}

type FollowReport struct {
	ActorID          string
	ObjectID         string
	OriginalActivity dict
}

func guessFollow(activity dict) (any, error) {
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
