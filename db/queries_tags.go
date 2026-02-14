// SPDX-FileCopyrightText: 2024 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package db

import (
	"git.sr.ht/~bouncepaw/betula/types"
)

func deleteTagDescription(tagName string) {
	mustExec(`delete from TagDescriptions where TagName = ?`, tagName)
}

func SetTagDescription(tagName string, description string) {
	const q = `
replace into TagDescriptions (TagName, Description)
values (?, ?);
`
	if description == "" {
		deleteTagDescription(tagName)
	} else {
		mustExec(q, tagName, description)
	}
}

func DeleteTag(tagName string) {
	deleteTagDescription(tagName)
	mustExec(`delete from TagsToPosts where TagName = ?`, tagName)
}

func DescriptionForTag(tagName string) (myco string) {
	rows := mustQuery(`select Description from TagDescriptions where TagName = ?`, tagName)
	for rows.Next() { // 0 or 1
		mustScan(rows, &myco)
		break
	}
	_ = rows.Close()

	return myco
}

// TagCount counts how many tags there are available to the user.
func TagCount(authorized bool) (count uint) {
	q := `
select
	count(distinct TagName)
from
	TagsToPosts
inner join 
	(select ID from Bookmarks where DeletionTime is null and (Visibility = 1 or ?)) 
as 
	Filtered
on 
	TagsToPosts.PostID = Filtered.ID
`
	rows := mustQuery(q, authorized)
	rows.Next()
	mustScan(rows, &count)
	_ = rows.Close()
	return count
}

// Tags returns all tags found on bookmarks one has access to. They all have BookmarkCount set to a non-zero value.
func Tags(authorized bool) (tags []types.Tag) {
	q := `
select
   TagName, 
   count(PostID)
from
   TagsToPosts
inner join 
    (select ID from Bookmarks where DeletionTime is null and (Visibility = 1 or ?)) 
as 
	Filtered
on 
    TagsToPosts.PostID = Filtered.ID
group by
	TagName;
`
	rows := mustQuery(q, authorized)
	for rows.Next() {
		var tag types.Tag
		mustScan(rows, &tag.Name, &tag.BookmarkCount)
		tags = append(tags, tag)
	}
	return tags
}

func TagExists(tagName string) (has bool) {
	const q = `select exists(select 1 from TagsToPosts where TagName = ?);`
	rows := mustQuery(q, tagName)
	rows.Next()
	mustScan(rows, &has)
	_ = rows.Close()
	return has
}

func RenameTag(oldTagName, newTagName string) {
	const q = `
update TagsToPosts
set TagName = ?
where TagName = ?;
`
	mustExec(q, newTagName, oldTagName)
}

func SetTagsFor(bookmarkID int, tags []types.Tag) {
	mustExec(`delete from TagsToPosts where PostID = ?;`, bookmarkID)

	for _, tag := range tags {
		if tag.Name == "" {
			continue
		}
		mustExec(`insert into TagsToPosts (TagName, PostID) values (?, ?);`, tag.Name, bookmarkID)
	}
}

func TagsForBookmarkByID(id int) (tags []types.Tag) {
	rows := mustQuery(`
select distinct TagName
from TagsToPosts
where PostID = ?
order by TagName;
`, id)
	for rows.Next() {
		var tag types.Tag
		mustScan(rows, &tag.Name)
		tags = append(tags, tag)
	}
	return tags
}

func getTagsForManyBookmarks(bookmarks []types.Bookmark) []types.Bookmark {
	for i, post := range bookmarks {
		post.Tags = TagsForBookmarkByID(post.ID)
		bookmarks[i] = post
	}
	return bookmarks
}
