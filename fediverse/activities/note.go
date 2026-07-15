// SPDX-FileCopyrightText: 2024 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2026 Danila Gorelko
// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package activities

import (
	"database/sql"
	"html/template"
	"slices"
	"strings"

	"git.sr.ht/~bouncepaw/betula/pkg/bxstr"
	"git.sr.ht/~bouncepaw/betula/types"
)

type (
	CreateNoteReport struct {
		Bookmark        types.RemoteBookmark
		LikesCollection *Collection
	}

	UpdateNoteReport struct {
		Bookmark        types.RemoteBookmark
		ActorID         string
		LikesCollection *Collection
	}

	DeleteNoteReport struct {
		ActorID    string
		BookmarkID string
	}
)

func RemoteBookmarkFromDict(object Dict) (note *types.RemoteBookmark, err error) {
	if typ := getString(object, "type"); typ != "Note" && typ != "Page" && typ != "Article" {
		return nil, ErrNotNote
	}
	bookmark := types.RemoteBookmark{
		// Invariants
		RepostOf: sql.NullString{},

		// Required fields
		ID:              getString(object, "id"),
		ActorID:         getString(object, "attributedTo"),
		Title:           getString(object, "name"),
		DescriptionHTML: template.HTML(getString(object, "content")),
		PublishedAt:     getTime(object, "published"),

		// Optional fields
		UpdatedAt: sql.NullString{},
		Source:    sql.NullString{},
		Tags:      nil,
	}

	if updated := getTime(object, "updated"); updated != "" {
		bookmark.UpdatedAt = sql.NullString{
			String: updated,
			Valid:  true,
		}
	}

	// Grabbing URL
	attachments, ok := object["attachment"].([]any)
	if !ok {
		return nil, ErrEmptyField
	}
	for _, rawamnt := range attachments {
		amnt, ok := rawamnt.(Dict)
		if !ok {
			continue
		}
		if getString(amnt, "type") == "Link" {
			if href := getString(amnt, "href"); bxstr.IsValidURL(href) {
				bookmark.URL = href
				break
			}
		}
	}

	// Lie detector
	if !bxstr.SameHost(bookmark.ActorID, bookmark.ID) {
		return nil, ErrHostMismatch
	}

	// Verify required fields.
	mustBeNonEmpty := []string{bookmark.ID, bookmark.ActorID, bookmark.Title, bookmark.PublishedAt, bookmark.URL}
	if slices.Contains(mustBeNonEmpty, "") {
		return nil, ErrEmptyField
	}

	// Grabbing the source text
	source, ok := object["source"].(Dict)
	if ok {
		switch types.SourceType(getString(source, "mediaType")) {
		case types.SourceMycomarkup:
			bookmark.SourceType = types.SourceMycomarkup
			bookmark.Source = sql.NullString{String: getString(source, "content"), Valid: true}
		case types.SourcePlainText:
			bookmark.SourceType = types.SourcePlainText
			bookmark.Source = sql.NullString{String: getString(source, "content"), Valid: true}
		}
	}

	// Collecting tags
	tags, ok := object["tag"].([]any)
	for _, anytag := range tags {
		tag, ok := anytag.(Dict)
		if !ok {
			continue
		}
		typ := getString(tag, "type")
		if typ != "Hashtag" {
			continue
		}

		name := strings.TrimPrefix(getString(tag, "name"), "#")
		bookmark.Tags = append(bookmark.Tags, types.Tag{
			Name: name,
			// Rest of struct not needed
		})
	}

	return &bookmark, nil
}

func guessCreateNote(activity Dict) (report any, err error) {
	object, ok := activity["object"].(Dict)
	if !ok {
		return nil, ErrNoObject
	}

	bookmark, err := RemoteBookmarkFromDict(object)
	if err != nil {
		return nil, err
	}
	bookmark.Activity = activity["original activity"].([]byte)

	cnr := CreateNoteReport{
		Bookmark: *bookmark,
	}

	if object["likes"] != nil {
		switch likesCollection := object["likes"].(type) {
		case string: // Don't care, not fetching.
		case Dict: // Now we're talking!
			cnr.LikesCollection, err = collectionFromDict(likesCollection)
			if err != nil {
				return nil, err
			}
		}
	}
	return cnr, nil
}

func guessUpdateNote(activity Dict) (report any, err error) {
	object, ok := activity["object"].(Dict)
	if !ok {
		return nil, ErrNoObject
	}

	bookmark, err := RemoteBookmarkFromDict(object)
	if err != nil {
		return nil, err
	}
	bookmark.Activity = activity["original activity"].([]byte)

	unr := UpdateNoteReport{
		ActorID:  getIDSomehow(activity, "actor"),
		Bookmark: *bookmark,
	}
	if unr.ActorID == "" {
		return nil, ErrNoActor
	}

	if object["likes"] != nil {
		switch likesCollection := object["likes"].(type) {
		case string: // Don't care, not fetching.
		case Dict: // Now we're talking!
			unr.LikesCollection, err = collectionFromDict(likesCollection)
			if err != nil {
				return nil, err
			}
		}
	}

	return unr, nil
}

func guessDeleteNote(activity Dict) (report any, err error) {
	deletion := DeleteNoteReport{
		ActorID:    getIDSomehow(activity, "actor"),
		BookmarkID: getIDSomehow(activity, "object"),
	}
	if !bxstr.SameHost(deletion.ActorID, deletion.BookmarkID) {
		return nil, ErrHostMismatch
	}
	return deletion, nil
}
