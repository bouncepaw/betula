// SPDX-FileCopyrightText: 2024 Danila Gorelko
// SPDX-FileCopyrightText: 2024 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2026 Danila Gorelko
//
// SPDX-License-Identifier: AGPL-3.0-only

package db

import (
	"testing"

	"git.sr.ht/~bouncepaw/betula/types"
	"github.com/nalgeon/be"
)

func TestBookmarkCount(t *testing.T) {
	InitInMemoryDB()
	be.Equal(t, BookmarkCount(true), 2)
	be.Equal(t, BookmarkCount(false), 1)
}

func TestAddPost(t *testing.T) {
	InitInMemoryDB()
	post := types.Bookmark{
		CreationTime: "2023-03-18",
		Tags: []types.Tag{
			{Name: "cat"},
			{Name: "dog"},
		},
		URL:         "https://betula.mycorrhiza.wiki",
		Title:       "Betula",
		Description: "",
		Visibility:  types.Public,
	}
	InsertBookmark(post)
	be.Equal(t, BookmarkCount(true), 3)
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

	for _, tc := range cases {
		bookmarks, total := RandomBookmarks(tc.authorized, tc.n)
		be.Equal(t, len(bookmarks), int(total))
		creationTime := bookmarks[0].CreationTime
		for _, bookmark := range bookmarks[1:] {
			be.True(t, bookmark.CreationTime <= creationTime)
			creationTime = bookmark.CreationTime
		}
	}
}
