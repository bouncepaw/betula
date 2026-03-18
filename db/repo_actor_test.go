// SPDX-FileCopyrightText: 2025 Danila Gorelko
// SPDX-FileCopyrightText: 2026 Danila Gorelko
//
// SPDX-License-Identifier: AGPL-3.0-only

package db

import (
	"testing"

	"git.sr.ht/~bouncepaw/betula/types"
	"github.com/nalgeon/be"
)

func TestGetRemoteBookmarks(t *testing.T) {
	InitInMemoryDB()
	actor1 := types.Actor{
		ID: "https://example.com/actor1",
	}
	actor2 := types.Actor{
		ID: "https://example.com/actor2",
	}
	StoreValidActor(actor1)
	StoreValidActor(actor2)
	AddPendingFollowing(actor1.ID)
	MarkAsSurelyFollowing(actor1.ID)
	bookmark1 := types.RemoteBookmark{
		ID:          "https://example.com/bookmark1",
		ActorID:     actor1.ID,
		Title:       "Bookmark from followed user",
		URL:         "https://example.com/1",
		PublishedAt: "2023-01-01T00:00:00Z",
	}
	bookmark2 := types.RemoteBookmark{
		ID:          "https://example.com/bookmark2",
		ActorID:     actor2.ID,
		Title:       "Bookmark from unfollowed user",
		URL:         "https://example.com/2",
		PublishedAt: "2023-01-02T00:00:00Z",
	}

	var repoRemoteBookmarks = NewRemoteBookmarkRepo()

	repoRemoteBookmarks.InsertRemoteBookmark(bookmark1)
	repoRemoteBookmarks.InsertRemoteBookmark(bookmark2)
	bookmarks, total := repoRemoteBookmarks.GetRemoteBookmarks(1)
	be.Equal(t, total, 1)
	be.Equal(t, len(bookmarks), 1)
	be.Equal(t, bookmarks[0].ID, bookmark1.ID)
	AddPendingFollowing(actor2.ID)
	MarkAsSurelyFollowing(actor2.ID)
	bookmarks, total = repoRemoteBookmarks.GetRemoteBookmarks(1)
	be.Equal(t, total, 2)
}

func TestGetRemoteBookmarks_Empty(t *testing.T) {
	InitInMemoryDB()

	var repoRemoteBookmarks = NewRemoteBookmarkRepo()
	bookmarks, total := repoRemoteBookmarks.GetRemoteBookmarks(1)
	be.Equal(t, total, 0)
	be.Equal(t, len(bookmarks), 0)
}
