// SPDX-FileCopyrightText: 2025 Danila Gorelko
// SPDX-FileCopyrightText: 2026 Danila Gorelko
//
// SPDX-License-Identifier: AGPL-3.0-only

package db

import (
	"testing"

	"github.com/nalgeon/be"

	"git.sr.ht/~bouncepaw/betula/types"
)

func TestGetRemoteBookmarks(t *testing.T) {
	InitInMemoryDB()
	ctx := t.Context()
	actorRepo := NewActorRepo()
	actor1 := validActor("https://example.com/actor1", "actor1")
	actor2 := validActor("https://example.com/actor2", "actor2")
	be.Err(t, actorRepo.StoreActor(ctx, actor1), nil)
	be.Err(t, actorRepo.StoreActor(ctx, actor2), nil)
	be.Err(t, actorRepo.AddPendingFollowing(ctx, actor1.ID), nil)
	be.Err(t, actorRepo.MarkAsSurelyFollowing(ctx, actor1.ID), nil)
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
	be.Err(t, actorRepo.AddPendingFollowing(ctx, actor2.ID), nil)
	be.Err(t, actorRepo.MarkAsSurelyFollowing(ctx, actor2.ID), nil)
	bookmarks, total = repoRemoteBookmarks.GetRemoteBookmarks(1)
	be.Equal(t, total, 2)
}

func validActor(id, username string) types.Actor {
	a := types.Actor{
		ID:                id,
		Inbox:             id + "/inbox",
		PreferredUsername: username,
		DisplayedName:     username,
		Domain:            "example.com",
	}
	a.PublicKey.ID = id + "#main-key"
	a.PublicKey.Owner = id
	a.PublicKey.PublicKeyPEM = "PEM"
	return a
}

func TestGetRemoteBookmarks_Empty(t *testing.T) {
	InitInMemoryDB()

	var repoRemoteBookmarks = NewRemoteBookmarkRepo()
	bookmarks, total := repoRemoteBookmarks.GetRemoteBookmarks(1)
	be.Equal(t, total, 0)
	be.Equal(t, len(bookmarks), 0)
}
