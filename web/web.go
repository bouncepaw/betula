package web

import (
	"context"
	"embed"
	"git.sr.ht/~bouncepaw/betula/db"
	"git.sr.ht/~bouncepaw/betula/types"
	"html/template"
	"log"
	"net/http"
)

var (
	//go:embed *.gohtml
	fs embed.FS
)

func init() {
	http.HandleFunc("/", HandlerFeed)
	http.HandleFunc("/add-link", HandlerAddLink)
}

func HandlerAddLink(w http.ResponseWriter, rq *http.Request) {
	var (
		url   = rq.FormValue("url")
		title = rq.FormValue("title")
	)
	db.AddPost(context.Background(), types.Post{
		URL:         url,
		Title:       title,
		Description: "",
	})
	http.Redirect(w, rq, "/", http.StatusSeeOther)
}

func HandlerFeed(w http.ResponseWriter, rq *http.Request) {
	t := template.Must(template.ParseFS(fs, "feed.gohtml"))
	err := t.Execute(w, feedData{YieldAllPosts: db.YieldAllPosts(context.Background())})
	if err != nil {
		log.Fatalln(err)
	}
}

type feedData struct {
	YieldAllPosts chan types.Post
}
