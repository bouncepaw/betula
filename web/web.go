// Package web provides web capabilities. Import this package to initialize the handlers and the templates.
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
	"net/url"
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
	http.HandleFunc("/post/", handlerPost)
	http.HandleFunc("/go/", handlerGo)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(fs))))
}

var templateAddLink = template.Must(template.New("skeleton.gohtml").Funcs(template.FuncMap{}).ParseFS(fs, "add-link.gohtml", "skeleton.gohtml"))
var templateAddLinkInvalidURL = template.Must(template.New("skeleton.gohtml").ParseFS(fs, "add-link-invalid-url.gohtml", "skeleton.gohtml"))

type dataAddLink struct {
	Authorized bool // TODO: authorize

	// The following three fields can be non-empty, when set through URL parameters or when an erroneous request was made.

	URL        string
	Title      string
	Visibility string
}

func handlerAddLink(w http.ResponseWriter, rq *http.Request) {
	switch rq.Method {
	case http.MethodGet:
		err := templateAddLink.ExecuteTemplate(
			w,
			"skeleton.gohtml",
			dataAddLink{
				URL:        rq.FormValue("url"),
				Title:      rq.FormValue("title"),
				Visibility: rq.FormValue("visibility"),
			})
		if err != nil {
			log.Fatalln(err)
		}
	case http.MethodPost:
		// TODO: Document the param behaviour
		addr := rq.FormValue("url")
		title := rq.FormValue("title")
		visibility := rq.FormValue("visibility")

		if _, err := url.ParseRequestURI(addr); err != nil {
			err := templateAddLinkInvalidURL.ExecuteTemplate(
				w,
				"skeleton.gohtml",
				dataAddLink{
					URL:        addr,
					Title:      title,
					Visibility: visibility,
				})
			if err != nil {
				log.Fatalln(err)
			}
			return
		}

		var (
			post = types.Post{
				URL:         addr,
				Title:       title,
				Description: "",
				Visibility:  types.VisibilityFromString(visibility),
			}

			id = db.AddPost(post)
		)

		http.Redirect(w, rq, fmt.Sprintf("/%d", id), http.StatusSeeOther)
	}
}

var templatePost = template.Must(template.New("skeleton.gohtml").Funcs(template.FuncMap{
	"randomGlobe": func() string {
		return string([]rune{[]rune("🌍🌎🌏")[rand.Intn(3)]})
	},
	"timestampToHuman": func(stamp int64) string {
		t := time.Unix(stamp, 0)
		return t.Format("2006-01-02 15:04")
	},
}).ParseFS(fs, "post.gohtml", "skeleton.gohtml"))

type dataPost struct {
	Post       types.Post
	Authorized bool // TODO: authorize
}

func handlerPost(w http.ResponseWriter, rq *http.Request) {
	id, err := strconv.Atoi(strings.TrimPrefix(strings.TrimPrefix(rq.URL.Path, "/"), "post/"))
	if err != nil {
		// TODO: Show 404
		log.Println(err)
		handlerFeed(w, rq)
		return
	}
	log.Printf("Viewing post %d\n", id)
	post, found := db.PostForID(id)
	if !found {
		// TODO: Show 404
		log.Println(err)
		handlerFeed(w, rq)
		return
	}
	err = templatePost.ExecuteTemplate(
		w,
		"skeleton.gohtml",
		dataPost{
			Post: post,
		})
	if err != nil {
		log.Fatalln(err)
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

var regexpPost = regexp.MustCompile("^/[0-9]+")

func handlerFeed(w http.ResponseWriter, rq *http.Request) {
	if regexpPost.MatchString(rq.URL.Path) {
		handlerPost(w, rq)
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

func handlerGo(w http.ResponseWriter, rq *http.Request) {
	id, err := strconv.Atoi(strings.TrimPrefix(rq.URL.Path, "/go/"))
	if err != nil {
		handlerFeed(w, rq)
		return
	}

	if url, found := db.URLForID(id); found {
		http.Redirect(w, rq, url, http.StatusSeeOther)
	} else {
		// TODO: Show 404
		http.Redirect(w, rq, "/", http.StatusSeeOther)
	}
}
