// SPDX-FileCopyrightText: 2025 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package likingports

import (
	"database/sql"
	"encoding/json"
	"git.sr.ht/~bouncepaw/betula/fediverse/activities"
	"time"
)

type (
	LikeRepository interface {
		InsertLike(like LikeModel) error
		UpsertLikeCollection(likeCollection LikeCollectionModel) error
	}
	LocalBookmarkRepository interface {
		Exists(id int) (bool, error)
	}
	RemoteBookmarkRepository interface {
		Exists(string int) (bool, error)
	}
)

type Service interface {
	HandleIncomingLikeActivity(report activities.LikeReport) error
	HandleIncomingUpdateNoteActivity(report activities.UpdateNoteReport) error
	LikeLocalBookmark(bookmarkID int) error
	LikeRemoteBookmark(bookmarkID string) error
	CountLikesForLocalBookmarks(bookmarkIDs []int) ([]int, error)
	CountLikesForRemoteBookmarks(bookmarkIDs []string) ([]int, error)
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
)

func (m LikeModel) SavedAt() (time.Time, error) {
	return time.Parse(time.DateTime, m.SerializedSavedAt.String)
}
