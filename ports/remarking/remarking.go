// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package remarkingports

import (
	"context"

	"git.sr.ht/~bouncepaw/betula/types"
)

type (
	Service interface {
		ReceiveLegacyRemark(context.Context, EventLegacyRemark) error
		ReceiveLegacyUnremark(context.Context, EventLegacyUnremark) error
	}

	Repository interface {
		// RemarksOf returns all remarks known about the specified bookmark.
		RemarksOf(ctx context.Context, bookmarkID int) ([]types.RepostInfo, error)
		SaveRemark(ctx context.Context, bookmarkID int, remark types.RepostInfo) error
		DeleteRemark(ctx context.Context, bookmarkID int, remarkURL string) error
	}

	EventLegacyRemark struct {
		ActorID        string
		AnnounceID     string // id of the repost
		ObjectID       string // object that was reposted
		SourceActivity []byte // raw JSON of the Announce activity
	}
	EventLegacyUnremark struct {
		ActorID    string
		AnnounceID string // id of the repost
		ObjectID   string // object that was reposted
	}
)
