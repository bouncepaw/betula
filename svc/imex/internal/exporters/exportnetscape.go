// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package exporters

import (
	"fmt"
	"io"
	"iter"
	"time"

	"git.sr.ht/~bouncepaw/betula/pkg/imex/netscape"
	"git.sr.ht/~bouncepaw/betula/types"
)

type NetscapeExporter struct {
	siteNameFn func() string
}

func NewNetscapeExporter(siteNameFn func() string) *NetscapeExporter {
	return &NetscapeExporter{
		siteNameFn: siteNameFn,
	}
}

func (n NetscapeExporter) Export(bookmarks iter.Seq[types.Bookmark], w io.Writer, now time.Time) error {
	var (
		publicFolder = &netscape.Folder{
			Title:    "Public bookmarks",
			Added:    time.Now().UTC(),
			Modified: time.Now().UTC(),
		}
		privateFolder = &netscape.Folder{
			Title:    "Private bookmarks",
			Added:    time.Now().UTC(),
			Modified: time.Now().UTC(),
		}
	)

	for bookmark := range bookmarks {
		nb := domainBookmarkToNetscape(bookmark, now)
		switch bookmark.Visibility {
		case types.Public:
			publicFolder.Items = append(publicFolder.Items, nb)
		case types.Private:
			privateFolder.Items = append(privateFolder.Items, nb)
		}
	}

	rootFolder := &netscape.Folder{
		Title:    fmt.Sprintf("%s — %s", n.siteNameFn(), now.Format(types.TimeLayout)),
		Added:    now.UTC(),
		Modified: now.UTC(),
		Items:    []netscape.Item{publicFolder},
	}
	if len(privateFolder.Items) > 0 {
		rootFolder.Items = append(rootFolder.Items, privateFolder)
	}

	return rootFolder.Write(w)
}

func domainBookmarkToNetscape(bookmark types.Bookmark, now time.Time) netscape.Bookmark {
	var tags []string
	for _, tag := range bookmark.Tags {
		tags = append(tags, tag.Name)
	}
	createdAt, err := time.Parse(types.TimeLayout, bookmark.CreationTime)
	if err != nil {
		createdAt = now // NOTE(bouncepaw): I mean, whatever.
	}

	return netscape.Bookmark{
		URL:         bookmark.URL,
		Title:       bookmark.Title,
		Description: bookmark.Description,
		Tags:        tags,
		Added:       createdAt,
		Modified:    now,
	}
}
