package web

import (
	"embed"
	"fmt"
	"git.sr.ht/~bouncepaw/betula/auth"
	"git.sr.ht/~bouncepaw/betula/settings"
	"html/template"
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
	//go:embed *.gohtml *.css *.js
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
	mux.HandleFunc("/settings", handlerSettings)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(fs))))
}

type dataSettings struct {
	types.Settings
	*dataCommon
	ErrBadPort bool
}

func handlerSettings(w http.ResponseWriter, rq *http.Request) {
	authed := auth.AuthorizedFromRequest(rq)
	if !authed {
		handler404(w, rq)
		return
	}

	if rq.Method == http.MethodGet {
		templateExec(w, templateSettings, dataSettings{
			Settings: types.Settings{
				NetworkPort:               settings.NetworkPort(),
				SiteName:                  settings.SiteName(),
				SiteTitle:                 settings.SiteTitle(),
				SiteDescriptionMycomarkup: settings.SiteDescriptionMycomarkup(),
			},
			dataCommon: emptyCommon(),
		}, rq)
		return
	}

	var newSettings = types.Settings{
		SiteName:                  rq.FormValue("site-name"),
		SiteTitle:                 template.HTML(rq.FormValue("site-title")),
		SiteDescriptionMycomarkup: rq.FormValue("site-description"),
	}

	if port, err := strconv.Atoi(rq.FormValue("network-port")); err != nil || port <= 0 {
		templateExec(w, templateSettings, dataSettings{
			Settings: types.Settings{
				NetworkPort: uint(port),
				SiteName:    settings.SiteName(),
				SiteTitle:   settings.SiteTitle(),
			},
			ErrBadPort: true,
			dataCommon: emptyCommon(),
		}, rq)
	} else {
		newSettings.NetworkPort = uint(port)
	}

	oldPort := settings.NetworkPort()
	settings.SetSettings(newSettings)
	if oldPort != settings.NetworkPort() {
		restartServer()
	}
	http.Redirect(w, rq, "/", http.StatusSeeOther)
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
	templateExec(w, template404, dataAuthorized{
		dataCommon: emptyCommon(),
	}, rq)
}

func handlerLogout(w http.ResponseWriter, rq *http.Request) {
	if rq.Method == http.MethodGet {
		templateExec(w, templateLogoutForm, dataAuthorized{
			dataCommon: emptyCommon(),
		}, rq)
		return
	}

	auth.LogoutFromRequest(w, rq)
	// TODO: Redirect to the previous (?) location, whatever it is
	http.Redirect(w, rq, "/", http.StatusSeeOther)
}

type dataLogin struct {
	*dataCommon
	Name      string
	Pass      string
	Incorrect bool
}

func handlerLogin(w http.ResponseWriter, rq *http.Request) {
	if rq.Method == http.MethodGet {
		templateExec(w, templateLoginForm, dataLogin{
			dataCommon: emptyCommon(),
		}, rq)
		return
	}

	var (
		name = rq.FormValue("name")
		pass = rq.FormValue("pass")
	)

	if !auth.CredentialsMatch(name, pass) {
		// If incorrect password, ask the client to try again.
		w.WriteHeader(http.StatusBadRequest)
		templateExec(w, templateLoginForm, dataLogin{
			Name:       name,
			Pass:       pass,
			Incorrect:  true,
			dataCommon: emptyCommon(),
		}, rq)
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

type dataCategories struct {
	*dataCommon
	Categories []types.Category
}

func handlerCategories(w http.ResponseWriter, rq *http.Request) {
	authed := auth.AuthorizedFromRequest(rq)
	templateExec(w, templateCategories, dataCategories{
		Categories: db.Categories(authed),
		dataCommon: emptyCommon(),
	}, rq)
}

type dataCategory struct {
	*dataCommon
	types.Category
	PostsInCategory []types.Post
}

func handlerCategory(w http.ResponseWriter, rq *http.Request) {
	catName := strings.TrimPrefix(rq.URL.Path, "/cat/")
	if catName == "" {
		handlerCategories(w, rq)
		return
	}
	authed := auth.AuthorizedFromRequest(rq)
	templateExec(w, templateCategory, dataCategory{
		Category:        types.Category{Name: catName},
		PostsInCategory: db.AuthorizedPostsForCategory(authed, catName),
		dataCommon:      emptyCommon(),
	}, rq)
}

type dataAbout struct {
	*dataCommon
	LinkCount       int
	OldestTime      *time.Time
	NewestTime      *time.Time
	SiteDescription template.HTML
}

func handlerAbout(w http.ResponseWriter, rq *http.Request) {
	templateExec(w, templateAbout, dataAbout{
		dataCommon:      emptyCommon(),
		LinkCount:       db.LinkCount(),
		OldestTime:      db.OldestTime(),
		NewestTime:      db.NewestTime(),
		SiteDescription: settings.SiteDescriptionHTML(),
	}, rq)
}

type dataEditLink struct {
	*dataCommon
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
		templateExec(w, templateEditLink, dataEditLink{
			Post:       post,
			dataCommon: emptyCommon(),
		}, rq)
	case http.MethodPost:
		post.URL = rq.FormValue("url")
		post.Title = rq.FormValue("title")
		post.Visibility = types.VisibilityFromString(rq.FormValue("visibility"))
		post.Description = rq.FormValue("description")
		post.Categories = types.SplitCategories(rq.FormValue("categories"))

		if _, err := url.ParseRequestURI(post.URL); err != nil {
			log.Printf("Invalid URL was passed, asking again: %s\n", post.URL)
			templateExec(w, templateEditLink, dataEditLink{
				ErrorInvalidURL: true,
				Post:            post,
				dataCommon:      emptyCommon(),
			}, rq)
			return
		}

		db.EditPost(post)
		http.Redirect(w, rq, fmt.Sprintf("/%d", id), http.StatusSeeOther)
		log.Printf("Edited post no. %d\n", id)
	}
}

type dataSaveLink struct {
	*dataCommon

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
		templateExec(w, templateSaveLink, dataSaveLink{
			URL:         rq.FormValue("url"),
			Title:       rq.FormValue("title"),
			Visibility:  types.VisibilityFromString(rq.FormValue("visibility")),
			Description: rq.FormValue("description"),
			Categories:  types.SplitCategories(rq.FormValue("categories")),
			dataCommon:  emptyCommon(),
		}, rq)
	case http.MethodPost:
		var (
			addr        = rq.FormValue("url")
			title       = rq.FormValue("title")
			visibility  = types.VisibilityFromString(rq.FormValue("visibility"))
			description = rq.FormValue("description")
			categories  = types.SplitCategories(rq.FormValue("categories"))
		)
		if _, err := url.ParseRequestURI(addr); err != nil {
			templateExec(w, templateAddLinkInvalidURL, dataSaveLink{
				URL:         addr,
				Title:       title,
				Visibility:  visibility,
				Description: description,
				Categories:  categories,
				dataCommon:  emptyCommon(),
			}, rq)
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
	Post types.Post
	*dataCommon
}

func handlerPost(w http.ResponseWriter, rq *http.Request) {
	id, err := strconv.Atoi(strings.TrimPrefix(strings.TrimPrefix(rq.URL.Path, "/"), "post/"))
	if err != nil {
		log.Println(err)
		handler404(w, rq)
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
	templateExec(w, templatePost, dataPost{
		Post:       post,
		dataCommon: emptyCommon(),
	}, rq)
}

type dataFeed struct {
	AllPosts        []types.Post
	SiteDescription template.HTML
	*dataCommon
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
	templateExec(w, templateFeed, dataFeed{
		AllPosts:        db.AuthorizedPosts(authed),
		SiteDescription: settings.SiteDescriptionHTML(),
		dataCommon:      emptyCommon(),
	}, rq)
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
