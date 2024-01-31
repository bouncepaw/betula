package activities

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
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
	activity := dict{
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

func makeNote(post types.Bookmark) (dict, error) {
	if post.ID == 0 {
		return nil, errors.New("an empty ID was passed")
	}
	// Generating the timestamp
	t, err := time.Parse(types.TimeLayout, post.CreationTime)
	if err != nil {
		return nil, err
	}
	published := t.Format(time.RFC3339)

	// Generating the contents
	var content strings.Builder
	var tags []any

	content.WriteString(fmt.Sprintf(`<h3><a href="%s"'>%s</a></h3>`, html.EscapeString(post.URL), html.EscapeString(post.Title)))
	content.WriteString(string(myco.MarkupToHTML(post.Description)))
	if len(post.Tags) > 0 {
		content.WriteString("<p>")
		for i, tag := range post.Tags {
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
			tags = append(tags, dict{
				"type": "Hashtag",
				"name": "#" + tag.Name, // The # is needed
				"href": fmt.Sprintf("%s/tag/%s", settings.SiteURL(), tag.Name),
			})
		}
		content.WriteString("</p>")
	}

	activity := dict{
		"@context": []any{
			atContext,
			dict{
				"Hashtag": "https://www.w3.org/ns/activitystreams#Hashtag",
			},
		},
		"actor": betulaActor,
		"object": dict{
			"type":         "Note",
			"id":           fmt.Sprintf("%s/%d", settings.SiteURL(), post.ID),
			"actor":        betulaActor,
			"attributedTo": betulaActor,
			"to": []string{
				publicAudience,
				fmt.Sprintf("%s/followers", settings.SiteURL()),
			},
			"content": strings.ReplaceAll(
				strings.ReplaceAll(content.String(), "\t", ""),
				">\n", ">"),
			"name": post.Title,
			"source": map[string]string{
				// Misskey-style. They put text/x.misskeymarkdown though.
				"content":   post.Description,
				"mediaType": "text/mycomarkup",
			},
			"published": published,
		},
	}
	if len(tags) > 0 {
		activity["object"].(dict)["tag"] = tags
	}
	return activity, nil
}

func CreateNote(post types.Bookmark) ([]byte, error) {
	activity, err := makeNote(post)
	if err != nil {
		return nil, err
	}
	activity["type"] = "Create"
	activity["id"] = fmt.Sprintf("%s/%d?create", settings.SiteURL(), post.ID)
	return json.Marshal(activity)
}

func UpdateNote(post types.Bookmark) ([]byte, error) {
	activity, err := makeNote(post)
	if err != nil {
		return nil, err
	}
	activity["type"] = "Update"
	activity["id"] = fmt.Sprintf("%s/%d?update", settings.SiteURL(), post.ID)
	activity["object"].(dict)["updated"] = time.Now().Format(time.RFC3339)
	return json.Marshal(activity)
}

type CreateNoteReport struct {
	Bookmark types.RemoteBookmark
}

func guessCreate(activity dict) (report any, err error) {
	object, ok := activity["object"].(dict)
	if !ok {
		return nil, ErrNoObject
	}

	bookmark := types.RemoteBookmark{
		// Invariants
		RepostOf:  sql.NullString{},
		UpdatedAt: sql.NullString{},
		Activity:  activity["original activity"].([]byte),

		// Required fields
		ID:              getIDSomehow(activity, "object"),
		ActorID:         getIDSomehow(activity, "actor"),
		Title:           getString(object, "name"),
		DescriptionHTML: template.HTML(getString(object, "content")),
		PublishedAt:     getString(object, "published"),

		// Optional fields
		DescriptionMycomarkup: sql.NullString{},
		Tags:                  nil,
	}

	// Verify required fields.
	mustBeNonEmpty := []string{bookmark.ID, bookmark.ActorID, bookmark.Title, bookmark.PublishedAt}
	for _, field := range mustBeNonEmpty {
		if field == "" {
			return nil, ErrEmptyField
		}
	}

	// Grabbing Mycomarkup
	source, ok := object["source"].(dict)
	if ok && getString(source, "mediaType") == "text/mycomarkup" {
		mycomarkup := getString(source, "content")
		bookmark.DescriptionMycomarkup = sql.NullString{
			String: mycomarkup,
			Valid:  mycomarkup == "",
		}
	}

	// Collecting tags
	tags, ok := object["tag"].([]dict)
	for _, tag := range tags {
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

	return CreateNoteReport{bookmark}, nil
}
