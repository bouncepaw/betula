// Package web provides web capabilities. Import this package to initialize the handlers and the templates.
package web

import (
	"embed"
	"fmt"
	"git.sr.ht/~bouncepaw/betula/auth"
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
	fs  embed.FS
	mux = http.NewServeMux()
)

func init() {
	mux.HandleFunc("/", handlerFeed)
	mux.HandleFunc("/save-link", handlerAddLink)
	mux.HandleFunc("/post/", handlerPost)
	mux.HandleFunc("/go/", handlerGo)
	mux.HandleFunc("/about", handlerAbout)
	mux.HandleFunc("/cat/", handlerCategory)
	mux.HandleFunc("/register", handlerRegister)
	mux.HandleFunc("/login", handlerLogin)
	mux.HandleFunc("/logout", handlerLogout)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(fs))))
}

func handlerLogout(w http.ResponseWriter, rq *http.Request) {
	if rq.Method == http.MethodGet {
		templateExec(templateLogoutForm, dataAuthorized{
			Authorized: auth.AuthorizedFromRequest(rq),
		}, w)
		return
	}

	auth.LogoutFromRequest(w, rq)
	// TODO: Redirect to the previous (?) location, whatever it is
	http.Redirect(w, rq, "/", http.StatusSeeOther)
}

type dataLogin struct {
	Authorized bool
	Name       string
	Pass       string
	Incorrect  bool
}

func handlerLogin(w http.ResponseWriter, rq *http.Request) {
	if rq.Method == http.MethodGet {
		templateExec(templateLoginForm, dataLogin{
			Authorized: auth.AuthorizedFromRequest(rq),
		}, w)
		return
	}

	var (
		name = rq.FormValue("name")
		pass = rq.FormValue("pass")
	)

	if !auth.CredentialsMatch(name, pass) {
		// If incorrect password, ask the client to try again.
		w.WriteHeader(http.StatusBadRequest)
		templateExec(templateLoginForm, dataLogin{
			Authorized: auth.AuthorizedFromRequest(rq),
			Name:       name,
			Pass:       pass,
			Incorrect:  true,
		}, w)
		return
	}

	auth.LogInResponse(w)
	// TODO: Redirect to the previous (?) location, whatever it is
	http.Redirect(w, rq, "/", http.StatusSeeOther)
}

func handlerRegister(w http.ResponseWriter, rq *http.Request) {
	log.Println("/register")
	if auth.Ready() {
		// TODO: Let admin change credentials.
		log.Println("Cannot reregister")
		return
	}
	var (
		name = rq.FormValue("name")
		pass = rq.FormValue("pass")
	)
	auth.SetCredentials(name, pass)
	auth.LogInResponse(w)
	http.Redirect(w, rq, "/", http.StatusSeeOther)
}

func Start() {
	log.Fatal(http.ListenAndServe(":1738", &auther{mux}))
}

type auther struct {
	http.Handler
}

type dataAuthorized struct {
	Authorized bool
}

func (a *auther) ServeHTTP(w http.ResponseWriter, rq *http.Request) {
	if auth.Ready() ||
		strings.HasPrefix(rq.URL.Path, "/static/") ||
		strings.HasPrefix(rq.URL.Path, "/register") {
		a.Handler.ServeHTTP(w, rq)
		return
	}
	templateExec(templateRegisterForm, dataAuthorized{}, w)
}

type dataCategories struct {
	Categories []types.Category
}

func handlerCategories(w http.ResponseWriter, rq *http.Request) {
	templateExec(templateCategories, dataCategories{
		Categories: db.Categories(),
	}, w)
}

type dataCategory struct {
	types.Category
	YieldPostsInCategory chan types.Post
}

func handlerCategory(w http.ResponseWriter, rq *http.Request) {
	s := strings.TrimPrefix(rq.URL.Path, "/cat/")
	if s == "" {
		handlerCategories(w, rq)
		return
	}
	id, err := strconv.Atoi(s)
	if err != nil {
		// TODO: Show 404
		log.Println(err)
		handlerFeed(w, rq)
		return
	}
	name, generator := db.PostsForCategoryAndNameByID(id)
	templateExec(templateCategory, dataCategory{
		Category: types.Category{
			ID:   id,
			Name: name,
		},
		YieldPostsInCategory: generator,
	}, w)
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
	post.Categories = db.CategoriesForPost(id)
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
		YieldAllPosts: db.YieldAllPosts(),
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
