// Package db encapsulates all used queries to the database.
//
// Do not forget to Initialize and Finalize.
//
// All functions in this package might crash vividly.
package db

import (
	"database/sql"
	"git.sr.ht/~bouncepaw/betula/types"
	"log"
	"time"
)

func AddSession(token string) {
	mustExec(`insert into Sessions values (?, ?);`,
		token, time.Now())
}

func HasSession(token string) (has bool) {
	const q = `select exists(select 1 from Sessions where Token = ?);`
	rows := mustQuery(q, token)
	rows.Next()
	mustScan(rows, &has)
	_ = rows.Close()
	return has
}

func StopSession(token string) {
	mustExec(`delete from Sessions where Token = ?;`, token)
}

func SetCredentials(name, hash string) {
	const q = `
insert or replace into BetulaMeta values
	('Admin username', ?),
	('Admin password hash', ?);
`
	mustExec(q, name, hash)
}

func MetaEntry[T any](key BetulaMetaKey) T {
	const q = `select Value from BetulaMeta where Key = ? limit 1;`
	return querySingleValue[T](q, key)
}

func SetMetaEntry[T any](key BetulaMetaKey, val T) {
	const q = `insert or replace into BetulaMeta values (?, ?);`
	mustExec(q, key, val)
}

func AuthorizedPostsForCategory(authorized bool, catName string) (posts []types.Post) {
	const q = `
select
	ID, URL, Title, Description, Visibility, CreationTime
from
	Posts
inner join
	CategoriesToPosts
where
	ID = PostID and CatName = ? and DeletionTime is null
order by
	CreationTime desc;
`
	rows := mustQuery(q, catName)
	for rows.Next() {
		var post types.Post
		mustScan(rows, &post.ID, &post.URL, &post.Title, &post.Description, &post.Visibility, &post.CreationTime)
		if !authorized && post.Visibility == types.Private {
			continue
		}
		posts = append(posts, post)
	}
	for i, post := range posts {
		post.Categories = CategoriesForPost(post.ID)
		posts[i] = post
	}
	return posts
}

func CategoriesForPost(id int) (cats []types.Category) {
	q := `
select distinct CatName
from CategoriesToPosts
where PostID = ?;
`
	rows := mustQuery(q, id)
	for rows.Next() {
		var cat types.Category
		mustScan(rows, &cat.Name)
		cats = append(cats, cat)
	}
	return cats
}

// Categories returns all categories found on posts one has access to. They all have PostCount set to a non-zero value.
func Categories(authorized bool) (cats []types.Category) {
	q := `
with
	IgnoredPosts as (
	   -- Ignore deleted posts always
		select ID from Posts where DeletionTime is not null
	   union
		-- Ignore private posts if so desired
	   select ID from Posts where Visibility = 0 and not ?
	)
select
   CatName, 
   count(PostID)
from
   CategoriesToPosts
where
   PostID not in IgnoredPosts
group by
	CatName;
`
	rows := mustQuery(q, authorized)
	for rows.Next() {
		var cat types.Category
		mustScan(rows, &cat.Name, &cat.PostCount)
		cats = append(cats, cat)
	}
	return cats
}

// AuthorizedPosts returns all posts stored in the database, along with their categories, but only if the viewer is authorized! Otherwise, only public posts will be given.
func AuthorizedPosts(authorized bool) (posts []types.Post) {
	const q = `
select ID, URL, Title, Description, Visibility, CreationTime
from Posts
where DeletionTime is null
order by CreationTime desc;
`
	rows := mustQuery(q)

	for rows.Next() {
		var post types.Post
		mustScan(rows, &post.ID, &post.URL, &post.Title, &post.Description, &post.Visibility, &post.CreationTime)
		if !authorized && post.Visibility == types.Private {
			continue
		}
		posts = append(posts, post)
	}
	for i, post := range posts {
		post.Categories = CategoriesForPost(post.ID)
		posts[i] = post
	}
	return posts
}

func SetCategoriesFor(postID int, categories []types.Category) {
	const qDelete = `delete from CategoriesToPosts where PostID = ?;`
	mustExec(qDelete, postID)

	var qAdd = `insert into CategoriesToPosts (CatName, PostID) values (?, ?);`
	for _, cat := range categories {
		if cat.Name == "" {
			continue
		}
		mustExec(qAdd, cat.Name, postID)
	}
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

// AddPost adds a new post to the database. Creation time is set by this function, ID is set by the database. The ID is returned.
func AddPost(post types.Post) int64 {
	const q = `
insert into Posts (URL, Title, Description, Visibility)
values (?, ?, ?, ?);
`
	res, err := db.Exec(q, post.URL, post.Title, post.Description, post.Visibility)
	if err != nil {
		log.Fatalln(err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		log.Fatalln(err)
	}
	post.ID = int(id)
	SetCategoriesFor(post.ID, post.Categories)
	return id
}

func EditCategory(category types.Category, newName string) {
	const q = `
update CategoriesToPosts
set
    CatName = ?
where
    CatName = ?;
`
	mustExec(q, newName, category.Name)
}

func HasCategory(category types.Category) (has bool) {
	const q = `select exists(select 1 from CategoriesToPosts where CatName = ?);`
	rows := mustQuery(q, category.Name)
	rows.Next()
	mustScan(rows, &has)
	_ = rows.Close()
	return has
}

func EditPost(post types.Post) {
	const q = `
update Posts
set
    URL = ?,
    Title = ?,
    Description = ?,
    Visibility = ?
where
    ID = ? and DeletionTime is null;
`
	mustExec(q, post.URL, post.Title, post.Description, post.Visibility, post.ID)
	SetCategoriesFor(post.ID, post.Categories)
}

// PostForID returns the post corresponding to the given id, if there is any.
func PostForID(id int) (post types.Post, found bool) {
	const q = `
select ID, URL, Title, Description, Visibility, CreationTime from Posts
where ID = ? and DeletionTime is null
limit 1;
`
	rows := mustQuery(q, id)
	for rows.Next() {
		mustScan(rows, &post.ID, &post.URL, &post.Title, &post.Description, &post.Visibility, &post.CreationTime)
		found = true
	}
	return post, found
}

// URLForID returns the URL of the post corresponding to the given ID, if there is any post like that.
func URLForID(id int) (url sql.NullString) {
	const q = `select URL from Posts where ID = ? and DeletionTime is null;`
	return querySingleValue[sql.NullString](q, id)
}

func LinkCount(authorized bool) int {
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
	return querySingleValue[int](q, authorized)
}

func OldestTime() *time.Time {
	const q = `select min(CreationTime) from Posts;`
	stamp := querySingleValue[sql.NullString](q)
	if stamp.Valid {
		val, err := time.Parse("2006-01-02 15:04:05", stamp.String)
		if err != nil {
			log.Fatalln(err)
		}
		return &val
	}
	return nil
}

func NewestTime() *time.Time {
	const q = `select max(CreationTime) from Posts;`
	stamp := querySingleValue[sql.NullString](q)
	if stamp.Valid {
		val, err := time.Parse(types.TimeLayout, stamp.String)
		if err != nil {
			log.Fatalln(err)
		}
		return &val
	}
	return nil
}
