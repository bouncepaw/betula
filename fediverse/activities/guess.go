// SPDX-FileCopyrightText: 2023 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2024 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package activities

import (
	"encoding/json"
	"log/slog"

	"git.sr.ht/~bouncepaw/betula/svc/activitypub/parsing"
)

var (
	noteParser     = parsing.NewNoteParser()
	followParser   = parsing.NewFollowParser()
	likeParser     = parsing.NewLikeParser()
	announceParser = parsing.NewAnnounceParser()
)

var guesserMap = map[string]func(Dict) (any, error){
	"Announce": announceParser.GuessAnnounce,
	"Undo":     guessUndo,
	"Follow":   followParser.GuessFollow,
	"Accept":   followParser.GuessAccept,
	"Reject":   followParser.GuessReject,
	"Create":   noteParser.GuessCreateNote,
	"Update":   noteParser.GuessUpdateNote,
	"Delete":   noteParser.GuessDeleteNote,
	"Like":     likeParser.GuessLike,
}

func guessUndo(activity Dict) (any, error) {
	objectMap, ok := activity["object"].(Dict)
	if !ok {
		return nil, ErrNoObject
	}

	switch objectMap["type"] {
	case "Announce":
		return announceParser.GuessUndoAnnounce(objectMap)
	case "Follow":
		return followParser.GuessUndoFollow(objectMap)
	case "Like":
		return likeParser.GuessUndoLike(activity, objectMap)
	default:
		return nil, ErrUnknownType
	}
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
			slog.Info("Somebody got deleted, scroll further")
			return nil, nil
		}

		f, ok := guesserMap[v]
		if !ok {
			slog.Info("Ignoring unknown kind of activity", "raw", json.RawMessage(raw))
			return nil, ErrUnknownType
		}

		slog.Info("Handling activity", "raw", json.RawMessage(raw))
		return f(activity)
	default:
		return nil, ErrNoType
	}
}
