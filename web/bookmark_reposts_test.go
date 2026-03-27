// SPDX-FileCopyrightText: 2025 Danila Gorelko
// SPDX-FileCopyrightText: 2026 Danila Gorelko
// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"git.sr.ht/~bouncepaw/betula/db"
	"git.sr.ht/~bouncepaw/betula/types"
	"github.com/nalgeon/be"
)

func TestRenderBookmarkIncludesReposts(t *testing.T) {
	db.InitInMemoryDB()

	bm := types.Bookmark{
		URL:        "https://example.com",
		Title:      "Example",
		Visibility: types.Public,
	}
	id := db.InsertBookmark(bm)
	bookmark, found := db.GetBookmarkByID(int(id))
	be.True(t, found)

	db.SaveRepost(bookmark.ID, types.RepostInfo{URL: "https://links.alice/1", Name: "Alice", Timestamp: time.Now()})
	db.SaveRepost(bookmark.ID, types.RepostInfo{URL: "https://links.bob/2", Name: "Bob", Timestamp: time.Now()})

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	data := renderBookmark(bookmark, w, r, false)
	be.Equal(t, len(data.Reposts), 2)
}
