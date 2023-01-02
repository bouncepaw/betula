// Package db encapsulates all used queries to the database.
//
// Do not forget to Initialize and Finalize.
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

func Finalize() {
	_ = db.Close()
}

const sqlGetAllPosts = `
select id, url, title, description, visibility, creationTime from posts;
`

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

func AddPost(ctx context.Context, post types.Post) int64 {
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
