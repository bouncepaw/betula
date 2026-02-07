// SPDX-FileCopyrightText: 2022-2025 Betula contributors
//
// SPDX-License-Identifier: AGPL-3.0-only

package fediverse

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log"
	"log/slog"
	"net/http"

	"git.sr.ht/~bouncepaw/betula/fediverse/activities"
	"git.sr.ht/~bouncepaw/betula/settings"
	"git.sr.ht/~bouncepaw/betula/types"
)

var (
	ErrNotBookmark = errors.New("fediverse: not a bookmark")
)

func fetchFedi(uri string) (*types.Bookmark, error) {
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", settings.UserAgent())
	req.Header.Set("Accept", types.OtherActivityType)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	var object activities.Dict
	if err := json.NewDecoder(io.LimitReader(resp.Body, 128_000)).Decode(&object); err != nil {
		return nil, err
	}

	bookmark, err := activities.RemoteBookmarkFromDict(object)
	if err != nil {
		return nil, err
	}
	log.Printf("tags %q\n%q\n", bookmark.Tags, object)

	return &types.Bookmark{
		Tags:        bookmark.Tags,
		URL:         bookmark.URL,
		Title:       bookmark.Title,
		Description: bookmark.DescriptionMycomarkup.String,
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
