// SPDX-FileCopyrightText: 2023 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2024 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2024 Danila Gorelko
// SPDX-FileCopyrightText: 2025 Danila Gorelko
// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package feedssvc

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"git.sr.ht/~bouncepaw/betula/db"
	"git.sr.ht/~bouncepaw/betula/pkg/myco"
	"git.sr.ht/~bouncepaw/betula/pkg/rss"
	feedsports "git.sr.ht/~bouncepaw/betula/ports/feeds"
	"git.sr.ht/~bouncepaw/betula/settings"
	"git.sr.ht/~bouncepaw/betula/types"
)

type Service struct{}

var _ feedsports.Service = &Service{}

func New() *Service {
	// TODO: feed repo https://codeberg.org/bouncepaw/betula/issues/138
	return &Service{}
}

func (svc *Service) DigestFeed() (*rss.Feed, error) {
	slog.Info("Generating a digest feed")
	author := settings.AdminUsername()

	now := time.Now()
	days, dayStamps, dayBookmarks := fiveLastDays(now)

	feed := rss.Feed{
		Title:       fmt.Sprintf("%s daily digest", settings.SiteName()),
		Link:        settings.SiteURL(),
		Description: fmt.Sprintf("Every day, a list of all bookmarks published that day is sent."),
		PubDate:     now.Format(rssTimeFormat),
		Items:       []*rss.Item{},
	}

	for i, bookmarks := range dayBookmarks {
		if bookmarks == nil {
			continue
		}
		var entry = &rss.Item{
			Title:  fmt.Sprintf("%s %s", settings.SiteName(), dayStamps[i]),
			Link:   fmt.Sprintf("%s/day/%s", settings.SiteURL(), dayStamps[i]),
			Author: author,
			Description: rss.CData{
				Data: descriptionFromBookmarks(bookmarks),
			},
			PubDate: days[i].Format(rssTimeFormat),
		}
		feed.Items = append(feed.Items, entry)
	}

	return &feed, nil
}

func (svc *Service) BookmarksFeed() (*rss.Feed, error) {
	slog.Info("Generating a bookmarks feed")
	author := settings.AdminUsername()

	now := time.Now().AddDate(0, 0, 1)
	_, _, dayBookmarks := fiveLastDays(now)

	feed := rss.Feed{
		Title:       fmt.Sprintf("%s bookmarks", settings.SiteName()),
		Link:        settings.SiteURL(),
		Description: fmt.Sprintf("All public bookmarks are sent to this feed."),
		PubDate:     now.Format(rssTimeFormat),
		Items:       []*rss.Item{},
	}

	for _, bookmarks := range dayBookmarks {
		for _, bm := range bookmarks {
			creationTime, err := time.Parse(types.TimeLayout, bm.CreationTime)
			if err != nil {
				slog.Error("Invalid creation time in bookmarks feed",
					"bookmarkID", bm.ID, "title", bm.Title, "creationTime", bm.CreationTime)
				continue
			}

			var entry = &rss.Item{
				Title:  bm.Title,
				Link:   bm.URL,
				Author: author,
				Description: rss.CData{
					bookmarkDescription(bm),
				},
				PubDate: creationTime.Format(rssTimeFormat),
			}
			feed.Items = append(feed.Items, entry)
		}
	}

	return &feed, nil
}

const rssTimeFormat = time.RFC822

func fiveLastDays(now time.Time) (days []time.Time, dayStamps []string, dayBookmarks [][]types.Bookmark) {
	days = make([]time.Time, 5)
	dayStamps = make([]string, 5)
	dayBookmarks = make([][]types.Bookmark, 5)
	for i := 0; i < 5; i++ {
		day := now.AddDate(0, 0, -i-1)
		day = time.Date(day.Year(), day.Month(), day.Day(), 23, 59, 59, 0, time.UTC)
		days[i] = day
		dayStamps[i] = day.Format(time.DateOnly)
		dayBookmarks[i] = db.BookmarksForDay(false, dayStamps[i])
	}
	return days, dayStamps, dayBookmarks
}

const descriptionTemplate = `
<h2><a href="%s">%s</a></h2>
<p>🔗 <a href="%s">%s</a></p>
%s
%s
`

func bookmarkDescription(bm types.Bookmark) string {
	var tagBuf strings.Builder
	for i, tag := range bm.Tags {
		if i > 0 {
			tagBuf.WriteString(", ")
		}
		tagBuf.WriteString(fmt.Sprintf(`<a href="/tag/%s">%s</a>`, tag.Name, tag.Name))
	}

	return fmt.Sprintf(
		descriptionTemplate,
		bm.URL,
		bm.Title,
		bm.URL,
		types.CleanerLink(bm.URL),
		func() string {
			if len(bm.Tags) > 0 {
				return "<p>🏷 " + tagBuf.String() + "</p>"
			}
			return ""
		}(),
		myco.MarkupToHTML(bm.Description),
	)
}

func descriptionFromBookmarks(bookmarks []types.Bookmark) string {
	var buf strings.Builder

	for _, bm := range bookmarks {
		buf.WriteString(bookmarkDescription(bm))
	}

	return buf.String()
}
