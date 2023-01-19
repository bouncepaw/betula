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

const schema = `
create table if not exists Posts (
    ID integer primary key autoincrement not null,
    URL text not null,
    Title text not null,
    Description text not null,
    Visibility integer not null,
    CreationTime integer not null                   
);

create view if not exists Categories as
select distinct CatName from CategoriesToPosts;

create table if not exists CategoriesToPosts (
    CatName text not null,
    PostID integer not null
);

create table if not exists BetulaMeta (
    Key text primary key,
    Value text
);

insert or ignore into BetulaMeta values
	('DB version', 0),
	('Admin username', null),
	('Admin password hash', null);

create table if not exists Sessions (
    Token text primary key,
    CreationTime integer not null
);`

func AddSession(token string) {
	mustExec(`insert into Sessions values (?, ?);`,
		token, time.Now().Unix())
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

func MetaEntry[T any](key string) T {
	const q = `select Value from BetulaMeta where Key = ? limit 1;`
	return querySingleValue[T](q, key)
}

func AuthorizedPostsForCategory(authorized bool, catName string) (out chan types.Post) {
	const q = `
select
	ID, URL, Title, Description, Visibility, CreationTime
from
	Posts
inner join
	CategoriesToPosts
where
	ID = PostID and CatName = ?
order by
	CreationTime desc;
`
	rows := mustQuery(q, catName)
	out = make(chan types.Post)

	go func() {
		for rows.Next() {
			var post types.Post
			mustScan(rows, &post.ID, &post.URL, &post.Title, &post.Description, &post.Visibility, &post.CreationTime)
			if !authorized && post.Visibility == types.Private {
				continue
			}
			// TODO: Probably can be optimized with a smart query.
			post.Categories = CategoriesForPost(post.ID)
			out <- post
		}
		close(out)
	}()
	return out
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

func Categories() (cats []types.Category) {
	rows := mustQuery(`select CatName from Categories;`)
	for rows.Next() {
		var cat types.Category
		mustScan(rows, &cat.Name)
		cats = append(cats, cat)
	}
	return cats
}

// YieldAuthorizedPosts returns a channel, from which you can get all posts stored in the database, along with their categories, but only if the viewer is authorized! Otherwise, only public posts will be given.
func YieldAuthorizedPosts(authorized bool) chan types.Post {
	const q = `
select ID, URL, Title, Description, Visibility, CreationTime
from Posts
order by CreationTime desc;
`
	rows := mustQuery(q)
	out := make(chan types.Post)

	go func() {
		for rows.Next() {
			var post types.Post
			mustScan(rows, &post.ID, &post.URL, &post.Title, &post.Description, &post.Visibility, &post.CreationTime)
			if !authorized && post.Visibility == types.Private {
				continue
			}
			// TODO: Probably can be optimized with a smart query.
			post.Categories = CategoriesForPost(post.ID)
			out <- post
		}
		close(out)
	}()
	return out
}

// AddPost adds a new post to the database. Creation time is set by this function, ID is set by the database. The ID is returned.
func AddPost(post types.Post) int64 {
	const q = `
insert into Posts (URL, Title, Description, Visibility, CreationTime)
values (?, ?, ?, ?, ?);
`
	post.CreationTime = time.Now().Unix()
	res, err := db.Exec(q, post.URL, post.Title, post.Description, post.Visibility, post.CreationTime)
	if err != nil {
		log.Fatalln(err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		log.Fatalln(err)
	}
	return id
}

func EditPost(post types.Post) {
	const q = `
update Posts
set
    URL = ?,
    Title = ?,
    Description = ?,
    Visibility = ?,
    CreationTime = ?
where
    ID = ?;
`
	mustExec(q, post.URL, post.Title, post.Description, post.Visibility, post.CreationTime, post.ID)
}

// PostForID returns the post corresponding to the given id, if there is any.
func PostForID(id int) (post types.Post, found bool) {
	const q = `
select ID, URL, Title, Description, Visibility, CreationTime from Posts
where ID = ?;
`
	rows := mustQuery(q, id)
	rows.Next()
	mustScan(rows, &post.ID, &post.URL, &post.Title, &post.Description, &post.Visibility, &post.CreationTime)
	_ = rows.Close()
	return post, true
}

// URLForID returns the URL of the post corresponding to the given ID, if there is any post like that.
func URLForID(id int) (url sql.NullString) {
	const q = `select URL from Posts where ID = ?;`
	return querySingleValue[sql.NullString](q, id)
}

func LinkCount() int {
	return querySingleValue[int](`select count(ID) from Posts;`)
}

func OldestTime() *time.Time {
	const q = `select min(CreationTime) from Posts;`
	stamp := querySingleValue[sql.NullInt64](q)
	if stamp.Valid {
		val := time.Unix(stamp.Int64, 0)
		return &val
	}
	return nil
}

func NewestTime() *time.Time {
	const q = `select max(CreationTime) from Posts;`
	stamp := querySingleValue[sql.NullInt64](q)
	if stamp.Valid {
		val := time.Unix(stamp.Int64, 0)
		return &val
	}
	return nil
}
