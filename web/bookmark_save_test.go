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
			if res.StatusCode != 303 {
				t.Fatalf("expected status 303, got %d", res.StatusCode)
			}

			bookmark, err := db.NewLocalBookmarksRepo().GetBookmarkByID(
				t.Context(),
				4, // TODO(d.gorelko): get rid of coupling to `InitInMemoryDB()`.
			)
			if err != nil {
				t.Fatalf("bookmark not found after insert")
			}

			if tc.wantEmptyDescr {
				if bookmark.Description != "" {
					t.Fatalf("expected empty description, got %q", bookmark.Description)
				}
			} else {
				if bookmark.Description != tc.description {
					t.Fatalf("expected description %q to be preserved, got %q", tc.description, bookmark.Description)
				}
			}
			if bookmark.URL != "https://example.com" {
				t.Fatalf("unexpected URL: %q", bookmark.URL)
			}
			if bookmark.Title != "Example" {
				t.Fatalf("unexpected title: %q", bookmark.Title)
			}
			if bookmark.Visibility != types.Public {
				t.Fatalf("unexpected visibility: %v", bookmark.Visibility)
			}
		})
	}
}
