// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package db

import (
	apports "git.sr.ht/~bouncepaw/betula/ports/activitypub"
)

type RepoRemoteBookmarks struct {
}

var _ apports.RemoteBookmarkRepository = &RepoRemoteBookmarks{}

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

func (repo *RepoRemoteBookmarks) GetActorIDFor(bookmarkID string) (string, error) {
	row := db.QueryRow(
		`select ActorID from RemoteBookmarks where ID = ?`,
		bookmarkID)
	var actorID string
	err := row.Scan(&actorID)
	return actorID, err
}

// TODO: port old queries to this repo
// https://codeberg.org/bouncepaw/betula/issues/138
