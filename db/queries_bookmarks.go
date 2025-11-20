// SPDX-FileCopyrightText: 2022-2025 Betula contributors
//
// SPDX-License-Identifier: AGPL-3.0-only

package db

import (
	"git.sr.ht/~bouncepaw/betula/types"
	"log"
)

// BookmarksForDay returns bookmarks for the given dayStamp, which looks like this: 2023-03-14. The result might as well be nil, that means there are no bookmarks for the day.
func BookmarksForDay(authorized bool, dayStamp string) (bookmarks []types.Bookmark) {
	const q = `
select
	ID, URL, Title, Description, Visibility, CreationTime, RepostOf, OriginalAuthorID
from
	Bookmarks
where
	DeletionTime is null and (Visibility = 1 or ?) and CreationTime like ?
order by
	CreationTime desc;
`
	rows := mustQuery(q, authorized, dayStamp+"%")
	for rows.Next() {
		var bm types.Bookmark
		mustScan(rows, &bm.ID, &bm.URL, &bm.Title, &bm.Description, &bm.Visibility, &bm.CreationTime, &bm.RepostOf, &bm.OriginalAuthor)
		bookmarks = append(bookmarks, bm)
	}

	return getTagsForManyBookmarks(bookmarks)
}

func BookmarksWithTag(authorized bool, tagName string, page uint) (bookmarks []types.Bookmark, total uint) {
	total = querySingleValue[uint](`
select
	count(ID)
from
	Bookmarks
inner join
	TagsToPosts
where
	ID = PostID and TagName = ? and DeletionTime is null and (Visibility = 1 or ?)
`, tagName, authorized)

	const q = `
select
	ID, URL, Title, Description, Visibility, CreationTime, RepostOf, OriginalAuthorID
from
	Bookmarks
inner join
	TagsToPosts
where
	ID = PostID and TagName = ? and DeletionTime is null and (Visibility = 1 or ?)
order by
	CreationTime desc
limit ? offset ?;
`
	rows := mustQuery(q, tagName, authorized, types.BookmarksPerPage, types.BookmarksPerPage*(page-1))
	for rows.Next() {
		var bm types.Bookmark
		mustScan(rows, &bm.ID, &bm.URL, &bm.Title, &bm.Description, &bm.Visibility, &bm.CreationTime, &bm.RepostOf, &bm.OriginalAuthor)
		bookmarks = append(bookmarks, bm)
	}

	return getTagsForManyBookmarks(bookmarks), total
}

// Bookmarks returns all bookmarks stored in the database on the given page, along with their tags, but only if the viewer is authorized! Otherwise, only public bookmarks will be given.
func Bookmarks(authorized bool, page uint) (bookmarks []types.Bookmark, total uint) {
	if page == 0 {
		panic("page 0 makes 0 sense")
	}

	total = querySingleValue[uint](`
select count(ID)
from Bookmarks
where DeletionTime is null and (Visibility = 1 or ?);
`, authorized)

	const q = `
select ID, URL, Title, Description, Visibility, CreationTime, RepostOf, OriginalAuthorID
from Bookmarks
where DeletionTime is null
order by CreationTime desc
limit ?
offset (? * (? - 1));
`
	rows := mustQuery(q, types.BookmarksPerPage, types.BookmarksPerPage, page) // same paging for remote bookmarks
	for rows.Next() {
		var bm types.Bookmark
		mustScan(rows, &bm.ID, &bm.URL, &bm.Title, &bm.Description, &bm.Visibility, &bm.CreationTime, &bm.RepostOf, &bm.OriginalAuthor)
		if !authorized && bm.Visibility == types.Private {
			continue
		}
		bookmarks = append(bookmarks, bm)
	}

	return getTagsForManyBookmarks(bookmarks), total
}

func RandomBookmarks(authorized bool, n uint) (bookmarks []types.Bookmark, total uint) {
	const q = `
select * from 
(
	select ID, URL, Title, Description, Visibility, CreationTime, RepostOf, OriginalAuthorID 
	from Bookmarks 
	where DeletionTime is null and (Visibility = 1 or ?)
	order by random() limit ?
)
order by CreationTime desc;`

	rows := mustQuery(q, authorized, n)
	for rows.Next() {
		var bm types.Bookmark
		mustScan(rows, &bm.ID, &bm.URL, &bm.Title, &bm.Description, &bm.Visibility, &bm.CreationTime, &bm.RepostOf, &bm.OriginalAuthor)
		bookmarks = append(bookmarks, bm)
	}
	return getTagsForManyBookmarks(bookmarks), uint(len(bookmarks))
}

func DeleteBookmark(id int) {
	mustExec(`update Bookmarks set DeletionTime = current_timestamp where ID = ?`, id)
}

// InsertBookmark adds a new local bookmark to the database. Creation time is set by this function, ID is set by the database. The ID is returned.
func InsertBookmark(bookmark types.Bookmark) int64 {
	const q = `
insert into Bookmarks (URL, Title, Description, Visibility, RepostOf, OriginalAuthorID)
values (?, ?, ?, ?, ?, ?);
`
	res, err := db.Exec(q, bookmark.URL, bookmark.Title, bookmark.Description, bookmark.Visibility, bookmark.RepostOf, bookmark.OriginalAuthor)
	if err != nil {
		log.Fatalln(err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		log.Fatalln(err)
	}
	bookmark.ID = int(id)
	SetTagsFor(bookmark.ID, bookmark.Tags)
	return id
}

func EditBookmark(bookmark types.Bookmark) {
	const q = `
update Bookmarks
set
    URL = ?,
    Title = ?,
    Description = ?,
    Visibility = ?,
	RepostOf = ?,
    OriginalAuthorID = ?
where
    ID = ? and DeletionTime is null;
`
	mustExec(q, bookmark.URL, bookmark.Title, bookmark.Description, bookmark.Visibility, bookmark.RepostOf, bookmark.OriginalAuthor, bookmark.ID)
	SetTagsFor(bookmark.ID, bookmark.Tags)
}

// GetBookmarkByID returns the bookmark corresponding to the given id, if there is any.
func GetBookmarkByID(id int) (b types.Bookmark, found bool) {
	const q = `
select ID, URL, Title, Description, Visibility, CreationTime, RepostOf, OriginalAuthorID from Bookmarks
where ID = ? and DeletionTime is null
limit 1;
`
	rows := mustQuery(q, id)
	for rows.Next() {
		mustScan(rows, &b.ID, &b.URL, &b.Title, &b.Description, &b.Visibility, &b.CreationTime, &b.RepostOf, &b.OriginalAuthor)
		found = true
	}
	return b, found
}

// GetBookmarkIDByURL returns the first bookmark ID with the given URL, if any.
func GetBookmarkIDByURL(url string) (id int, found bool) {
	const q = `
select ID 
from Bookmarks 
where URL = ? and DeletionTime is null
limit 1;`
	rows := mustQuery(q, url)
	for rows.Next() {
		mustScan(rows, &id)
		found = true
	}
	return id, found
}

func BookmarkCount(authorized bool) uint {
	const q = `
with
	IgnoredBookmarks as (
		-- Ignore deleted bookmarks always
		select ID from Bookmarks where DeletionTime is not null
		union
		-- Ignore private bookmarks if so desired
		select ID from Bookmarks where Visibility = 0 and not ?
	)
select 
	count(ID)
from 
	Bookmarks 
where 
	ID not in IgnoredBookmarks;
`
	return querySingleValue[uint](q, authorized)
}
