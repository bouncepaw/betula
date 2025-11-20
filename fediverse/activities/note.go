// SPDX-FileCopyrightText: 2022-2025 Betula contributors
//
// SPDX-License-Identifier: AGPL-3.0-only

package activities

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"git.sr.ht/~bouncepaw/betula/stricks"
	"golang.org/x/net/html"
	"html/template"
	"strings"
	"time"

	"git.sr.ht/~bouncepaw/betula/myco"
	"git.sr.ht/~bouncepaw/betula/settings"
	"git.sr.ht/~bouncepaw/betula/types"
)

func DeleteNote(postId int) ([]byte, error) {
	id := fmt.Sprintf("%s/%d", settings.SiteURL(), postId)
	activity := Dict{
		"@context": atContext,
		"type":     "Delete",
		"actor":    betulaActor,
		"to": []string{
			publicAudience,
			fmt.Sprintf("%s/followers", settings.SiteURL()),
		},
		"id":     id + "?delete",
		"object": id,
	}
	return json.Marshal(activity)
}

func NoteFromBookmark(bookmark types.Bookmark) (Dict, error) {
	if bookmark.ID == 0 {
		return nil, errors.New("an empty ID was passed")
	}
	// Generating the timestamp
	t, err := time.Parse(types.TimeLayout, bookmark.CreationTime)
	if err != nil {
		return nil, err
	}
	published := t.Format(time.RFC3339)

	// Generating the contents
	var content strings.Builder
	var tags []any

	content.WriteString(fmt.Sprintf(`<h3><a href="%s"'>%s</a></h3>`, html.EscapeString(bookmark.URL), html.EscapeString(bookmark.Title)))
	content.WriteString(string(myco.MarkupToHTML(bookmark.Description)))
	if len(bookmark.Tags) > 0 {
		content.WriteString("<p>")
		for i, tag := range bookmark.Tags {
			if tag.Name == "" {
				continue
			}
			if i > 0 {
				content.WriteString(", ")
			}
			/*
				Here's an example of a real tag markup from a real Mastodon activity:

				<a href="https://merveilles.town/tags/FediDev" class="mention hashtag" rel="tag">#<span>FediDev</span></a>

				Copying the markup so the tags look nice in Mastodon.
			*/
			content.WriteString(
				fmt.Sprintf(`<a href="%s/tag/%s" class="mention hashtag" rel="tag">#<span>%s</span></a>`,
					settings.SiteURL(),
					tag.Name,
					tag.Name,
				))

			// https://docs.joinmastodon.org/spec/activitypub/#Hashtag
			tags = append(tags, Dict{
				"type": "Hashtag",
				"name": "#" + tag.Name, // The # is needed
				"href": fmt.Sprintf("%s/tag/%s", settings.SiteURL(), tag.Name),
			})
		}
		content.WriteString("</p>")
	}

	object := Dict{
		"@context": []any{
			atContext,
			Dict{
				"Hashtag": "https://www.w3.org/ns/activitystreams#Hashtag",
			},
		},
		"type":         "Note",
		"id":           fmt.Sprintf("%s/%d", settings.SiteURL(), bookmark.ID),
		"actor":        betulaActor,
		"attributedTo": betulaActor,
		"to": []string{
			publicAudience,
			fmt.Sprintf("%s/followers", settings.SiteURL()),
		},
		"content": strings.ReplaceAll(
			strings.ReplaceAll(content.String(), "\t", ""),
			">\n", ">"),
		"name": bookmark.Title,
		"attachment": []Dict{
			{ // Lemmy-style.
				"href": bookmark.URL,
				"type": "Link",
			},
		},
		"source": map[string]string{
			// Misskey-style. They put text/x.misskeymarkdown though.
			"content":   bookmark.Description,
			"mediaType": "text/mycomarkup",
		},
		"published": published,
	}

	if len(tags) > 0 {
		object["tag"] = tags
	}
	return object, nil
}

func makeNoteAction(bookmark types.Bookmark) (Dict, error) {
	object, err := NoteFromBookmark(bookmark)
	if err != nil {
		return nil, err
	}

	delete(object, "@context")

	activity := Dict{
		"@context": []any{
			atContext,
			Dict{
				"Hashtag": "https://www.w3.org/ns/activitystreams#Hashtag",
			},
		},
		"actor":  betulaActor,
		"object": object,
	}
	return activity, nil
}

func CreateNote(post types.Bookmark) ([]byte, error) {
	activity, err := makeNoteAction(post)
	if err != nil {
		return nil, err
	}
	activity["type"] = "Create"
	activity["id"] = fmt.Sprintf("%s/%d?create", settings.SiteURL(), post.ID)
	return json.Marshal(activity)
}

func UpdateNote(post types.Bookmark) ([]byte, error) {
	activity, err := makeNoteAction(post)
	if err != nil {
		return nil, err
	}
	activity["type"] = "Update"
	activity["id"] = fmt.Sprintf("%s/%d?update", settings.SiteURL(), post.ID)
	activity["object"].(Dict)["updated"] = time.Now().Format(time.RFC3339)
	return json.Marshal(activity)
}

type CreateNoteReport struct {
	Bookmark types.RemoteBookmark
}
type UpdateNoteReport struct {
	Bookmark types.RemoteBookmark
}
type DeleteNoteReport struct {
	ActorID    string
	BookmarkID string
}

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
		UpdatedAt:             sql.NullString{},
		DescriptionMycomarkup: sql.NullString{},
		Tags:                  nil,
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
			if href := getString(amnt, "href"); stricks.ValidURL(href) {
				bookmark.URL = href
				break
			}
		}
	}

	// Lie detector
	if !stricks.SameHost(bookmark.ActorID, bookmark.ID) {
		return nil, ErrHostMismatch
	}

	// Verify required fields.
	mustBeNonEmpty := []string{bookmark.ID, bookmark.ActorID, bookmark.Title, bookmark.PublishedAt, bookmark.URL}
	for _, field := range mustBeNonEmpty {
		if field == "" {
			return nil, ErrEmptyField
		}
	}

	// Grabbing Mycomarkup
	source, ok := object["source"].(Dict)
	if ok && getString(source, "mediaType") == "text/mycomarkup" {
		mycomarkup := getString(source, "content")
		bookmark.DescriptionMycomarkup = sql.NullString{
			String: mycomarkup,
			Valid:  true,
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

func guessNoteObject(activity Dict) (note *types.RemoteBookmark, err error) {
	object, ok := activity["object"].(Dict)
	if !ok {
		return nil, ErrNoObject
	}

	bookmark, err := RemoteBookmarkFromDict(object)
	if err != nil {
		return nil, err
	}
	bookmark.Activity = activity["original activity"].([]byte)

	return bookmark, nil
}

func guessCreateNote(activity Dict) (report any, err error) {
	bookmark, err := guessNoteObject(activity)
	if err != nil {
		return nil, err
	}
	return CreateNoteReport{
		Bookmark: *bookmark,
	}, nil
}

func guessUpdateNote(activity Dict) (report any, err error) {
	bookmark, err := guessNoteObject(activity)
	if err != nil {
		return nil, err
	}
	return UpdateNoteReport{
		Bookmark: *bookmark,
	}, nil
}

func guessDeleteNote(activity Dict) (report any, err error) {
	deletion := DeleteNoteReport{
		ActorID:    getIDSomehow(activity, "actor"),
		BookmarkID: getIDSomehow(activity, "object"),
	}
	if !stricks.SameHost(deletion.ActorID, deletion.BookmarkID) {
		return nil, ErrHostMismatch
	}
	return deletion, nil
}
