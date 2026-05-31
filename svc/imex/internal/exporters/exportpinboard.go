// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package exporters

import (
	"crypto/md5"
	"fmt"
	"io"
	"iter"
	"strings"
	"time"

	"git.sr.ht/~bouncepaw/betula/pkg/bxtime"
	"git.sr.ht/~bouncepaw/betula/pkg/imex/pinboard"
	"git.sr.ht/~bouncepaw/betula/types"
)

type PinboardExporter struct{}

func NewPinboardExporter() *PinboardExporter {
	return &PinboardExporter{}
}

func (p PinboardExporter) Export(bookmarks iter.Seq[types.Bookmark], w io.Writer, now time.Time) error {
	var pbs []pinboard.Bookmark
	for bm := range bookmarks {
		pbs = append(pbs, domainBookmarkToPinboard(bm, now))
	}
	return pinboard.Write(pbs, w)
}

func domainBookmarkToPinboard(bm types.Bookmark, now time.Time) pinboard.Bookmark {
	var tagNames []string
	for _, tag := range bm.Tags {
		tagNames = append(tagNames, tag.Name)
	}
	tagsStr := strings.Join(tagNames, " ")

	t, err := time.Parse(types.TimeLayout, bm.CreationTime)
	if err != nil {
		t = now
	}

	shared := "no"
	if bm.Visibility == types.Public {
		shared = "yes"
	}

	return pinboard.Bookmark{
		Href:        bm.URL,
		Description: bm.Title,
		Extended:    bm.Description,
		Meta:        fmt.Sprintf("%x", md5.Sum([]byte(bm.Title+bm.Description+tagsStr+shared+"no"))),
		Hash:        fmt.Sprintf("%x", md5.Sum([]byte(bm.URL))),
		Time:        bxtime.TimeRFC3339(t.UTC()),
		Shared:      shared,
		ToRead:      "no",
		Tags:        tagNames,
	}
}
