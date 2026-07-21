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

	"github.com/nalgeon/be"

	"git.sr.ht/~bouncepaw/betula/db"
	"git.sr.ht/~bouncepaw/betula/types"
)

func TestRenderBookmarkIncludesRemarks(t *testing.T) {
	db.InitInMemoryDB()

	bm := types.Bookmark{
		URL:        "https://example.com",
		Title:      "Example",
		Visibility: types.Public,
	}
	id, err := localBookmarks.InsertBookmark(t.Context(), bm)
	be.Err(t, err, nil)
	bookmark, err := localBookmarks.GetBookmarkByID(t.Context(), int(id))
	be.Err(t, err, nil)

	ctrl.RepoRemarks = db.NewRemarksRepo()
	ctrl.RepoTags = db.NewTagsRepo()
	var (
		re1 = types.RemarkInfo{URL: "https://links.alice/1", Name: "Alice", Timestamp: time.Now()}
		re2 = types.RemarkInfo{URL: "https://links.bob/2", Name: "Bob", Timestamp: time.Now()}
	)
	be.Err(t, ctrl.RepoRemarks.SaveRemark(t.Context(), bookmark.ID, re1), nil)
	be.Err(t, ctrl.RepoRemarks.SaveRemark(t.Context(), bookmark.ID, re2), nil)

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	data := renderBookmark(bookmark, w, r, false)
	be.Equal(t, len(data.Remarks), 2)
}
