// SPDX-FileCopyrightText: 2024 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2024 arne
// SPDX-FileCopyrightText: 2026 Danila Gorelko
// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package fediverse

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	apports "git.sr.ht/~bouncepaw/betula/ports/activitypub"
	"git.sr.ht/~bouncepaw/betula/settings"
	"git.sr.ht/~bouncepaw/betula/types"
)

var (
	ErrNotBookmark = errors.New("fediverse: not a bookmark")
)

func fetchFedi(uri string) (*types.Bookmark, error) {
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", settings.UserAgent())
	req.Header.Set("Accept", types.OtherActivityType)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var object apports.Dict
	if err := json.NewDecoder(io.LimitReader(resp.Body, 128_000)).Decode(&object); err != nil {
		return nil, err
	}

	bookmark, err := noteParser.BookmarkFromNote(object)
	if err != nil {
		return nil, err
	}
	slog.Debug("Fetched remote bookmark tags", "tags", bookmark.Tags, "object", object)

	return &types.Bookmark{
		Tags:        bookmark.Tags,
		URL:         bookmark.URL,
		Title:       bookmark.Title,
		Description: bookmark.Source.String,
		Visibility:  types.Public,
		RepostOf:    &uri,
		OriginalAuthor: sql.NullString{
			String: bookmark.ActorID,
			Valid:  true,
		},
	}, nil
}

// FetchBookmarkAsRepost fetches a bookmark on the given address.
func FetchBookmarkAsRepost(uri string) (*types.Bookmark, error) {
	bookmark, err := fetchFedi(uri)
	if err != nil {
		slog.Error("Failed to fetch bookmark for repost", "uri", uri, "err", err)
		return nil, err
	}

	slog.Info("Fetched bookmark for repost", "uri", uri, "bookmark", bookmark)
	return bookmark, nil
	// NOTE: IndieWeb-style reposts are no longer supported.
}
