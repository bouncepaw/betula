// Package db encapsulates all used queries to the database.
//
// Do not forget to Initialize and Finalize.
//
// All functions in this package might crash vividly.
package db

import (
	"context"
	"database/sql"
	"git.sr.ht/~bouncepaw/betula/types"
	"log"
	"time"
)

var (
	db *sql.DB
)

const schema = `
create table if not exists posts (
    id integer primary key autoincrement not null,
    url text not null,
    title text not null,
    description text not null,
    visibility integer not null,
    creationTime integer not null                   
);

create table if not exists betula_meta (
    key text primary key,
    value text
);

insert or replace into betula_meta values
	('db version', 0);
`

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
select id, url, title, description, visibility, creationTime from posts;
`

// YieldAllPosts returns a channel, from which you can get all posts stored in the database.
func YieldAllPosts(ctx context.Context) chan types.Post {
	rows, err := db.QueryContext(ctx, sqlGetAllPosts)
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
			out <- post
		}
		close(out)
	}()
	return out
}

const sqlAddPost = `
insert into posts (url, title, description, visibility, creationTime) VALUES (?, ?, ?, ?, ?);
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
select id, url, title, description, visibility, creationTime from posts where id = ?;
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
		return post, true
	}
	return
}

const sqlURLForID = `
select url from posts where id = ?;
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

const sqlLinkCount = `select count(id) from posts;`
const sqlOldestTime = `select min(creationTime) from posts;`
const sqlNewestTime = `select max(creationTime) from posts;`

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
