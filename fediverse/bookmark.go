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
	"git.sr.ht/~bouncepaw/betula/types"
)

var (
	ErrNotBookmark = errors.New("fediverse: not a bookmark")
)

func fetchFedi(uri string) (*types.RemoteBookmark, error) {
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", types.OtherActivityType)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	var object activities.Dict
	if err := json.NewDecoder(io.LimitReader(resp.Body, 128_000)).Decode(&object); err != nil {
		return nil, err
	}

	return activities.NoteFromDict(object)
}

// FetchBookmark fetches a bookmark on the given address somehow. First, it tries to get a Note ActivityPub object formatted with Betula rules. If it fails to do so, it resorts to the readpage method.
func FetchBookmark(uri string) (*types.RemoteBookmark, error) {
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

	return &types.RemoteBookmark{
		ID:              uri,
		RepostOf:        sql.NullString{},
		ActorID:         "", // How to find...
		Title:           foundData.PostName,
		URL:             foundData.BookmarkOf,
		DescriptionHTML: "", // We don't fetch that...
		DescriptionMycomarkup: sql.NullString{
			String: foundData.Mycomarkup,
			Valid:  foundData.Mycomarkup != "",
		},
		PublishedAt: "",               // We don't fetch that!?
		UpdatedAt:   sql.NullString{}, // And that too...
		Activity:    nil,              // of course it's nil
		Tags:        types.TagsFromStringSlice(foundData.Tags),
	}, nil
}
