package db

import (
	"git.sr.ht/~bouncepaw/betula/types"
	"log"
)

// PostsForDay returns posts for the given dayStamp, which looks like this: 2023-03-14. The result might as well be nil, that means there are no posts for the day.
func PostsForDay(authorized bool, dayStamp string) (posts []types.Bookmark) {
	const q = `
select
	ID, URL, Title, Description, Visibility, CreationTime, RepostOf
from
	Posts
where
	DeletionTime is null and (Visibility = 1 or ?) and CreationTime like ?
order by
	CreationTime desc;
`
	rows := mustQuery(q, authorized, dayStamp+"%")
	for rows.Next() {
		var post types.Bookmark
		mustScan(rows, &post.ID, &post.URL, &post.Title, &post.Description, &post.Visibility, &post.CreationTime, &post.RepostOf)
		posts = append(posts, post)
	}
	for i, post := range posts {
		post.Tags = TagsForPost(post.ID)
		posts[i] = post
	}
	return posts
}

func PostsWithTag(authorized bool, tagName string, page uint) (posts []types.Bookmark, totalPosts uint) {
	totalPosts = querySingleValue[uint](`
select
	count(ID)
from
	Posts
inner join
	TagsToPosts
where
	ID = PostID and TagName = ? and DeletionTime is null and (Visibility = 1 or ?)
`, tagName, authorized)

	const q = `
select
	ID, URL, Title, Description, Visibility, CreationTime, RepostOf
from
	Posts
inner join
	TagsToPosts
where
	ID = PostID and TagName = ? and DeletionTime is null and (Visibility = 1 or ?)
order by
	CreationTime desc
limit ? offset ?;
`
	rows := mustQuery(q, tagName, authorized, types.PostsPerPage, types.PostsPerPage*(page-1))
	for rows.Next() {
		var post types.Bookmark
		mustScan(rows, &post.ID, &post.URL, &post.Title, &post.Description, &post.Visibility, &post.CreationTime, &post.RepostOf)
		posts = append(posts, post)
	}
	for i, post := range posts {
		post.Tags = TagsForPost(post.ID)
		posts[i] = post
	}
	return posts, totalPosts
}

// Posts returns all posts stored in the database, along with their tags, but only if the viewer is authorized! Otherwise, only public posts will be given.
func Posts(authorized bool, page uint) (posts []types.Bookmark, totalPosts uint) {
	if page == 0 {
		panic("page 0 makes no sense")
	}

	totalPosts = querySingleValue[uint](`
select count(ID)
from Posts
where DeletionTime is null and (Visibility = 1 or ?);
`, authorized)

	const q = `
select ID, URL, Title, Description, Visibility, CreationTime, RepostOf
from Posts
where DeletionTime is null
order by CreationTime desc
limit ?
offset (? * (? - 1));
`
	rows := mustQuery(q, types.PostsPerPage, types.PostsPerPage, page) // same paging for remote bookmarks

	for rows.Next() {
		var post types.Bookmark
		mustScan(rows, &post.ID, &post.URL, &post.Title, &post.Description, &post.Visibility, &post.CreationTime, &post.RepostOf)
		if !authorized && post.Visibility == types.Private {
			continue
		}
		posts = append(posts, post)
	}
	for i, post := range posts {
		post.Tags = TagsForPost(post.ID)
		posts[i] = post
	}
	return posts, totalPosts
}

func HasPost(id int) (has bool) {
	const q = `select exists(select 1 from Posts where ID = ? and DeletionTime is null);`
	rows := mustQuery(q, id)
	rows.Next()
	mustScan(rows, &has)
	_ = rows.Close()
	return has
}

func DeletePost(id int) {
	const q = `
update Posts
set DeletionTime = current_timestamp
where ID = ?;
`
	mustExec(q, id)
}

// AddPost adds a new original post to the database. Creation time is set by this function, ID is set by the database. The ID is returned.
func AddPost(post types.Bookmark) int64 {
	const q = `
insert into Posts (URL, Title, Description, Visibility, RepostOf)
values (?, ?, ?, ?, ?);
`
	res, err := db.Exec(q, post.URL, post.Title, post.Description, post.Visibility, post.RepostOf)
	if err != nil {
		log.Fatalln(err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		log.Fatalln(err)
	}
	post.ID = int(id)
	SetTagsFor(post.ID, post.Tags)
	return id
}

func EditPost(post types.Bookmark) {
	const q = `
update Posts
set
    URL = ?,
    Title = ?,
    Description = ?,
    Visibility = ?,
	RepostOf = ?
where
    ID = ? and DeletionTime is null;
`
	mustExec(q, post.URL, post.Title, post.Description, post.Visibility, post.RepostOf, post.ID)
	SetTagsFor(post.ID, post.Tags)
}

// PostForID returns the post corresponding to the given id, if there is any.
func PostForID(id int) (post types.Bookmark, found bool) {
	const q = `
select ID, URL, Title, Description, Visibility, CreationTime, RepostOf from Posts
where ID = ? and DeletionTime is null
limit 1;
`
	rows := mustQuery(q, id)
	for rows.Next() {
		mustScan(rows, &post.ID, &post.URL, &post.Title, &post.Description, &post.Visibility, &post.CreationTime, &post.RepostOf)
		found = true
	}
	return post, found
}

func BookmarkCount(authorized bool) uint {
	const q = `
with
	IgnoredPosts as (
	   -- Ignore deleted posts always
		select ID from Posts where DeletionTime is not null
	   union
		-- Ignore private posts if so desired
	   select ID from Posts where Visibility = 0 and not ?
	)
select 
	count(ID)
from 
	Posts 
where 
	ID not in IgnoredPosts;
`
	return querySingleValue[uint](q, authorized)
}

func LastPost(authorized bool) (post types.Bookmark, found bool) {
	const q = `
with
	IgnoredPosts as (
	   -- Ignore deleted posts always
		select ID from Posts where DeletionTime is not null
	   union
		-- Ignore private posts if so desired
	   select ID from Posts where Visibility = 0 and not ?
	)
select 
    ID, URL, Title, Description, Visibility, CreationTime, RepostOf 
from 
    Posts 
where 
    ID not in IgnoredPosts
order by 
    CreationTime desc 
limit 1;
`
	rows := mustQuery(q, authorized)
	for rows.Next() {
		mustScan(rows, &post.ID, &post.URL, &post.Title, &post.Description, &post.Visibility, &post.CreationTime, &post.RepostOf)
		found = true
	}
	return post, found
}
