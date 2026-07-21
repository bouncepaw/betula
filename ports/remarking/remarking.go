// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package remarkingports

import (
	"context"
	"fmt"

	"git.sr.ht/~bouncepaw/betula/types"
)

var (
	ErrNotRemark      = fmt.Errorf("not a remark")
	ErrRemarkOfRemote = fmt.Errorf("remark of remote")
)

type (
	Service interface {
		// BroadcastCreateRemark broadcasts the remark to followers and the author
		// of the remarked bookmark. Fails if the argument is not a remark.
		BroadcastCreateRemark(context.Context, types.Bookmark) error

		// ReceiveCreateRemark saves information about the remark being a remark
		// of a bookmark of ours. Fails if it's not the case.
		ReceiveCreateRemark(context.Context, EventCreateRemark) error
		ReceiveUpdateRemark(context.Context, EventUpdateRemark) error
		ReceiveDeleteRemark(context.Context, EventDeleteRemark) error

		ReceiveLegacyRemark(context.Context, EventLegacyRemark) error
		ReceiveLegacyUnremark(context.Context, EventLegacyUnremark) error
	}

	Repository interface {
		// RemarksOf returns all remarks known about the specified bookmark.
		RemarksOf(ctx context.Context, bookmarkID int) ([]types.RemarkInfo, error)
		SaveRemark(ctx context.Context, bookmarkID int, remark types.RemarkInfo) error
		DeleteRemark(ctx context.Context, bookmarkID int, remarkURL string) error
	}

	EventCreateRemark struct {
		Bookmark types.RemoteBookmark
	}
	EventUpdateRemark struct {
		Bookmark types.RemoteBookmark
	}
	EventDeleteRemark struct {
		RemarkID   string
		BookmarkID string
	}
	EventLegacyRemark struct {
		ActorID        string
		AnnounceID     string // id of the remark
		ObjectID       string // object that was remarked
		SourceActivity []byte // raw JSON of the Announce activity
	}
	EventLegacyUnremark struct {
		ActorID    string
		AnnounceID string // id of the remark
		ObjectID   string // object that was remarked
	}
)
