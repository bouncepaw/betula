// SPDX-FileCopyrightText: 2022-2025 Betula contributors
//
// SPDX-License-Identifier: AGPL-3.0-only

package fediverse

import (
	"database/sql"
	"encoding/json"
	"errors"
	"git.sr.ht/~bouncepaw/betula/readpage"
	"io"
	"log"
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

// FetchBookmarkAsRepost fetches a bookmark on the given address somehow. First, it tries to get a Note ActivityPub object formatted with Betula rules. If it fails to do so, it resorts to the readpage method.
func FetchBookmarkAsRepost(uri string) (*types.Bookmark, error) {
	log.Printf("Fetching remote bookmark from %s\n", uri)
	bookmark, err := fetchFedi(uri)
	if err != nil {
		log.Printf("Tried to fetch a remote bookmark from %s, failed with: %s. Falling back to microformats\n", uri, err)
		// no return
	} else {
		log.Printf("Fetched a remote bookmark from %s\n", uri)
		return bookmark, nil
	}

	foundData, err := readpage.FindDataForMyRepost(uri)
	if err != nil {
		return nil, err
	} else if foundData.IsHFeed || foundData.BookmarkOf == "" || foundData.PostName == "" {
		return nil, ErrNotBookmark
	}

	log.Printf("Fetched a remote bookmark from %s with readpage\n", uri)
	return &types.Bookmark{
		Tags:           types.TagsFromStringSlice(foundData.Tags),
		URL:            foundData.BookmarkOf,
		Title:          foundData.PostName,
		Description:    foundData.Mycomarkup,
		Visibility:     types.Public,
		RepostOf:       &uri,             // TODO: transitive reposts are a thing...
		OriginalAuthor: sql.NullString{}, // actors are found only in activities
	}, nil
}
