// SPDX-FileCopyrightText: 2022-2025 Betula contributors
//
// SPDX-License-Identifier: AGPL-3.0-only

package activities

import (
	"encoding/json"
	"log"
)

var guesserMap = map[string]func(Dict) (any, error){
	"Announce": guessAnnounce,
	"Undo":     guessUndo,
	"Follow":   guessFollow,
	"Accept":   guessAccept,
	"Reject":   guessReject,
	"Create":   guessCreateNote,
	"Update":   guessUpdateNote,
	"Delete":   guessDeleteNote,
}

func Guess(raw []byte) (report any, err error) {
	var (
		activity = Dict{
			"original activity": raw,
		}
		val any
		ok  bool
	)
	if err = json.Unmarshal(raw, &activity); err != nil {
		return nil, err
	}

	if val, ok = activity["type"]; !ok {
		return nil, ErrNoType
	}
	switch v := val.(type) {
	case string:
		// Special case
		if v == "Delete" && getString(activity, "actor") == getString(activity, "object") {
			// Waiting for https://github.com/mastodon/mastodon/pull/22273 to get rid of this branch
			log.Println("Somebody got deleted, scroll further.")
			return nil, nil
		}

		f, ok := guesserMap[v]
		if !ok {
			log.Printf("Ignoring unknown kind of activity: %s\n", raw)
			return nil, ErrUnknownType
		}

		log.Printf("Handling activity: %s\n", raw)
		return f(activity)
	default:
		return nil, ErrNoType
	}
}
