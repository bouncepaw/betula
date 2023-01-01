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
    isPublic integer not null,
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
select id, url, title, description, isPublic, creationTime from posts;
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
			err = rows.Scan(&post.ID, &post.URL, &post.Title, &post.Description, &post.IsPublic, &post.CreationTime)
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
insert into posts (url, title, description, isPublic, creationTime) VALUES (?, ?, ?, ?, ?);
`

func AddPost(ctx context.Context, post types.Post) {
	post.CreationTime = time.Now().Unix()
	_, err := db.Exec(sqlAddPost, post.URL, post.Title, post.Description, post.IsPublic, post.CreationTime)
	if err != nil {
		log.Fatalln(err)
	}
}
