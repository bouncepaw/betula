// SPDX-FileCopyrightText: 2023 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2024 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package parsing

import (
	"encoding/json"
	"log/slog"

	apports "git.sr.ht/~bouncepaw/betula/ports/activitypub"
)

type Guesser struct {
	guesserMap map[string]func(apports.Dict) (any, error)

	noteParser     apports.NoteParser
	followParser   apports.FollowParser
	likeParser     apports.LikeParser
	announceParser apports.AnnounceParser
}

var _ apports.Guesser = (*Guesser)(nil)

func NewGuesser(siteURLFn func() string) *Guesser {
	g := &Guesser{
		noteParser:     NewNoteParser(siteURLFn),
		followParser:   NewFollowParser(),
		likeParser:     NewLikeParser(),
		announceParser: NewAnnounceParser(),
	}
	g.guesserMap = map[string]func(apports.Dict) (any, error){
		"Announce": g.announceParser.GuessAnnounce,
		"Undo":     g.guessUndo,
		"Follow":   g.followParser.GuessFollow,
		"Accept":   g.followParser.GuessAccept,
		"Reject":   g.followParser.GuessReject,
		"Create":   g.noteParser.GuessCreateNote,
		"Update":   g.noteParser.GuessUpdateNote,
		"Delete":   g.noteParser.GuessDeleteNote,
		"Like":     g.likeParser.GuessLike,
	}
	return g
}

func (g *Guesser) guessUndo(activity apports.Dict) (any, error) {
	objectMap, ok := activity["object"].(apports.Dict)
	if !ok {
		return nil, ErrNoObject
	}

	switch objectMap["type"] {
	case "Announce":
		return g.announceParser.GuessUndoAnnounce(objectMap)
	case "Follow":
		return g.followParser.GuessUndoFollow(objectMap)
	case "Like":
		return g.likeParser.GuessUndoLike(activity, objectMap)
	default:
		return nil, ErrUnknownType
	}
}

func (g *Guesser) Guess(raw []byte) (report any, err error) {
	var (
		activity = apports.Dict{
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

		f, ok := g.guesserMap[v]
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
