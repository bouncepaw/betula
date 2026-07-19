// SPDX-FileCopyrightText: 2023 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2024 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2026 Danila Gorelko
// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package apports

import (
	"encoding/json"
	"errors"

	"git.sr.ht/~bouncepaw/betula/types"
)

type (
	AcceptReport struct {
		ActorID  string
		ObjectID string
		Object   Dict
	}

	RejectReport struct {
		ActorID  string
		ObjectID string
		Object   Dict
	}

	AnnounceReport struct {
		ActorID    string
		AnnounceID string // id of the repost
		ObjectID   string // object that was reposted
	}

	FollowReport struct {
		ActorID          string
		ObjectID         string
		OriginalActivity Dict
	}

	// LikeReport reports that actor with ActorID liked the object with ObjectID.
	LikeReport struct {
		ID       string
		ActorID  string
		ObjectID string
		Activity json.RawMessage
	}

	UndoAnnounceReport struct {
		AnnounceReport
	}

	UndoFollowReport struct {
		FollowReport
	}

	UndoLikeReport struct {
		ID       string
		Object   LikeReport
		Activity json.RawMessage
	}

	CreateNoteReport struct {
		Bookmark        types.RemoteBookmark
		LikesCollection *Collection
	}

	UpdateNoteReport struct {
		Bookmark        types.RemoteBookmark
		ActorID         string
		LikesCollection *Collection
	}

	DeleteNoteReport struct {
		ActorID    string
		BookmarkID string
	}
)

var (
	errLikeNoID     = errors.New("apports: like has no id")
	errLikeNoActor  = errors.New("apports: like has no actor")
	errLikeNoObject = errors.New("apports: like has no object")
)

func (lr LikeReport) Valid() error {
	switch {
	case lr.ID == "":
		return errLikeNoID
	case lr.ActorID == "":
		return errLikeNoActor
	case lr.ObjectID == "":
		return errLikeNoObject
	default:
		return nil
	}
}

type Collection struct {
	ID         *string `json:"id"`
	Type       string  `json:"type"`
	TotalItems int     `json:"totalItems"`
	// No Items.
}

func (c Collection) Valid() error {
	// Empty ID allowed.
	switch {
	case c.Type != "Collection" && c.Type != "OrderedCollection":
		return errors.New("invalid collection type")
	case c.TotalItems < 0:
		return errors.New("sub-zero total items")
	default:
		return nil
	}
}
