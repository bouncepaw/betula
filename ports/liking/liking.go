// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package likingports

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	apports "git.sr.ht/~bouncepaw/betula/ports/activitypub"
	"git.sr.ht/~bouncepaw/betula/types"
)

type (
	LikeRepository interface {
		InsertLike(ctx context.Context, like LikeModel) error
		DeleteOurLikeOf(ctx context.Context, objectID string) error
		DeleteLikeBy(ctx context.Context, likeID, actorID string) error
		StatiFor(ctx context.Context, objectIDs []string) (map[string]LikeStatus, error)
		LikedObjectForLike(
			ctx context.Context,
			likeID string,
		) (string, error)

		// ActorsThatLiked returns IDs of actors that liked the bookmark,
		// whether we liked it ourselves or an error.
		ActorsThatLiked(ctx context.Context, objectID string) ([]string, bool, error)
	}

	LikeCollectionRepository interface {
		UpsertLikeCollection(ctx context.Context, likeCollection LikeCollectionModel) error
		GetTotalItemsFor(ctx context.Context, objectID string) (int, error)
		IncrementIfPresent(ctx context.Context, objectID string) error
		DecrementIfPresent(ctx context.Context, objectID string) error
	}

	//nolint:interfacebloat // Temporary measure while migrating to the Repository pattern.
	LocalBookmarkRepository interface {
		Exists(context.Context, int) (bool, error)
		GetBookmarkByID(context.Context, int) (types.Bookmark, error)
		InsertBookmark(context.Context, types.Bookmark) (int64, error)
		GetBookmarkIDByURL(context.Context, string) (int, error)
		Bookmarks(ctx context.Context, authorized bool, page uint) ([]types.Bookmark, uint, error)
		BookmarksForDay(ctx context.Context, authorized bool, dayStamp string) ([]types.Bookmark, error)
		BookmarksWithTag(ctx context.Context, authorized bool, tagName string, page uint) ([]types.Bookmark, uint, error)
		RandomBookmarks(ctx context.Context, authorized bool, n uint) ([]types.Bookmark, uint, error)
		DeleteBookmark(ctx context.Context, id int) error
		EditBookmark(context.Context, types.Bookmark) error
		BookmarkCount(ctx context.Context, authorized bool) (uint, error)
	}
)

type Service interface {
	Like(ctx context.Context, bookmarkID string) error
	Unlike(ctx context.Context, bookmarkID string) error

	FillLikes(context.Context, []types.RenderedLocalBookmark, []types.RenderedRemoteBookmark) error

	ReceiveLike(context.Context, EventLike) error
	ReceiveUndoLike(context.Context, EventUndoLike) error
	ReceiveLikeCollection(context.Context, EventLikeCollectionSeen) error

	ActorsThatLiked(ctx context.Context, bookmarkID int) ([]apports.Actor, bool, error)
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

	EventLikeCollectionSeen struct {
		ID            *string
		Type          string
		TotalItems    int
		LikedObjectID string
		SourceJSON    json.RawMessage
	}
)
