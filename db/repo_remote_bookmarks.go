// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package db

import (
	"context"

	apports "git.sr.ht/~bouncepaw/betula/ports/activitypub"
	remotebookmarksports "git.sr.ht/~bouncepaw/betula/ports/remotebookmarks"
)

type RepoRemoteBookmarks struct {
}

var (
	_ apports.RemoteBookmarkRepository              = &RepoRemoteBookmarks{}
	_ remotebookmarksports.RemoteBookmarkRepository = &RepoRemoteBookmarks{}
)

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

func (repo *RepoRemoteBookmarks) Delete(ctx context.Context, bookmarkID string) error {
	_, err := db.ExecContext(ctx, `delete from RemoteBookmarks where ID = ?`, bookmarkID)
	return err
}

// TODO: port old queries to this repo
// https://codeberg.org/bouncepaw/betula/issues/138
