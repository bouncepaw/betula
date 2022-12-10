package db

import (
	"context"
	"database/sql"
	"git.sr.ht/~bouncepaw/betula/types"
	"log"
)

var (
	db *sql.DB
)

const schema = `
create table if not exists posts (
    id integer primary key autoincrement,
    url text,
    title text,
    description text
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
select id, url, title, description from posts;
`

func YieldAllPosts(ctx context.Context) chan types.Post {
	rows, err := db.QueryContext(ctx, sqlGetAllPosts)
	if err != nil {
		log.Println("tapa")
		log.Fatalln(err)
	}
	out := make(chan types.Post)

	go func() {
		for rows.Next() {
			var post types.Post
			err = rows.Scan(&post.ID, &post.URL, &post.Title, &post.Description)
			if err != nil {
				log.Println("muco")
				log.Fatalln(err)
			}
			out <- post
		}
		close(out)
	}()
	return out
}

const sqlAddPost = `
insert into posts (url, title, description) VALUES (?, ?, ?);
`

func AddPost(ctx context.Context, post types.Post) {
	_, err := db.Exec(sqlAddPost, post.URL, post.Title, post.Description)
	if err != nil {
		log.Fatalln(err)
	}
}
