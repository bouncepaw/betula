package db

import "git.sr.ht/~bouncepaw/betula/types"

func deleteTagDescription(catName string) {
	const q = `
delete from CategoryDescriptions where CatName = ?;
`
	mustExec(q, catName)
}

func SetTagDescription(tagName string, description string) {
	const q = `
replace into CategoryDescriptions (CatName, Description)
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
delete from CategoriesToPosts where CategoriesToPosts.CatName = ?
`
	deleteTagDescription(tagName)
	mustExec(q, tagName)
}

func DescriptionForTag(tagName string) (myco string) {
	const q = `
select Description from CategoryDescriptions where CatName = ?;
`
	rows := mustQuery(q, tagName)
	for rows.Next() { // 0 or 1
		mustScan(rows, &myco)
		break
	}
	_ = rows.Close()

	return myco
}

// Tags returns all tags found on posts one has access to. They all have PostCount set to a non-zero value.
func Tags(authorized bool) (tags []types.Tag) {
	q := `
select
   CatName, 
   count(PostID)
from
   CategoriesToPosts
inner join 
    (select ID from main.Posts where DeletionTime is null and (Visibility = 1 or ?)) 
as 
	Filtered
on 
    CategoriesToPosts.PostID = Filtered.ID
group by
	CatName;
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
	const q = `select exists(select 1 from CategoriesToPosts where CatName = ?);`
	rows := mustQuery(q, tagName)
	rows.Next()
	mustScan(rows, &has)
	_ = rows.Close()
	return has
}

func RenameTag(oldTagName, newTagName string) {
	const q = `
update CategoriesToPosts
set CatName = ?
where CatName = ?;
`
	mustExec(q, newTagName, oldTagName)
}

func SetTagsFor(postID int, tags []types.Tag) {
	const q = `delete from CategoriesToPosts where PostID = ?;`
	mustExec(q, postID)

	var qAdd = `insert into CategoriesToPosts (CatName, PostID) values (?, ?);`
	for _, tag := range tags {
		if tag.Name == "" {
			continue
		}
		mustExec(qAdd, tag.Name, postID)
	}
}

func TagsForPost(id int) (tags []types.Tag) {
	q := `
select distinct CatName
from CategoriesToPosts
where PostID = ?
order by CatName;
`
	rows := mustQuery(q, id)
	for rows.Next() {
		var tag types.Tag
		mustScan(rows, &tag.Name)
		tags = append(tags, tag)
	}
	return tags
}
