package activities

import "encoding/json"

func NewAccept(acceptedActivity map[string]any) ([]byte, error) {
	delete(acceptedActivity, "@context")
	activity := map[string]any{
		"@context": "https://www.w3.org/ns/activitystreams",
		"type":     "Accept",
		"actor":    betulaActor,
		"object":   acceptedActivity,
	}
	return json.Marshal(activity)
}

type AcceptReport struct {
	ActorID  string
	ObjectID string
	Object   map[string]any
}

func guessAccept(activity map[string]any) (any, error) {
	report := AcceptReport{
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
		case map[string]any:
			report.Object = v
		}
	}

	return report, nil
}
