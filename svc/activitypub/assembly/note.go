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

	apports "git.sr.ht/~bouncepaw/betula/ports/activitypub"

	"golang.org/x/net/html"

	"git.sr.ht/~bouncepaw/betula/pkg/myco"
	"git.sr.ht/~bouncepaw/betula/types"
)

func (asm *Assembler) DeleteNote(postID int) (json.RawMessage, error) {
	id := fmt.Sprintf("%s/%d", asm.siteURLFn(), postID)
	activity := apports.Dict{
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

var theCoolContext = []any{
	atContext,
	apports.Dict{
		"Hashtag": "https://www.w3.org/ns/activitystreams#Hashtag",

		// https://codeberg.org/fediverse/fep/src/branch/main/fep/044f/fep-044f.md#advertising-a-quote-policy
		"gts": "https://gotosocial.org/ns#",
		"interactionPolicy": map[string]string{
			"@id":   "gts:interactionPolicy",
			"@type": "@id",
		},
		"canQuote": map[string]string{
			"@id":   "gts:canQuote",
			"@type": "@id",
		},
		"automaticApproval": map[string]string{
			"@id":   "gts:automaticApproval",
			"@type": "@id",
		},

		// https://codeberg.org/fediverse/fep/src/branch/main/fep/044f/fep-044f.md#compatibility-with-other-quote-implementations
		"quoteUrl":       "as:quoteUrl",
		"quoteUri":       "http://fedibird.com/ns#quoteUri",
		"_misskey_quote": "https://misskey-hub.net/ns/#_misskey_quote",
		"quote": map[string]string{
			"@id":   "https://w3id.org/fep/044f#quote",
			"@type": "@id",
		},
	},
}

func (asm *Assembler) writeContentAndTags(
	object apports.Dict,
	url *string,
	title *string,
	text string,
	quotedObjectID *string,
	bmTags []types.Tag,
) {
	// Generating the contents
	var content strings.Builder

	if url != nil && title != nil {
		content.WriteString(fmt.Sprintf(
			`<h3><a href="%s">%s</a></h3>`,
			html.EscapeString(*url),
			html.EscapeString(*title),
		))

		object["attachment"] = []apports.Dict{
			{ // Lemmy-style.
				"href": *url,
				"type": "Link",
			},
		}
		object["name"] = *title
	}
	content.WriteString(string(myco.MarkupToHTML(text)))

	var tags []any
	if len(bmTags) > 0 {
		content.WriteString("<p>")
		for i, tag := range bmTags {
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
			tags = append(tags, apports.Dict{
				"type": "Hashtag",
				"name": "#" + tag.Name, // The # is needed
				"href": fmt.Sprintf("%s/tag/%s", asm.siteURLFn(), tag.Name),
			})
		}
		content.WriteString("</p>")
	}
	if len(tags) > 0 {
		object["tag"] = tags
	}

	// https://codeberg.org/fediverse/fep/src/branch/main/fep/044f/fep-044f.md#backward-compatibility-considerations
	if quotedObjectID != nil {
		re := html.EscapeString(*quotedObjectID)
		notice := fmt.Sprintf(
			`

<span class="quote-inline"><br/>RE: <a href="%s">%s</a></span>`,
			re, re)
		content.WriteString(notice)
	}

	object["content"] = strings.ReplaceAll(
		strings.ReplaceAll(content.String(), "\t", ""),
		">\n", ">")
	object["source"] = map[string]string{
		// Misskey-style. They put text/x.misskeymarkdown though.
		"content":   text,
		"mediaType": "text/mycomarkup",
	}
}

func (asm *Assembler) NoteFromBookmark(bookmark types.Bookmark) (apports.Dict, error) {
	if bookmark.ID == 0 {
		return nil, errors.New("an empty ID was passed")
	}
	// Generating the timestamp
	t, err := time.Parse(types.TimeLayout, bookmark.CreationTime)
	if err != nil {
		return nil, err
	}
	published := t.Format(time.RFC3339)

	object := apports.Dict{
		"@context":     theCoolContext,
		"type":         "Note",
		"id":           fmt.Sprintf("%s/%d", asm.siteURLFn(), bookmark.ID),
		"actor":        asm.actor(),
		"attributedTo": asm.actor(),
		"to": []string{
			publicAudience,
			fmt.Sprintf("%s/followers", asm.siteURLFn()),
		},
		"published": published,

		// https://codeberg.org/fediverse/fep/src/branch/main/fep/044f/fep-044f.md#advertising-a-quote-policy
		"interactionPolicy": apports.Dict{
			"canQuote": map[string]string{
				"automaticApproval": "https://www.w3.org/ns/activitystreams#Public",
			},
		},
	}

	if bookmark.RepostOf == nil {
		asm.writeContentAndTags(object, &bookmark.URL, &bookmark.Title, bookmark.Description, nil, bookmark.Tags)
	} else {
		// https://codeberg.org/fediverse/fep/src/branch/main/fep/044f/fep-044f.md#compatibility-with-other-quote-implementations
		// Deliberately no quote Link object in the tag list.
		quoteFields := []string{"quote", "quoteUrl", "quoteUri", "_misskey_quote"}
		for _, field := range quoteFields {
			object[field] = *bookmark.RepostOf
		}

		text := ""
		if bookmark.RemarkText != nil {
			text = *bookmark.RemarkText
		}
		asm.writeContentAndTags(object, nil, nil, text, bookmark.RepostOf, bookmark.Tags)
	}

	return object, nil
}

func (asm *Assembler) makeNoteAction(bookmark types.Bookmark) (apports.Dict, error) {
	object, err := asm.NoteFromBookmark(bookmark)
	if err != nil {
		return nil, err
	}

	delete(object, "@context")

	activity := apports.Dict{
		"@context": theCoolContext,
		"actor":    asm.actor(),
		"object":   object,
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
	activity["object"].(apports.Dict)["updated"] = time.Now().UTC().Format(time.RFC3339)
	return json.Marshal(activity)
}

func (asm *Assembler) UpdateNoteWithLikes(post types.Bookmark, likeCounter int) (json.RawMessage, error) {
	activity, err := asm.makeNoteAction(post)
	if err != nil {
		return nil, err
	}
	activity["type"] = "Update"
	activity["id"] = fmt.Sprintf("%s/%d?update", asm.siteURLFn(), post.ID)
	activity["object"].(apports.Dict)["updated"] = time.Now().UTC().Format(time.RFC3339)

	likeCollectionID := fmt.Sprintf("%s/%d?likes", asm.siteURLFn(), post.ID)
	activity["object"].(apports.Dict)["likes"] = apports.Collection{
		ID:         &likeCollectionID,
		Type:       "Collection",
		TotalItems: likeCounter,
	}
	return json.Marshal(activity)
}
