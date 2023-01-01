package web

import (
	"context"
	"embed"
	"git.sr.ht/~bouncepaw/betula/db"
	"git.sr.ht/~bouncepaw/betula/types"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"time"
)

var (
	//go:embed *.gohtml *.css
	fs embed.FS
)

func init() {
	http.HandleFunc("/", HandlerFeed)
	http.HandleFunc("/add-link", HandlerAddLink)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(fs))))
}

func HandlerAddLink(w http.ResponseWriter, rq *http.Request) {
	var (
		url      = rq.FormValue("url")
		title    = rq.FormValue("title")
		isPublic = rq.FormValue("isPublic") == "public"
	)
	db.AddPost(context.Background(), types.Post{
		URL:         url,
		Title:       title,
		Description: "",
		IsPublic:    isPublic,
	})
	http.Redirect(w, rq, "/", http.StatusSeeOther)
}

func HandlerFeed(w http.ResponseWriter, rq *http.Request) {
	t := template.Must(template.New("feed.gohtml").Funcs(template.FuncMap{
		"randomGlobe": func() string {
			return string([]rune{[]rune("🌍🌎🌏")[rand.Intn(3)]})
		},
		"timestampToHuman": func(stamp int64) string {
			t := time.Unix(stamp, 0)
			return t.Format("2006-01-02 15:04")
		},
	}).ParseFS(fs, "feed.gohtml"))
	err := t.Execute(
		w,
		feedData{
			YieldAllPosts: db.YieldAllPosts(context.Background()),
		},
	)
	if err != nil {
		log.Fatalln(err)
	}
}

type feedData struct {
	YieldAllPosts chan types.Post
}
