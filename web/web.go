// Package web provides web capabilities. Import this package to initialize the handlers and the templates.
package web

import (
	"context"
	"embed"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"git.sr.ht/~bouncepaw/betula/db"
	"git.sr.ht/~bouncepaw/betula/types"
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
	http.HandleFunc("/about", handlerAbout)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(fs))))
}

type dataAbout struct {
	LinkCount  int
	OldestTime time.Time
	NewestTime time.Time
}

func handlerAbout(w http.ResponseWriter, rq *http.Request) {
	templateExec(templateAbout, dataAbout{
		LinkCount:  db.LinkCount(),
		OldestTime: db.OldestTime(),
		NewestTime: db.NewestTime(),
	}, w)
}

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
		templateExec(templateAddLink, dataAddLink{
			URL:        rq.FormValue("url"),
			Title:      rq.FormValue("title"),
			Visibility: rq.FormValue("visibility"),
		}, w)
	case http.MethodPost:
		// TODO: Document the param behaviour
		var (
			addr       = rq.FormValue("url")
			title      = rq.FormValue("title")
			visibility = rq.FormValue("visibility")
		)
		if _, err := url.ParseRequestURI(addr); err != nil {
			templateExec(templateAddLinkInvalidURL, dataAddLink{
				URL:        addr,
				Title:      title,
				Visibility: visibility,
			}, w)
			return
		}

		id := db.AddPost(types.Post{
			URL:         addr,
			Title:       title,
			Description: "",
			Visibility:  types.VisibilityFromString(visibility),
		})

		http.Redirect(w, rq, fmt.Sprintf("/%d", id), http.StatusSeeOther)
	}
}

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
	templateExec(templatePost, dataPost{
		Post: post,
	}, w)
}

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
	templateExec(templateFeed, dataFeed{
		YieldAllPosts: db.YieldAllPosts(context.Background()),
	}, w)
}

func handlerGo(w http.ResponseWriter, rq *http.Request) {
	id, err := strconv.Atoi(strings.TrimPrefix(rq.URL.Path, "/go/"))
	if err != nil {
		handlerFeed(w, rq)
		return
	}

	if addr, found := db.URLForID(id); found {
		http.Redirect(w, rq, addr, http.StatusSeeOther)
	} else {
		// TODO: Show 404
		http.Redirect(w, rq, "/", http.StatusSeeOther)
	}
}
