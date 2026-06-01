// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package importers

import (
	"io"
	"iter"
	"log/slog"
	"maps"
	"net/url"
	"regexp"
	"strings"
	"time"

	wwwports "git.sr.ht/~bouncepaw/betula/ports/www"
	"git.sr.ht/~bouncepaw/betula/types"
)

var reURL = regexp.MustCompile(`(http|https|gemini|gopher)://\S+`)

type PlainImporter struct {
	nWorkers int
	www      wwwports.WorldWideWeb
}

func NewPlainImporter(nWorkers int, www wwwports.WorldWideWeb) *PlainImporter {
	return &PlainImporter{
		nWorkers: nWorkers,
		www:      www,
	}
}

func (p *PlainImporter) Probe(_ io.ReadSeeker) (bool, error) {
	return true, nil
}

func (p *PlainImporter) Import(r io.Reader) (iter.Seq2[types.Bookmark, error], error) {
	urls, err := extractURLs(r)
	if err != nil {
		return nil, err
	}

	var (
		now       = time.Now().UTC().Format(types.TimeLayout)
		bookmarks = p.urlsToBookmarks(now, urls)
	)
	return func(yield func(types.Bookmark, error) bool) {
		var count int
		for bm := range bookmarks {
			count++
			if !yield(bm, nil) {
				slog.Info("Plain import stopped early", "bookmarkCount", count)
				return
			}
		}
	}, nil
}

func (p *PlainImporter) urlsToBookmarks(now string, urls iter.Seq[string]) <-chan types.Bookmark {
	var (
		bookmarks = make(chan types.Bookmark, p.nWorkers)
		sema      = make(chan struct{}, p.nWorkers)
	)

	go func() {
		for u := range urls {
			// Rawdogging.
			sema <- struct{}{}
			bookmarks <- p.urlToBookmark(u, now)
			<-sema
		}

		close(sema)
		close(bookmarks)
	}()
	return bookmarks
}

func (p *PlainImporter) urlToBookmark(u, now string) types.Bookmark {
	var title string
	if strings.HasPrefix(u, "http") {
		var err error
		title, err = p.www.TitleOfPage(u)
		if err != nil {
			slog.Warn("Failed to fetch page title", "url", u, "err", err)
		}
	} else {
		slog.Info("Not fetching page title for non-web page", "url", u)
	}
	if title == "" {
		title = types.CleanerLink(u)
	}

	return types.Bookmark{
		CreationTime: now,
		URL:          u,
		Title:        title,
		Visibility:   types.Private,
	}
}

func extractURLs(r io.Reader) (iter.Seq[string], error) {
	// NOTE(bouncepaw): We'll have beautiful iterator regexp functions.
	// Consider refactoring once the bright future is here.
	//
	// => https://github.com/golang/go/issues/61902

	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	matches := reURL.FindAllString(string(data), -1)
	urls := make(map[string]struct{})
	var skipped int
	for _, match := range matches {
		// NOTE(bouncepaw): Not a fan of this.
		candidate := strings.TrimRight(match, ".,;:!?(<{[]}>)")

		u, err := url.Parse(candidate)
		if err != nil || u.Scheme == "" || (u.Host == "" && u.Opaque == "") {
			skipped++
			slog.Debug("Skipping invalid URL candidate", "candidate", candidate, "err", err)
			continue
		}
		urls[candidate] = struct{}{}
	}

	slog.Info("Extracted URLs from plain text",
		"bytes", len(data),
		"matches", len(matches),
		"unique", len(urls),
		"skipped", skipped,
	)
	return maps.Keys(urls), nil
}
