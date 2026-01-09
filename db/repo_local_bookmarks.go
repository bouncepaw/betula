// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package db

import (
	"context"
	likingports "git.sr.ht/~bouncepaw/betula/ports/liking"
	"git.sr.ht/~bouncepaw/betula/types"
)

type RepoLocalBookmarks struct{}

var _ likingports.LocalBookmarkRepository = &RepoLocalBookmarks{}

func NewLocalBookmarksRepo() *RepoLocalBookmarks {
	return &RepoLocalBookmarks{}
}

func (repo *RepoLocalBookmarks) Exists(
	ctx context.Context,
	bookmarkID int,
) (bool, error) {
	row := db.QueryRowContext(
		ctx,
		`select exists(select 1 from Bookmarks where ID = ?)`,
		bookmarkID,
	)
	var exists bool
	err := row.Scan(&exists)
	return exists, err
}

func (repo *RepoLocalBookmarks) GetBookmarkByID(
	ctx context.Context,
	id int,
) (types.Bookmark, error) {
	row := db.QueryRowContext(ctx, `
		select ID, URL, Title, Description, Visibility, CreationTime, RepostOf, OriginalAuthorID 
		from Bookmarks
		where ID = ? and DeletionTime is null
	`, id)

	var b types.Bookmark
	err := row.Scan(&b.ID, &b.URL, &b.Title, &b.Description, &b.Visibility, &b.CreationTime, &b.RepostOf, &b.OriginalAuthor)
	return b, err
}

// TODO: port old queries to this repo
// https://codeberg.org/bouncepaw/betula/issues/138
