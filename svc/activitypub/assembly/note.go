// SPDX-FileCopyrightText: 2024 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package assembly

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/net/html"

	"git.sr.ht/~bouncepaw/betula/pkg/myco"
	"git.sr.ht/~bouncepaw/betula/types"
)

func (asm *Assembler) DeleteNote(postID int) (json.RawMessage, error) {
	id := fmt.Sprintf("%s/%d", asm.siteURLFn(), postID)
	activity := Dict{
		"@context": atContext,
		"type":     "Delete",
		"actor":    asm.actor(),
		"to": []string{
			publicAudience,
			fmt.Sprintf("%s/followers", asm.siteURLFn()),
		},
		"id":     id + "?delete",
		"object": id,
	}
	return json.Marshal(activity)
}

func (asm *Assembler) NoteFromBookmark(bookmark types.Bookmark) (Dict, error) {
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

	content.WriteString(fmt.Sprintf(`<h3><a href="%s">%s</a></h3>`, html.EscapeString(bookmark.URL), html.EscapeString(bookmark.Title)))
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
					asm.siteURLFn(),
					tag.Name,
					tag.Name,
				))

			// https://docs.joinmastodon.org/spec/activitypub/#Hashtag
			tags = append(tags, Dict{
				"type": "Hashtag",
				"name": "#" + tag.Name, // The # is needed
				"href": fmt.Sprintf("%s/tag/%s", asm.siteURLFn(), tag.Name),
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
		"id":           fmt.Sprintf("%s/%d", asm.siteURLFn(), bookmark.ID),
		"actor":        asm.actor(),
		"attributedTo": asm.actor(),
		"to": []string{
			publicAudience,
			fmt.Sprintf("%s/followers", asm.siteURLFn()),
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

func (asm *Assembler) makeNoteAction(bookmark types.Bookmark) (Dict, error) {
	object, err := asm.NoteFromBookmark(bookmark)
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
		"actor":  asm.actor(),
		"object": object,
	}
	return activity, nil
}

func (asm *Assembler) CreateNote(post types.Bookmark) (json.RawMessage, error) {
	activity, err := asm.makeNoteAction(post)
	if err != nil {
		return nil, err
	}
	activity["type"] = "Create"
	activity["id"] = fmt.Sprintf("%s/%d?create", asm.siteURLFn(), post.ID)
	return json.Marshal(activity)
}

func (asm *Assembler) UpdateNote(post types.Bookmark) (json.RawMessage, error) {
	activity, err := asm.makeNoteAction(post)
	if err != nil {
		return nil, err
	}
	activity["type"] = "Update"
	activity["id"] = fmt.Sprintf("%s/%d?update", asm.siteURLFn(), post.ID)
	activity["object"].(Dict)["updated"] = time.Now().UTC().Format(time.RFC3339)
	return json.Marshal(activity)
}

func (asm *Assembler) UpdateNoteWithLikes(post types.Bookmark, likeCounter int) (json.RawMessage, error) {
	activity, err := asm.makeNoteAction(post)
	if err != nil {
		return nil, err
	}
	activity["type"] = "Update"
	activity["id"] = fmt.Sprintf("%s/%d?update", asm.siteURLFn(), post.ID)
	activity["object"].(Dict)["updated"] = time.Now().UTC().Format(time.RFC3339)

	likeCollectionID := fmt.Sprintf("%s/%d?likes", asm.siteURLFn(), post.ID)
	activity["object"].(Dict)["likes"] = Collection{
		ID:         &likeCollectionID,
		Type:       "Collection",
		TotalItems: likeCounter,
	}
	return json.Marshal(activity)
}
