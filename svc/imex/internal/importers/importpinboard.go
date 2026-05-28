// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package importers

import (
	"io"
	"iter"
	"time"

	"git.sr.ht/~bouncepaw/betula/pkg/imex/pinboard"
	"git.sr.ht/~bouncepaw/betula/types"
)

type PinboardImporter struct{}

func NewPinboardImporter() *PinboardImporter {
	return &PinboardImporter{}
}

func (p *PinboardImporter) Probe(seeker io.ReadSeeker) (bool, error) {
	return pinboard.Probe(seeker)
}

func (p *PinboardImporter) Import(r io.Reader) (iter.Seq2[types.Bookmark, error], error) {
	bookmarks, err := pinboard.Read(r)
	if err != nil {
		return nil, err
	}
	return func(yield func(types.Bookmark, error) bool) {
		for _, b := range bookmarks {
			if !yield(pinboardBookmarkToDomain(b), nil) {
				return
			}
		}
	}, nil
}

func pinboardBookmarkToDomain(b pinboard.Bookmark) types.Bookmark {
	var tags []types.Tag
	for _, tag := range b.Tags {
		tags = append(tags, types.Tag{Name: types.CanonicalTagName(tag)})
	}

	visibility := types.Private
	if b.Shared == "yes" {
		visibility = types.Public
	}

	return types.Bookmark{
		CreationTime: time.Time(b.Time).Format(types.TimeLayout),
		URL:          b.Href,
		Title:        b.Description,
		Description:  b.Extended,
		Tags:         tags,
		Visibility:   visibility,
	}
}
