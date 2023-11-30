package activities

import "encoding/json"

func NewFollow(objectID string) ([]byte, error) {
	activity := dict{
		"@context": atContext,
		"type":     "Follow",
		"actor":    betulaActor,
		"object":   objectID,
	}
	return json.Marshal(activity)
}

type FollowReport struct {
	ActorID  string
	ObjectID string
}

func guessFollow(activity map[string]any) (any, error) {
	report := FollowReport{
		ActorID:  getIDSomehow(activity, "actor"),
		ObjectID: getIDSomehow(activity, "object"),
	}
	if report.ActorID == "" {
		return nil, ErrNoActor
	}
	if report.ObjectID == "" {
		return nil, ErrNoActor
	}
	return report, nil
}
