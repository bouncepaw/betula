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
	mux.HandleFunc("/save-link", handlerSaveLink)
	mux.HandleFunc("/edit-link/", handlerEditLink)
	mux.HandleFunc("/delete-link/", handlerDeleteLink)
	mux.HandleFunc("/post/", handlerPost)
	mux.HandleFunc("/go/", handlerGo)
	mux.HandleFunc("/about", handlerAbout)
	mux.HandleFunc("/cat/", handlerCategory)
	mux.HandleFunc("/register", handlerRegister)
	mux.HandleFunc("/login", handlerLogin)
	mux.HandleFunc("/logout", handlerLogout)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(fs))))
}

func handlerDeleteLink(w http.ResponseWriter, rq *http.Request) {
	if rq.Method != http.MethodPost {
		handler404(w, rq)
		return
	}

	authed := auth.AuthorizedFromRequest(rq)
	if !authed {
		log.Printf("Unauthorized attempt to access %s. 404.\n", rq.URL.Path)
		handler404(w, rq)
		return
	}

	s := strings.TrimPrefix(rq.URL.Path, "/delete-link/")
	if s == "" {
		handler404(w, rq)
		return
	}

	id, err := strconv.Atoi(s)
	if err != nil {
		log.Println(err)
		handler404(w, rq)
		return
	}

	if confirmed := rq.FormValue("confirmed"); confirmed != "true" {
		http.Redirect(w, rq, fmt.Sprintf("/edit-link/%d", id), http.StatusSeeOther)
		return
	}

	if !db.HasPost(id) {
		log.Println("Trying to delete a non-existent post.")
		handler404(w, rq)
		return
	}

	db.DeletePost(id)
	http.Redirect(w, rq, "/", http.StatusSeeOther)
}

func handler404(w http.ResponseWriter, rq *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	templateExec(template404, dataAuthorized{
		Authorized: auth.AuthorizedFromRequest(rq),
	}, w)
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
	Authorized bool
}

func handlerCategories(w http.ResponseWriter, rq *http.Request) {
	templateExec(templateCategories, dataCategories{
		Categories: db.Categories(),
		Authorized: auth.AuthorizedFromRequest(rq),
	}, w)
}

type dataCategory struct {
	types.Category
	PostsInCategory []types.Post
	Authorized      bool
}

func handlerCategory(w http.ResponseWriter, rq *http.Request) {
	catName := strings.TrimPrefix(rq.URL.Path, "/cat/")
	if catName == "" {
		handlerCategories(w, rq)
		return
	}
	authed := auth.AuthorizedFromRequest(rq)
	templateExec(templateCategory, dataCategory{
		Category:        types.Category{Name: catName},
		PostsInCategory: db.AuthorizedPostsForCategory(authed, catName),
		Authorized:      authed,
	}, w)
}

type dataAbout struct {
	LinkCount  int
	OldestTime *time.Time
	NewestTime *time.Time
	Authorized bool
}

func handlerAbout(w http.ResponseWriter, rq *http.Request) {
	templateExec(templateAbout, dataAbout{
		LinkCount:  db.LinkCount(),
		OldestTime: db.OldestTime(),
		NewestTime: db.NewestTime(),
		Authorized: auth.AuthorizedFromRequest(rq),
	}, w)
}

type dataEditLink struct {
	Authorized      bool
	ErrorInvalidURL bool
	types.Post
}

func handlerEditLink(w http.ResponseWriter, rq *http.Request) {
	authed := auth.AuthorizedFromRequest(rq)
	if !authed {
		log.Printf("Unauthorized attempt to access %s. 404.\n", rq.URL.Path)
		handler404(w, rq)
		return
	}

	s := strings.TrimPrefix(rq.URL.Path, "/edit-link/")
	if s == "" {
		http.Redirect(w, rq, "/save-link", http.StatusSeeOther)
		return
	}

	id, err := strconv.Atoi(s)
	if err != nil {
		log.Println(err)
		handler404(w, rq)
		return
	}

	post, found := db.PostForID(id)
	if !found {
		log.Printf("Trying to edit post no. %d that does not exist. 404.\n", id)
		handler404(w, rq)
		return
	}
	post.Categories = db.CategoriesForPost(id)

	switch rq.Method {
	case http.MethodGet:
		templateExec(templateEditLink, dataEditLink{
			Authorized: authed,
			Post:       post,
		}, w)
	case http.MethodPost:
		post.URL = rq.FormValue("url")
		post.Title = rq.FormValue("title")
		post.Visibility = types.VisibilityFromString(rq.FormValue("visibility"))
		post.Description = rq.FormValue("description")
		post.Categories = types.SplitCategories(rq.FormValue("categories"))

		if _, err := url.ParseRequestURI(post.URL); err != nil {
			log.Printf("Invalid URL was passed, asking again: %s\n", post.URL)
			templateExec(templateEditLink, dataEditLink{
				ErrorInvalidURL: true,
				Authorized:      authed,
				Post:            post,
			}, w)
			return
		}

		db.EditPost(post)
		http.Redirect(w, rq, fmt.Sprintf("/%d", id), http.StatusSeeOther)
		log.Printf("Edited post no. %d\n", id)
	}
}

type dataSaveLink struct {
	Authorized bool

	// The following three fields can be non-empty, when set through URL parameters or when an erroneous request was made.

	URL         string
	Title       string
	Visibility  types.Visibility
	Description string
	Categories  []types.Category
}

func handlerSaveLink(w http.ResponseWriter, rq *http.Request) {
	switch rq.Method {
	case http.MethodGet:
		// TODO: Document the param behaviour
		templateExec(templateSaveLink, dataSaveLink{
			Authorized:  auth.AuthorizedFromRequest(rq),
			URL:         rq.FormValue("url"),
			Title:       rq.FormValue("title"),
			Visibility:  types.VisibilityFromString(rq.FormValue("visibility")),
			Description: rq.FormValue("description"),
			Categories:  types.SplitCategories(rq.FormValue("categories")),
		}, w)
	case http.MethodPost:
		var (
			addr        = rq.FormValue("url")
			title       = rq.FormValue("title")
			visibility  = types.VisibilityFromString(rq.FormValue("visibility"))
			description = rq.FormValue("description")
			categories  = types.SplitCategories(rq.FormValue("categories"))
		)
		if _, err := url.ParseRequestURI(addr); err != nil {
			templateExec(templateAddLinkInvalidURL, dataSaveLink{
				URL:         addr,
				Title:       title,
				Visibility:  visibility,
				Description: description,
				Categories:  categories,
			}, w)
			return
		}

		id := db.AddPost(types.Post{
			URL:         addr,
			Title:       title,
			Description: description,
			Visibility:  visibility,
			Categories:  categories,
		})

		http.Redirect(w, rq, fmt.Sprintf("/%d", id), http.StatusSeeOther)
	}
}

type dataPost struct {
	Post       types.Post
	Authorized bool
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
		log.Println(err)
		handler404(w, rq)
		return
	}
	post.Categories = db.CategoriesForPost(id)
	templateExec(templatePost, dataPost{
		Post:       post,
		Authorized: auth.AuthorizedFromRequest(rq),
	}, w)
}

type dataFeed struct {
	AllPosts   []types.Post
	Authorized bool
}

var regexpPost = regexp.MustCompile("^/[0-9]+")

func handlerFeed(w http.ResponseWriter, rq *http.Request) {
	if regexpPost.MatchString(rq.URL.Path) {
		handlerPost(w, rq)
		return
	}
	if rq.URL.Path != "/" {
		handler404(w, rq)
		return
	}
	authed := auth.AuthorizedFromRequest(rq)
	templateExec(templateFeed, dataFeed{
		AllPosts:   db.AuthorizedPosts(authed),
		Authorized: authed,
	}, w)
}

func handlerGo(w http.ResponseWriter, rq *http.Request) {
	id, err := strconv.Atoi(strings.TrimPrefix(rq.URL.Path, "/go/"))
	if err != nil {
		handlerFeed(w, rq)
		return
	}

	if addr := db.URLForID(id); addr.Valid {
		http.Redirect(w, rq, addr.String, http.StatusSeeOther)
	} else {
		handler404(w, rq)
	}
}
