// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package db

import likingports "git.sr.ht/~bouncepaw/betula/ports/liking"

type RepoRemoteBookmarks struct {
}

var _ likingports.RemoteBookmarkRepository = &RepoRemoteBookmarks{}

func NewRemoteBookmarkRepo() *RepoRemoteBookmarks {
	return &RepoRemoteBookmarks{}
}

func (repo *RepoRemoteBookmarks) Exists(bookmarkID string) (bool, error) {
	row := db.QueryRow(
		`select exists(select 1  from RemoteBookmarks where ID = ?)`,
		bookmarkID,
	)
	var exists bool
	err := row.Scan(&exists)
	return exists, err
}

// TODO: port old queries to this repo
// https://codeberg.org/bouncepaw/betula/issues/138
