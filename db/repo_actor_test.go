// SPDX-FileCopyrightText: 2025 Danila Gorelko
//
// SPDX-License-Identifier: AGPL-3.0-only

package db

import (
	"testing"

	"git.sr.ht/~bouncepaw/betula/types"
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
	InsertRemoteBookmark(bookmark1)
	InsertRemoteBookmark(bookmark2)
	bookmarks, total := GetRemoteBookmarks(1)
	if total != 1 {
		t.Errorf("Expected 1 bookmark from followed users, got %d", total)
	}
	if len(bookmarks) != 1 {
		t.Fatalf("Expected 1 bookmark in the result, got %d", len(bookmarks))
	}
	if bookmarks[0].ID != bookmark1.ID {
		t.Errorf("Expected bookmark ID %s, got %s", bookmark1.ID, bookmarks[0].ID)
	}
	AddPendingFollowing(actor2.ID)
	MarkAsSurelyFollowing(actor2.ID)
	bookmarks, total = GetRemoteBookmarks(1)
	if total != 2 {
		t.Errorf("Expected 2 bookmarks after following second user, got %d", total)
	}
}

func TestGetRemoteBookmarks_Empty(t *testing.T) {
	InitInMemoryDB()
	bookmarks, total := GetRemoteBookmarks(1)
	if total != 0 {
		t.Errorf("Expected 0 bookmarks with no followed users, got %d", total)
	}
	if len(bookmarks) != 0 {
		t.Errorf("Expected empty bookmarks slice, got %d items", len(bookmarks))
	}
}
