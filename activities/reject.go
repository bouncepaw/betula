package activities

import "encoding/json"

func NewReject(rejectedActivity map[string]any) ([]byte, error) {
	delete(rejectedActivity, "@context")
	activity := map[string]any{
		"@context": "https://www.w3.org/ns/activitystreams",
		"type":     "Reject",
		"actor":    betulaActor,
		"object":   rejectedActivity,
	}
	return json.Marshal(activity)
}

type RejectReport struct {
	ActorID  string
	ObjectID string
	Object   map[string]any
}

func guessReject(activity map[string]any) (any, error) {
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
		case map[string]any:
			report.Object = v
		}
	}

	return report, nil
}
