package activities

import (
	"encoding/json"
)

var guesserMap = map[string]func(dict) (any, error){
	"Announce": guessAnnounce,
	"Undo":     guessUndo,
	"Follow":   guessFollow,
	"Accept":   guessAccept,
	"Reject":   guessReject,
}

func Guess(raw []byte) (report any, err error) {
	var (
		activity = dict{}
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
		if f, ok := guesserMap[v]; !ok {
			return nil, ErrUnknownType
		} else {
			return f(activity)
		}
	default:
		return nil, ErrNoType
	}
}
