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
