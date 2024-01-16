package activities

import (
	"encoding/json"
	"fmt"
	"golang.org/x/net/html"
	"strings"
	"time"

	"git.sr.ht/~bouncepaw/betula/myco"
	"git.sr.ht/~bouncepaw/betula/settings"
	"git.sr.ht/~bouncepaw/betula/types"
)

func CreateNote(post types.Post) ([]byte, error) {
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
			if i > 0 {
				content.WriteString(", ")
			}
			content.WriteString("#" + tag.Name)

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
			atContextMycomarkupExtension,
		},
		"type":  "Create",
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
			"content": content.String(),
			"name":    post.Title,
			"mycomarkup": dict{
				"sourceText": post.Description,
			},
			"published": published,
		},
	}
	if len(tags) > 0 {
		activity["tag"] = tags
	}
	return json.Marshal(activity)
}
