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

var (
	db *sql.DB
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

create table if not exists Tags (
    ID integer primary key autoincrement not null,
    Name text not null
);

create table if not exists TagsToPosts (
    TagID integer not null,
    PostID integer not null
);

create table if not exists BetulaMeta (
    Key text primary key,
    Value text
);

insert or replace into BetulaMeta values
	('DB version', 0),
	('Admin username', null),
	('Admin password hash', null);`

const sqlTagsForPost = `
select
    TagID, Name
from 
    TagsToPosts
inner join
    Tags
where
    ID = TagID and PostID = ?;
`

func TagsForPost(id int) (tags []types.Tag) {
	rows, err := db.Query(sqlTagsForPost, id)
	if err != nil {
		log.Fatalln(err)
	}
	for rows.Next() {
		var tag types.Tag
		err = rows.Scan(&tag.ID, &tag.Name)
		if err != nil {
			log.Fatalln(err)
		}
		tags = append(tags, tag)
	}
	return tags
}

// Initialize opens a SQLite3 database with the given filename. The connection is encapsulated, you cannot access the database directly, you are to use the functions provided by the package.
func Initialize(filename string) {
	var err error
	db, err = sql.Open("sqlite3", filename)
	if err != nil {
		log.Fatalln(err)
	}
	_, err = db.Exec(schema)
	if err != nil {
		log.Fatalln(err)
	}
}

// Finalize closes the connection with the database.
func Finalize() {
	err := db.Close()
	if err != nil {
		log.Fatalln(err)
	}
}

const sqlGetAllPosts = `
select ID, URL, Title, Description, Visibility, CreationTime from Posts;
`

// YieldAllPosts returns a channel, from which you can get all posts stored in the database, along with their tags.
func YieldAllPosts() chan types.Post {
	rows, err := db.Query(sqlGetAllPosts)
	if err != nil {
		log.Fatalln(err)
	}
	out := make(chan types.Post)

	go func() {
		for rows.Next() {
			var post types.Post
			err = rows.Scan(&post.ID, &post.URL, &post.Title, &post.Description, &post.Visibility, &post.CreationTime)
			if err != nil {
				log.Fatalln(err)
			}
			// TODO: Probably can be optimized with a smart query.
			post.Tags = TagsForPost(post.ID)
			out <- post
		}
		close(out)
	}()
	return out
}

const sqlAddPost = `
insert into Posts (URL, Title, Description, Visibility, CreationTime) VALUES (?, ?, ?, ?, ?);
`

// AddPost adds a new post to the database. Creation time is set by this function, ID is set by the database. The ID is returned.
func AddPost(post types.Post) int64 {
	post.CreationTime = time.Now().Unix()
	res, err := db.Exec(sqlAddPost, post.URL, post.Title, post.Description, post.Visibility, post.CreationTime)
	if err != nil {
		log.Fatalln(err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		log.Fatalln(err)
	}
	return id
}

const sqlPostForID = `
select ID, URL, Title, Description, Visibility, CreationTime from Posts where ID = ?;
`

// PostForID returns the post corresponding to the given id, if there is any.
func PostForID(id int) (post types.Post, found bool) {
	rows, err := db.Query(sqlPostForID, id)
	if err != nil {
		log.Fatalln(err)
	}
	for rows.Next() {
		var post types.Post
		err = rows.Scan(&post.ID, &post.URL, &post.Title, &post.Description, &post.Visibility, &post.CreationTime)
		if err != nil {
			log.Fatalln(err)
		}
		_ = rows.Close()
		return post, true
	}
	return
}

const sqlURLForID = `
select URL from Posts where ID = ?;
`

// URLForID returns the URL of the post corresponding to the given ID, if there is any post like that.
func URLForID(id int) (url string, found bool) {
	rows, err := db.Query(sqlURLForID, id)
	if err != nil {
		log.Fatalln(err)
	}
	for rows.Next() {
		var res string
		err = rows.Scan(&res)
		if err != nil {
			log.Fatalln(err)
		}
		return res, true
	}
	return "", false
}

const sqlLinkCount = `select count(ID) from Posts;`
const sqlOldestTime = `select min(CreationTime) from Posts;`
const sqlNewestTime = `select max(CreationTime) from Posts;`

func LinkCount() int        { return querySingleValue[int](sqlLinkCount) }
func OldestTime() time.Time { return time.Unix(querySingleValue[int64](sqlOldestTime), 0) }
func NewestTime() time.Time { return time.Unix(querySingleValue[int64](sqlNewestTime), 0) }

func querySingleValue[T any](query string) T {
	rows, err := db.Query(query)
	if err != nil {
		log.Fatalln(err)
	}
	rows.Next()
	var res T
	err = rows.Scan(&res)
	if err != nil {
		log.Fatalln(err)
	}
	return res
}
