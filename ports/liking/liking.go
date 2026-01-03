// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package likingports

import (
	"database/sql"
	"encoding/json"
	"git.sr.ht/~bouncepaw/betula/types"
	"time"
)

type (
	LikeRepository interface {
		InsertLike(like LikeModel) error
		DeleteOurLikeOf(objectID string) error
		StatiFor(objectIDs []string) (map[string]LikeStatus, error)
	}
	LocalBookmarkRepository interface {
		Exists(id int) (bool, error)
	}
	RemoteBookmarkRepository interface {
		Exists(id string) (bool, error)
	}
)

type Service interface {
	LikeAnyBookmark(bookmarkID string) error
	UnlikeAnyBookmark(bookmarkID string) error

	FillLikes([]types.RenderedLocalBookmark, []types.RenderedRemoteBookmark) error
	// todo: locals too
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
