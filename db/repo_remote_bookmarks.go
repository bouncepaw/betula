// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package db

import (
	"context"

	apports "git.sr.ht/~bouncepaw/betula/ports/activitypub"
	remotebookmarksports "git.sr.ht/~bouncepaw/betula/ports/remotebookmarks"
	"git.sr.ht/~bouncepaw/betula/types"
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

func (repo *RepoRemoteBookmarks) GetRemoteBookmarksBy(authorID string, page uint) (bookmarks []types.RemoteBookmark, total uint) {
	total = querySingleValue[uint](`select count(ID) from RemoteBookmarks where ActorID = ?`, authorID)

	rows := mustQuery(`
select ID, RepostOf, ActorID, Title, DescriptionHTML, DescriptionMycomarkup, PublishedAt, UpdatedAt, URL
from RemoteBookmarks
where ActorID = ?
order by PublishedAt desc
limit ?
offset (? * (? - 1))
`, authorID, types.BookmarksPerPage, types.BookmarksPerPage, page) // same paging for local bookmarks

	for rows.Next() {
		var b types.RemoteBookmark
		mustScan(rows, &b.ID, &b.RepostOf, &b.ActorID, &b.Title, &b.DescriptionHTML, &b.DescriptionMycomarkup, &b.PublishedAt, &b.UpdatedAt, &b.URL)
		bookmarks = append(bookmarks, b)
	}

	// huh up to 64 additional queries??
	for i := range bookmarks {
		rows = mustQuery(`select Name from RemoteTags where BookmarkID = ?`, bookmarks[i].ID)
		for rows.Next() {
			var tag types.Tag
			mustScan(rows, &tag.Name)
			bookmarks[i].Tags = append(bookmarks[i].Tags, tag)
		}
	}

	return
}

func (repo *RepoRemoteBookmarks) GetRemoteBookmarks(page uint) (bookmarks []types.RemoteBookmark, total uint) {
	total = querySingleValue[uint](`
select count(RB.ID) 
from RemoteBookmarks RB
inner join Following F on RB.ActorID = F.ActorID
where F.AcceptedStatus = 1`)

	rows := mustQuery(`
select RB.ID, RB.RepostOf, RB.ActorID, RB.Title, RB.DescriptionHTML, RB.DescriptionMycomarkup, RB.PublishedAt, RB.UpdatedAt, RB.URL
from RemoteBookmarks RB
inner join Following F on RB.ActorID = F.ActorID
where F.AcceptedStatus = 1
order by RB.PublishedAt desc
limit ?
offset (? * (? - 1))
`, types.BookmarksPerPage, types.BookmarksPerPage, page) // same paging for local bookmarks

	for rows.Next() {
		var b types.RemoteBookmark
		mustScan(rows, &b.ID, &b.RepostOf, &b.ActorID, &b.Title, &b.DescriptionHTML, &b.DescriptionMycomarkup, &b.PublishedAt, &b.UpdatedAt, &b.URL)
		bookmarks = append(bookmarks, b)
	}

	// huh up to 64 additional queries??
	for i := range bookmarks {
		rows = mustQuery(`select Name from RemoteTags where BookmarkID = ?`, bookmarks[i].ID)
		for rows.Next() {
			var tag types.Tag
			mustScan(rows, &tag.Name)
			bookmarks[i].Tags = append(bookmarks[i].Tags, tag)
		}
	}

	return
}

func (repo *RepoRemoteBookmarks) InsertRemoteBookmark(b types.RemoteBookmark) {
	mustExec(`
insert into RemoteBookmarks
    (ID,  RepostOf,   ActorID,   Title,   URL, DescriptionHTML,   DescriptionMycomarkup, PublishedAt,  UpdatedAt, Activity)
values
	(?, ?, ?, ?, ?, ?, ?, ?, null, ?)
on conflict do nothing`,
		b.ID, b.RepostOf, b.ActorID, b.Title, b.URL, b.DescriptionHTML, b.DescriptionMycomarkup, b.PublishedAt, b.Activity)

	for _, tag := range b.Tags {
		mustExec(`insert or replace into RemoteTags (Name, BookmarkID) values (?, ?)`, tag.Name, b.ID)
	}
}

func (repo *RepoRemoteBookmarks) UpdateRemoteBookmark(b types.RemoteBookmark) {
	// Only own bookmarks can be updated. Ownership can't be changed this way. Publishing date too. The id remains.
	mustExec(`
update RemoteBookmarks
set Title = ?, DescriptionHTML = ?, DescriptionMycomarkup = ?, UpdatedAt = ?, Activity = ?, URL = ?
where ID = ?`,
		b.Title, b.DescriptionHTML, b.DescriptionMycomarkup, b.UpdatedAt, b.Activity, b.URL, b.ID)

	mustExec(`delete from RemoteTags where BookmarkID = ?`, b.ID)

	for _, tag := range b.Tags {
		mustExec(`insert or replace into RemoteTags (Name, BookmarkID) values (?, ?)`, tag.Name, b.ID)
	}
}
