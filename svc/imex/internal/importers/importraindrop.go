// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package importers

import (
	"io"
	"iter"

	"git.sr.ht/~bouncepaw/betula/pkg/imex/raindrop"
	"git.sr.ht/~bouncepaw/betula/types"
)

type RaindropImporter struct{}

func NewRaindropImporter() *RaindropImporter {
	return &RaindropImporter{}
}

func (r *RaindropImporter) Probe(seeker io.ReadSeeker) (bool, error) {
	return raindrop.Probe(seeker)
}

func (r *RaindropImporter) Import(rd io.Reader) (iter.Seq2[types.Bookmark, error], error) {
	bookmarks, err := raindrop.Read(rd)
	if err != nil {
		return nil, err
	}
	return func(yield func(types.Bookmark, error) bool) {
		for _, b := range bookmarks {
			if !yield(raindropBookmarkToDomain(b), nil) {
				return
			}
		}
	}, nil
}

func raindropBookmarkToDomain(b raindrop.Bookmark) types.Bookmark {
	var tags []types.Tag
	for _, tag := range b.Tags {
		tags = append(tags, types.Tag{Name: types.CanonicalTagName(tag)})
	}

	// Turning Raindrop folders (collections) into Betula tags.
	// Should Betula have folders too?
	// People had asked for that, but I still don't see any value 😔
	if b.Folder != "" {
		tags = append(tags, types.Tag{Name: types.CanonicalTagName(b.Folder)})
	}

	description := b.Note
	if description == "" {
		description = b.Excerpt
	}

	return types.Bookmark{
		CreationTime: b.Created.Format(types.TimeLayout),
		URL:          b.URL,
		Title:        b.Title,
		Description:  description,
		Tags:         tags,
		Visibility:   types.Private,
	}
}
