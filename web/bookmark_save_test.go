// SPDX-FileCopyrightText: 2026 Danila Gorelko
//
// SPDX-License-Identifier: AGPL-3.0-only

package web

import (
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"git.sr.ht/~bouncepaw/betula/db"
	"git.sr.ht/~bouncepaw/betula/types"
	"github.com/nalgeon/be"
)

func TestPostSaveBookmarkStripsPlaceholderDescription(t *testing.T) {
	type testCase struct {
		name           string
		description    string
		wantEmptyDescr bool
	}

	testCases := []testCase{
		{
			name:           "exact greater-than",
			description:    ">",
			wantEmptyDescr: true,
		},
		{
			name:           "leading space",
			description:    " >",
			wantEmptyDescr: false,
		},
		{
			name:           "trailing space",
			description:    "> ",
			wantEmptyDescr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db.InitInMemoryDB()

			form := url.Values{}
			form.Set("url", "https://example.com")
			form.Set("title", "Example")
			form.Set("visibility", "public")
			form.Set("description", tc.description)

			r := httptest.NewRequest("POST", "/save-link", strings.NewReader(form.Encode()))
			r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			w := httptest.NewRecorder()

			postSaveBookmark(w, r)

			res := w.Result()
			be.Equal(t, res.StatusCode, 303)

			bookmark, err := db.NewLocalBookmarksRepo().GetBookmarkByID(
				t.Context(),
				4, // TODO(d.gorelko): get rid of coupling to `InitInMemoryDB()`.
			)
			be.Err(t, err, nil)

			if tc.wantEmptyDescr {
				be.Equal(t, bookmark.Description, "")
			} else {
				be.Equal(t, bookmark.Description, tc.description)
			}
			be.Equal(t, bookmark.URL, "https://example.com")
			be.Equal(t, bookmark.Title, "Example")
			be.Equal(t, bookmark.Visibility, types.Public)
		})
	}
}
