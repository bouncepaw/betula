// SPDX-FileCopyrightText: 2024 Danila Gorelko
// SPDX-FileCopyrightText: 2024 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2026 Danila Gorelko
//
// SPDX-License-Identifier: AGPL-3.0-only

package db

import (
	"testing"

	"github.com/nalgeon/be"

	"git.sr.ht/~bouncepaw/betula/types"
)

func TestBookmarkCount(t *testing.T) {
	InitInMemoryDB()
	repo := NewLocalBookmarksRepo()

	count, err := repo.BookmarkCount(t.Context(), true)
	be.Err(t, err, nil)
	be.Equal(t, count, 2)

	count, err = repo.BookmarkCount(t.Context(), false)
	be.Err(t, err, nil)
	be.Equal(t, count, 1)
}

func TestAddPost(t *testing.T) {
	InitInMemoryDB()
	post := types.Bookmark{
		CreationTime: "2023-03-18",
		Tags: []types.Tag{
			{Name: "cat"},
			{Name: "dog"},
		},
		URL:         "https://joinbetula.org",
		Title:       "Betula",
		Description: "",
		Visibility:  types.Public,
	}
	repo := NewLocalBookmarksRepo()
	_, err := repo.InsertBookmark(t.Context(), post)
	be.Err(t, err, nil)
	count, err := repo.BookmarkCount(t.Context(), true)
	be.Err(t, err, nil)
	be.Equal(t, count, 3)
}

func TestRandomBookmarks(t *testing.T) {
	InitInMemoryDB()
	MoreTestingBookmarks()

	cases := []struct {
		authorized bool
		n          uint
	}{
		{true, 20},
		{false, 20},
	}

	repo := NewLocalBookmarksRepo()
	for _, tc := range cases {
		bookmarks, total, err := repo.RandomBookmarks(t.Context(), tc.authorized, tc.n)
		be.Err(t, err, nil)
		be.Equal(t, len(bookmarks), int(total))
		creationTime := bookmarks[0].CreationTime
		for _, bookmark := range bookmarks[1:] {
			be.True(t, bookmark.CreationTime <= creationTime)
			creationTime = bookmark.CreationTime
		}
	}
}
