package db

import (
	"fmt"
	"git.sr.ht/~bouncepaw/betula/types"
	"sort"
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

func Search(text string, includedTags []string, excludedTags []string, authorized bool) (posts []types.Post) {
	sort.Strings(includedTags)
	sort.Strings(excludedTags)

	const q = `
select ID, URL, Title, Description, Visibility, CreationTime
from Posts
where DeletionTime is null and (Visibility = 1 or ?) and (
	Title like ? or URL like ? or Description like ?
)
order by CreationTime desc;
`
	text = fmt.Sprintf("%%%s%%", text)
	rows := mustQuery(q, authorized, text, text, text)

	var unfilteredPosts []types.Post
	for rows.Next() {
		var post types.Post
		mustScan(rows, &post.ID, &post.URL, &post.Title, &post.Description, &post.Visibility, &post.CreationTime)
		unfilteredPosts = append(unfilteredPosts, post)
	}

	// ‘Say, Bouncepaw, why did not you implement tag inclusion/exclusion
	//  part in SQL directly?’, some may ask.
	// ‘I did, and it was not worth it’, so I would respond.
	for _, post := range unfilteredPosts {
		post.Tags = TagsForPost(post.ID)
		if keepForSearch(post.Tags, includedTags, excludedTags) {
			posts = append(posts, post)
		}
	}
	return posts
}

// true if keep, false if discard. All slices are sorted.
func keepForSearch(postTags []types.Tag, includedTags, excludedTags []string) bool {
	J, K := len(includedTags), len(excludedTags)
	j, k := 0, 0
	includeMask := make([]int, J)
	for _, postTag := range postTags {
		name := postTag.Name
		switch {
		case k < K && excludedTags[k] == name:
			return false
		case j < J && includedTags[j] == name:
			includeMask[j] = 1
			j++
			continue
		}

		for j < J && includedTags[j] < name {
			j++
		}

		for k < K && excludedTags[k] < name {
			k++
		}
	}

	for _, marker := range includeMask {
		if marker == 0 {
			return false
		}
	}
	return true
}
