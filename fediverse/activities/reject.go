package activities

import "encoding/json"

func NewReject(rejectedActivity dict) ([]byte, error) {
	delete(rejectedActivity, "@context")
	activity := dict{
		"@context": atContext,
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
