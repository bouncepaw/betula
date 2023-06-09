package db

import (
	"git.sr.ht/~bouncepaw/betula/types"
)

func deleteTagDescription(tagName string) {
	const q = `
delete from TagDescriptions where TagName = ?;
`
	mustExec(q, tagName)
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
	const q = `
delete from TagsToPosts where TagName = ?
`
	deleteTagDescription(tagName)
	mustExec(q, tagName)
}

func DescriptionForTag(tagName string) (myco string) {
	const q = `
select Description from TagDescriptions where TagName = ?;
`
	rows := mustQuery(q, tagName)
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
	(select ID from main.Posts where DeletionTime is null and (Visibility = 1 or ?)) 
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

// Tags returns all tags found on posts one has access to. They all have PostCount set to a non-zero value.
func Tags(authorized bool) (tags []types.Tag) {
	q := `
select
   TagName, 
   count(PostID)
from
   TagsToPosts
inner join 
    (select ID from main.Posts where DeletionTime is null and (Visibility = 1 or ?)) 
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
		mustScan(rows, &tag.Name, &tag.PostCount)
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

func SetTagsFor(postID int, tags []types.Tag) {
	const q = `delete from TagsToPosts where PostID = ?;`
	mustExec(q, postID)

	var qAdd = `insert into TagsToPosts (TagName, PostID) values (?, ?);`
	for _, tag := range tags {
		if tag.Name == "" {
			continue
		}
		mustExec(qAdd, tag.Name, postID)
	}
}

func TagsForPost(id int) (tags []types.Tag) {
	q := `
select distinct TagName
from TagsToPosts
where PostID = ?
order by TagName;
`
	rows := mustQuery(q, id)
	for rows.Next() {
		var tag types.Tag
		mustScan(rows, &tag.Name)
		tags = append(tags, tag)
	}
	return tags
}
