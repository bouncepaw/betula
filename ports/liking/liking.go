// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package likingports

import (
	"context"
	"database/sql"
	"encoding/json"
	"git.sr.ht/~bouncepaw/betula/types"
	"time"
)

type (
	LikeRepository interface {
		InsertLike(ctx context.Context, like LikeModel) error
		DeleteOurLikeOf(ctx context.Context, objectID string) error
		DeleteLikeBy(ctx context.Context, likeID, actorID string) error
		StatiFor(ctx context.Context, objectIDs []string) (map[string]LikeStatus, error)
	}
	LocalBookmarkRepository interface {
		Exists(ctx context.Context, id int) (bool, error)
	}
)

type Service interface {
	Like(ctx context.Context, bookmarkID string) error
	Unlike(ctx context.Context, bookmarkID string) error

	FillLikes(context.Context, []types.RenderedLocalBookmark, []types.RenderedRemoteBookmark) error

	ReceiveLike(context.Context, EventLike) error
	ReceiveUndoLike(context.Context, EventUndoLike) error
}

type (
	LikeModel struct {
		ID                sql.NullString
		ActorID           sql.NullString
		ObjectID          string
		SerializedSavedAt sql.NullString
		SourceJSON        json.RawMessage
	}

	LikeCollectionModel struct {
		ID            sql.NullString
		LikedObjectID string
		TotalItems    int
		SourceJSON    json.RawMessage
	}

	LikeStatus struct {
		Count     int
		LikedByUs bool
	}
)

func (m LikeModel) SavedAt() (time.Time, error) {
	return time.Parse(time.DateTime, m.SerializedSavedAt.String)
}

type (
	EventLike struct {
		LikeID        string
		ActorID       string
		LikedObjectID string
		Activity      json.RawMessage
	}

	EventUndoLike struct {
		UndoLikeID string
		ActorID    string
		LikeID     string
		Activity   json.RawMessage
	}
)
