// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package importers

import (
	"io"
	"iter"

	"git.sr.ht/~bouncepaw/betula/pkg/imex/netscape"
	"git.sr.ht/~bouncepaw/betula/types"
)

type NetscapeImporter struct{}

func NewNetscapeImporter() *NetscapeImporter {
	return &NetscapeImporter{}
}

func (n *NetscapeImporter) Probe(seeker io.ReadSeeker) (bool, error) {
	return netscape.Probe(seeker)
}

func (n *NetscapeImporter) Import(r io.Reader) (iter.Seq2[types.Bookmark, error], error) {
	rootFolder, err := netscape.Read(r)
	if err != nil {
		return nil, err
	}

	return importNetscapeFolder(rootFolder, nil), nil
}

func importNetscapeFolder(folder *netscape.Folder, acc []string) iter.Seq2[types.Bookmark, error] {
	return func(yield func(types.Bookmark, error) bool) {
		title := types.CanonicalTagName(folder.Title)
		for _, item := range folder.Items {
			switch item := item.(type) {
			case netscape.Bookmark:
				if !yield(netscapeBookmarkToDomain(item, title), nil) {
					return
				}
			case *netscape.Folder:
				for bm, err := range importNetscapeFolder(item, append(acc, title)) {
					if !yield(bm, err) {
						return
					}
				}
			}
		}
	}
}

func netscapeBookmarkToDomain(bookmark netscape.Bookmark, folderName string) types.Bookmark {
	var tags []types.Tag
	for _, tag := range bookmark.Tags {
		tags = append(tags, types.Tag{Name: types.CanonicalTagName(tag)})
	}
	return types.Bookmark{
		CreationTime: bookmark.Added.Format(types.TimeLayout),
		Tags:         append(tags, types.Tag{Name: folderName}),
		URL:          bookmark.URL,
		Title:        bookmark.Title,
		Description:  bookmark.Description,
		Visibility:   types.Private,
	}
}
