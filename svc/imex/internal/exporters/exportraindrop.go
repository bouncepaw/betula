// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package exporters

import (
	"io"
	"iter"
	"time"

	"git.sr.ht/~bouncepaw/betula/pkg/imex/raindrop"
	"git.sr.ht/~bouncepaw/betula/types"
)

type RaindropExporter struct{}

func NewRaindropExporter() *RaindropExporter {
	return &RaindropExporter{}
}

func (r RaindropExporter) Export(bookmarks iter.Seq[types.Bookmark], w io.Writer, now time.Time) error {
	var rds []raindrop.Bookmark
	for bm := range bookmarks {
		rds = append(rds, domainBookmarkToRaindrop(bm, now))
	}
	return raindrop.Write(rds, w)
}

func domainBookmarkToRaindrop(bm types.Bookmark, now time.Time) raindrop.Bookmark {
	var tags []string
	for _, tag := range bm.Tags {
		tags = append(tags, tag.Name)
	}

	created, err := time.Parse(types.TimeLayout, bm.CreationTime)
	if err != nil {
		created = now
	}

	return raindrop.Bookmark{
		Title:   bm.Title,
		Note:    bm.Description,
		URL:     bm.URL,
		Tags:    tags,
		Created: created.UTC(),
	}
}
