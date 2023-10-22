package activities

import (
	"encoding/json"
	"errors"
)

var (
	ErrNoType          = errors.New("activities: type absent or invalid")
	ErrNoActor         = errors.New("activities: actor absent or invalid")
	ErrNoActorUsername = errors.New("activities: actor with absent or invalid username")
	ErrUnknownType     = errors.New("activities: unknown activity type")
	ErrNoId            = errors.New("activities: id absent or invalid")
	ErrNoObject        = errors.New("activities: object absent or invalid")
)

func Guess(raw []byte) (report any, err error) {
	var (
		activity map[string]any
		val      any
		ok       bool
	)
	if err = json.Unmarshal(raw, &activity); err != nil {
		return nil, err
	}

	if val, ok = activity["type"]; !ok {
		return nil, ErrNoType
	}
	switch v := val.(type) {
	case string:
		switch v {
		case "Announce":
			return guessAnnounce(activity)
		case "Undo":
			return guessUndo(activity)
		default:
			return nil, ErrUnknownType
		}
	default:
		return nil, ErrNoType
	}
}
