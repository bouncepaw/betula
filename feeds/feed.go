// Package feeds manages RSS feed generation.
package feeds

import (
	"fmt"
	"git.sr.ht/~bouncepaw/betula/db"
	"git.sr.ht/~bouncepaw/betula/myco"
	"git.sr.ht/~bouncepaw/betula/settings"
	"git.sr.ht/~bouncepaw/betula/types"
	"github.com/gorilla/feeds"
	"log"
	"strings"
	"time"
)

func fiveLastDays(now time.Time) (days []time.Time, dayStamps []string, dayPosts [][]types.Post) {
	days = make([]time.Time, 5)
	dayStamps = make([]string, 5)
	dayPosts = make([][]types.Post, 5)
	for i := 0; i < 5; i++ {
		day := now.AddDate(0, 0, -i-1)
		day = time.Date(day.Year(), day.Month(), day.Day(), 23, 59, 59, 0, time.UTC)
		days[i] = day

		dayStamps[i] = day.Format("2006-01-02")

		dayPosts[i] = db.PostsForDay(false, dayStamps[i])
	}
	return days, dayStamps, dayPosts
}

func Posts() *feeds.Feed {
	author := &feeds.Author{
		Name: db.MetaEntry[string](db.BetulaMetaAdminUsername),
	}
	now := time.Now()
	_, _, dayPosts := fiveLastDays(now)

	feed := &feeds.Feed{
		Title:       fmt.Sprintf("%s posts", settings.SiteName()),
		Link:        &feeds.Link{Href: settings.SiteURL()},
		Description: fmt.Sprintf("All public posts are sent to this feed."),
		Author:      author,
		Created:     now,
		Items:       []*feeds.Item{},
	}

	for _, posts := range dayPosts {
		for _, post := range posts {
			creationTime, err := time.Parse(types.TimeLayout, post.CreationTime)
			if err != nil {
				log.Printf("The timestamp for post no. %d ‘%s’ is invalid: %s\n",
					post.ID, post.Title, post.CreationTime)
				continue
			}

			var entry = &feeds.Item{
				Title: post.Title,
				Link: &feeds.Link{
					Href: fmt.Sprintf("%s/%d", settings.SiteURL(), post.ID),
				},
				Author:      author,
				Description: descriptionForOnePost(post),
				Created:     creationTime,
			}
			feed.Items = append(feed.Items, entry)
		}
	}

	return feed
}

func Digest() *feeds.Feed {
	author := &feeds.Author{
		Name: db.MetaEntry[string](db.BetulaMetaAdminUsername),
	}
	now := time.Now()
	days, dayStamps, dayPosts := fiveLastDays(now)

	feed := &feeds.Feed{
		Title:       fmt.Sprintf("%s daily digest", settings.SiteName()),
		Link:        &feeds.Link{Href: settings.SiteURL()},
		Description: fmt.Sprintf("Every day, a list of all links posted that day is sent."),
		Author:      author,
		Created:     now,
		Items:       []*feeds.Item{},
	}

	for i, posts := range dayPosts {
		if posts == nil {
			continue
		}
		var entry = &feeds.Item{
			Title: fmt.Sprintf("%s %s", settings.SiteName(), dayStamps[i]),
			Link: &feeds.Link{
				Href: fmt.Sprintf("%s/day/%s", settings.SiteURL(), dayStamps[i]),
			},
			Author:      author,
			Description: descriptionFromPosts(posts, dayStamps[i]),
			Created:     days[i],
		}
		feed.Items = append(feed.Items, entry)
		log.Println("iter")
	}

	return feed
}

const descriptionTemplate = `
<p>🔗 <a href="%s">%s</a></p>
%s
%s
`

func descriptionForOnePost(post types.Post) string {
	var catBuf strings.Builder
	for i, cat := range post.Categories {
		if i > 0 {
			catBuf.WriteString(", ")
		}
		catBuf.WriteString(fmt.Sprintf(`<a href="/cat/%s">%s</a>`, cat.Name, cat.Name))
	}

	return fmt.Sprintf(
		descriptionTemplate,
		post.URL,
		types.StripCommonProtocol(post.URL),
		func() string {
			if len(post.Categories) > 0 {
				return "<p>🏷 " + catBuf.String() + "</p>"
			}
			return ""
		}(),
		myco.MarkupToHTML(post.Description),
	)
}

func descriptionFromPosts(posts []types.Post, dayStamp string) string {
	var buf strings.Builder

	for _, post := range posts {
		buf.WriteString(fmt.Sprintf(`<h2><a href="/%d">%s</a></h2>`, post.ID, post.Title))
		buf.WriteString(descriptionForOnePost(post))
	}

	return buf.String()
}
