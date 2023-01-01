package web

import (
	"context"
	"embed"
	"fmt"
	"git.sr.ht/~bouncepaw/betula/db"
	"git.sr.ht/~bouncepaw/betula/types"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	//go:embed *.gohtml *.css
	fs embed.FS
)

func init() {
	http.HandleFunc("/", handlerFeed)
	http.HandleFunc("/add-link", handlerAddLink)
	http.HandleFunc("/link/", handlerLink)
	http.HandleFunc("/go/", handlerGo)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(fs))))
}

var templateAddLink = template.Must(template.New("skeleton.gohtml").Funcs(template.FuncMap{}).ParseFS(fs, "add-link.gohtml", "skeleton.gohtml"))

type dataAddLink struct {
	Authorized bool // TODO: authorize
}

func handlerAddLink(w http.ResponseWriter, rq *http.Request) {
	switch rq.Method {
	case http.MethodGet:
		err := templateAddLink.ExecuteTemplate(
			w,
			"skeleton.gohtml",
			dataAddLink{})
		if err != nil {
			log.Fatalln(err)
		}
	case http.MethodPost:
		var (
			post = types.Post{
				URL:         rq.FormValue("url"),
				Title:       rq.FormValue("title"),
				Description: "",
				Visibility:  types.VisibilityFromString(rq.FormValue("visibility")),
			}

			id = db.AddPost(context.Background(), post)
		)

		http.Redirect(w, rq, fmt.Sprintf("/%d", id), http.StatusSeeOther)
	}
}

var templateFeed = template.Must(template.New("skeleton.gohtml").Funcs(template.FuncMap{
	"randomGlobe": func() string {
		return string([]rune{[]rune("🌍🌎🌏")[rand.Intn(3)]})
	},
	"timestampToHuman": func(stamp int64) string {
		t := time.Unix(stamp, 0)
		return t.Format("2006-01-02 15:04")
	},
}).ParseFS(fs, "feed.gohtml", "skeleton.gohtml"))

type dataFeed struct {
	YieldAllPosts chan types.Post
	Authorized    bool // TODO: authorize
}

var regexpPost = regexp.MustCompile("^/%d+")

func handlerFeed(w http.ResponseWriter, rq *http.Request) {
	// This handler also routes away URL:s like /%d.
	if regexpPost.MatchString(rq.URL.Path) {
		handlerLink(w, rq)
		return
	}

	err := templateFeed.ExecuteTemplate(
		w,
		"skeleton.gohtml",
		dataFeed{
			YieldAllPosts: db.YieldAllPosts(context.Background()),
		},
	)
	if err != nil {
		log.Fatalln(err)
	}
}

func handlerLink(w http.ResponseWriter, rq *http.Request) {
	id, err := strconv.Atoi(strings.TrimPrefix(rq.URL.Path, "/link/"))
	if err != nil {
		handlerFeed(w, rq)
		return
	}
	log.Println(id)
	// TODO: Implement
}

func handlerGo(w http.ResponseWriter, rq *http.Request) {
	id, err := strconv.Atoi(strings.TrimPrefix(rq.URL.Path, "/link/"))
	if err != nil {
		handlerFeed(w, rq)
		return
	}
	log.Println(id)
	// TODO: get URL for ID
	http.Redirect(w, rq, "", http.StatusSeeOther)
}
