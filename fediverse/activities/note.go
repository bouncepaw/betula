package activities

import (
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/net/html"
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
