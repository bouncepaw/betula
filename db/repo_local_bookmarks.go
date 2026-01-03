// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package db

import likingports "git.sr.ht/~bouncepaw/betula/ports/liking"

type RepoLocalBookmarks struct{}

var _ likingports.LocalBookmarkRepository = &RepoLocalBookmarks{}

func NewLocalBookmarksRepo() *RepoLocalBookmarks {
	return &RepoLocalBookmarks{}
}

func (repo *RepoLocalBookmarks) Exists(bookmarkID int) (bool, error) {
	row := db.QueryRow(
		`select exists(select 1 from Bookmarks where ID = ?)`,
		bookmarkID,
	)
	var exists bool
	err := row.Scan(&exists)
	return exists, err
}

// TODO: port old queries to this repo
// https://codeberg.org/bouncepaw/betula/issues/138
