// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package apports

import (
	"git.sr.ht/~bouncepaw/betula/types"
)

type (
	Guesser interface {
		Guess(raw []byte) (any, error)
	}

	NoteParser interface {
		BookmarkFromNote(object Dict) (*types.RemoteBookmark, error)
		GuessCreateNote(activity Dict) (any, error)
		GuessUpdateNote(activity Dict) (any, error)
		GuessDeleteNote(activity Dict) (any, error)
	}

	FollowParser interface {
		GuessFollow(activity Dict) (any, error)
		GuessAccept(activity Dict) (any, error)
		GuessReject(activity Dict) (any, error)
		GuessUndoFollow(object Dict) (any, error)
	}

	LikeParser interface {
		GuessLike(activity Dict) (any, error)
		GuessUndoLike(activity, object Dict) (any, error)
	}

	AnnounceParser interface {
		GuessAnnounce(activity Dict) (any, error)
		GuessUndoAnnounce(object Dict) (any, error)
	}
)
