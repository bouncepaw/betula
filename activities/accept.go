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
