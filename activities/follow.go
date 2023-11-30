package activities

import "encoding/json"

func NewFollow(objectID string) ([]byte, error) {
	activity := map[string]any{
		"@context": "https://www.w3.org/ns/activitystreams",
		"type":     "Follow",
		"actor":    betulaActor,
		"object":   objectID,
	}
	return json.Marshal(activity)
}

func getIDSomehow(activity map[string]any, field string) string {
	m := activity[field]
	switch v := m.(type) {
	case string:
		return v
	}
	for k, v := range m.(map[string]any) {
		if k != "id" {
			continue
		}
		switch v := v.(type) {
		case string:
			return v
		default:
			return ""
		}
	}
	return ""
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
