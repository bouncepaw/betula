package activities

import (
	"encoding/json"
	"log"
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
			log.Printf("Ignoring unknown kind of activity: %s\n", raw)
			return nil, ErrUnknownType
		} else if v == "Delete" && activity["actor"] == activity["object"] {
			// Waiting for https://github.com/mastodon/mastodon/pull/22273
			log.Println("Somebody got deleted, scroll further.")
			return nil, nil
		} else {
			log.Printf("Handling activity: %s\n", raw)
			return f(activity)
		}
	default:
		return nil, ErrNoType
	}
}
