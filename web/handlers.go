package web

import (
	"embed"
	"fmt"
	"git.sr.ht/~bouncepaw/betula/auth"
	"git.sr.ht/~bouncepaw/betula/feeds"
	"git.sr.ht/~bouncepaw/betula/settings"
	"html/template"
	"io"
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
	mux.HandleFunc("/digest-rss", handlerDigestRss)
	mux.HandleFunc("/posts-rss", handlerPostsRss)
	mux.HandleFunc("/save-link", handlerSaveLink)
	mux.HandleFunc("/edit-link/", handlerEditLink)
	mux.HandleFunc("/delete-link/", handlerDeleteLink)
	mux.HandleFunc("/post/", handlerPost)
	mux.HandleFunc("/last/", handlerPostLast)
	mux.HandleFunc("/go/", handlerGo)
	mux.HandleFunc("/about", handlerAbout)
	mux.HandleFunc("/tag/", handlerTag)
	mux.HandleFunc("/edit-tag/", handlerEditTag)
	mux.HandleFunc("/delete-tag/", handlerDeleteTag)
	mux.HandleFunc("/day/", handlerDay)
	mux.HandleFunc("/register", handlerRegister)
	mux.HandleFunc("/login", handlerLogin)
	mux.HandleFunc("/logout", handlerLogout)
	mux.HandleFunc("/settings", handlerSettings)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(fs))))
}

func handlerPostsRss(w http.ResponseWriter, rq *http.Request) {
	feed := feeds.Posts()
	rss, err := feed.ToRss()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/rss+xml")
	_, _ = io.WriteString(w, rss)
}

func handlerDigestRss(w http.ResponseWriter, rq *http.Request) {
	feed := feeds.Digest()
	rss, err := feed.ToRss()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, err.Error()) // Ain't that failing on my watch.
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/rss+xml")
	_, _ = io.WriteString(w, rss)
}

var dayStampRegex = regexp.MustCompile("^[0-9]{4}-[0-9]{2}-[0-9]{2}$")

type dataDay struct {
	*dataCommon
	DayStamp string
	Posts    []types.Post
}

func handlerDay(w http.ResponseWriter, rq *http.Request) {
	authed := auth.AuthorizedFromRequest(rq)
	dayStamp := strings.TrimPrefix(rq.URL.Path, "/day/")
	// If no day given, default to today.
	if dayStamp == "" {
		now := time.Now()
		dayStamp = fmt.Sprintf("%d-%02d-%02d", now.Year(), now.Month(), now.Day())
	} else if !dayStampRegex.MatchString(dayStamp) {
		handlerNotFound(w, rq)
		return
	}
	templateExec(w, templateDay, dataDay{
		dataCommon: emptyCommon(),
		DayStamp:   dayStamp,
		Posts:      db.PostsForDay(authed, dayStamp),
	}, rq)
}

type dataSettings struct {
	types.Settings
	*dataCommon
	ErrBadPort bool
}

func handlerSettings(w http.ResponseWriter, rq *http.Request) {
	authed := auth.AuthorizedFromRequest(rq)
	if !authed {
		log.Printf("Unauthorized attempt to access %s. %d.\n", rq.URL.Path, http.StatusUnauthorized)
		handlerUnauthorized(w, rq)
		return
	}

	if rq.Method == http.MethodGet {
		templateExec(w, templateSettings, dataSettings{
			Settings: types.Settings{
				NetworkPort:               settings.NetworkPort(),
				SiteName:                  settings.SiteName(),
				SiteURL:                   settings.SiteURL(),
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
		SiteURL:                   rq.FormValue("site-url"),
	}

	if port, err := strconv.Atoi(rq.FormValue("network-port")); err != nil || port <= 0 {
		templateExec(w, templateSettings, dataSettings{
			Settings: types.Settings{
				NetworkPort:               uint(port),
				SiteName:                  rq.FormValue("site-name"),
				SiteTitle:                 template.HTML(rq.FormValue("site-title")),
				SiteDescriptionMycomarkup: rq.FormValue("site-description"),
				SiteURL:                   rq.FormValue("site-url"),
			},
			ErrBadPort: true,
			dataCommon: emptyCommon(),
		}, rq)
	} else {
		newSettings.NetworkPort = settings.Uintport(port).ValidatePort()
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
		handlerNotFound(w, rq)
		return
	}

	authed := auth.AuthorizedFromRequest(rq)
	if !authed {
		log.Printf("Unauthorized attempt to access %s. %d.\n", rq.URL.Path, http.StatusUnauthorized)
		handlerUnauthorized(w, rq)
		return
	}

	s := strings.TrimPrefix(rq.URL.Path, "/delete-link/")
	if s == "" {
		handlerNotFound(w, rq)
		return
	}

	id, err := strconv.Atoi(s)
	if err != nil {
		log.Println(err)
		handlerNotFound(w, rq)
		return
	}

	if confirmed := rq.FormValue("confirmed"); confirmed != "true" {
		http.Redirect(w, rq, fmt.Sprintf("/edit-link/%d", id), http.StatusSeeOther)
		return
	}

	if !db.HasPost(id) {
		log.Println("Trying to delete a non-existent post.")
		handlerNotFound(w, rq)
		return
	}

	db.DeletePost(id)
	http.Redirect(w, rq, "/", http.StatusSeeOther)
}

func handlerNotFound(w http.ResponseWriter, rq *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	templateExec(w, templateStatus, dataAuthorized{
		dataCommon: emptyCommon(),
		Status:     http.StatusText(http.StatusNotFound),
	}, rq)
}

func handlerUnauthorized(w http.ResponseWriter, rq *http.Request) {
	w.WriteHeader(http.StatusUnauthorized)
	templateExec(w, templateStatus, dataAuthorized{
		dataCommon: emptyCommon(),
		Status:     http.StatusText(http.StatusUnauthorized),
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

type dataTags struct {
	*dataCommon
	Tags []types.Tag
}

func handlerTags(w http.ResponseWriter, rq *http.Request) {
	authed := auth.AuthorizedFromRequest(rq)
	templateExec(w, templateTags, dataTags{
		Tags:       db.Tags(authed),
		dataCommon: emptyCommon(),
	}, rq)
}

type dataTag struct {
	*dataCommon
	types.Tag
	PostsInTag []types.Post
}

func handlerTag(w http.ResponseWriter, rq *http.Request) {
	catName := strings.TrimPrefix(rq.URL.Path, "/tag/")
	if catName == "" {
		handlerTags(w, rq)
		return
	}
	authed := auth.AuthorizedFromRequest(rq)
	templateExec(w, templateTag, dataTag{
		Tag: types.Tag{
			Name:        catName,
			Description: db.DescriptionForTag(catName),
		},
		PostsInTag: db.PostsWithTag(authed, catName),
		dataCommon: emptyCommon(),
	}, rq)
}

type dataAbout struct {
	*dataCommon
	LinkCount       int
	TagCount        uint
	OldestTime      *time.Time
	NewestTime      *time.Time
	SiteDescription template.HTML
}

func handlerAbout(w http.ResponseWriter, rq *http.Request) {
	authed := auth.AuthorizedFromRequest(rq)
	templateExec(w, templateAbout, dataAbout{
		dataCommon:      emptyCommon(),
		LinkCount:       db.PostCount(authed),
		TagCount:        db.TagCount(authed),
		OldestTime:      db.OldestTime(authed),
		NewestTime:      db.NewestTime(authed),
		SiteDescription: settings.SiteDescriptionHTML(),
	}, rq)
}

type dataEditLink struct {
	*dataCommon
	ErrorInvalidURL bool
	types.Post
}

func MixUpTitleLink(title *string, addr *string) {
	// If addr is a valid url we do not mix up
	_, err := url.ParseRequestURI(*addr)
	if err == nil {
		return
	}

	_, err = url.ParseRequestURI(*title)
	if err == nil {
		*addr, *title = *title, *addr
	}
}

func handlerEditLink(w http.ResponseWriter, rq *http.Request) {
	authed := auth.AuthorizedFromRequest(rq)
	if !authed {
		log.Printf("Unauthorized attempt to access %s. %d.\n", rq.URL.Path, http.StatusUnauthorized)
		handlerUnauthorized(w, rq)
		return
	}

	common := emptyCommon()
	common.head = `<script defer src="/static/autocompletion.js"></script>`

	s := strings.TrimPrefix(rq.URL.Path, "/edit-link/")
	if s == "" {
		http.Redirect(w, rq, "/save-link", http.StatusSeeOther)
		return
	}

	id, err := strconv.Atoi(s)
	if err != nil {
		log.Println(err)
		handlerNotFound(w, rq)
		return
	}

	post, found := db.PostForID(id)
	if !found {
		log.Printf("Trying to edit post no. %d that does not exist. %d.\n", id, http.StatusNotFound)
		handlerNotFound(w, rq)
		return
	}
	post.Tags = db.TagsForPost(id)

	switch rq.Method {
	case http.MethodGet:
		templateExec(w, templateEditLink, dataEditLink{
			Post:       post,
			dataCommon: common,
		}, rq)
	case http.MethodPost:
		post.URL = rq.FormValue("url")
		post.Title = rq.FormValue("title")
		post.Visibility = types.VisibilityFromString(rq.FormValue("visibility"))
		post.Description = rq.FormValue("description")
		post.Tags = types.SplitTags(rq.FormValue("tags"))

		MixUpTitleLink(&post.Title, &post.URL)

		if _, err := url.ParseRequestURI(post.URL); err != nil {
			log.Printf("Invalid URL was passed, asking again: %s\n", post.URL)
			templateExec(w, templateEditLink, dataEditLink{
				ErrorInvalidURL: true,
				Post:            post,
				dataCommon:      common,
			}, rq)
			return
		}

		db.EditPost(post)
		http.Redirect(w, rq, fmt.Sprintf("/%d", id), http.StatusSeeOther)
		log.Printf("Edited post no. %d\n", id)
	}
}

type dataEditTag struct {
	*dataCommon
	types.Tag
	ErrorTakenName   bool
	ErrorNonExistent bool
}

func handlerEditTag(w http.ResponseWriter, rq *http.Request) {
	authed := auth.AuthorizedFromRequest(rq)
	if !authed {
		log.Printf("Unauthorized attempt to access %s. %d.\n", rq.URL.Path, http.StatusUnauthorized)
		handlerUnauthorized(w, rq)
		return
	}

	var oldTag types.Tag
	oldName := strings.TrimPrefix(rq.URL.Path, "/edit-tag/")
	oldTag.Name = oldName
	oldTag.Description = db.DescriptionForTag(oldName)

	switch rq.Method {
	case http.MethodGet:
		templateExec(w, templateEditTag, dataEditTag{
			Tag:        oldTag,
			dataCommon: emptyCommon(),
		}, rq)
	case http.MethodPost:
		var newTag types.Tag
		newName := types.CanonicalTagName(rq.FormValue("new-name"))
		newTag.Name = newName
		newTag.Description = strings.TrimSpace(rq.FormValue("description"))

		merge := rq.FormValue("merge")

		if db.TagExists(newTag.Name) && merge != "true" && newTag.Name != oldTag.Name {
			log.Printf("Trying to rename a tag %s to a taken name %s.\n", oldTag.Name, newTag.Name)
			templateExec(w, templateEditTag, dataEditTag{
				Tag:            oldTag,
				ErrorTakenName: true,
				dataCommon:     emptyCommon(),
			}, rq)
			return
		} else if !db.TagExists(oldTag.Name) {
			log.Printf("Trying to rename a non-existent tag %s.\n", oldTag.Name)
			templateExec(w, templateEditTag, dataEditTag{
				Tag:              oldTag,
				ErrorNonExistent: true,
				dataCommon:       emptyCommon(),
			}, rq)
			return
		} else {
			db.RenameTag(oldTag.Name, newTag.Name)
			db.SetTagDescription(oldTag.Name, "")
			db.SetTagDescription(newTag.Name, newTag.Description)
			http.Redirect(w, rq, fmt.Sprintf("/tag/%s", newTag.Name), http.StatusSeeOther)
			if oldTag.Name != newTag.Name {
				log.Printf("Renamed tag %s to %s\n", oldTag.Name, newTag.Name)
			}
			if oldTag.Description != newTag.Description {
				log.Printf("Set new description for tag %s\n", newTag.Name)
			}
		}
	}
}

func handlerDeleteTag(w http.ResponseWriter, rq *http.Request) {
	if rq.Method != http.MethodPost {
		handlerNotFound(w, rq)
		return
	}

	authed := auth.AuthorizedFromRequest(rq)
	if !authed {
		log.Printf("Unauthorized attempt to access %s. %d.\n", rq.URL.Path, http.StatusUnauthorized)
		handlerUnauthorized(w, rq)
		return
	}

	catName := strings.TrimPrefix(rq.URL.Path, "/delete-tag/")
	if catName == "" {
		handlerNotFound(w, rq)
		return
	}

	if confirmed := rq.FormValue("confirmed"); confirmed != "true" {
		http.Redirect(w, rq, fmt.Sprintf("/edit-tag/%s", catName), http.StatusSeeOther)
		return
	}

	if !db.TagExists(catName) {
		log.Println("Trying to delete a non-existent tag.")
		handlerNotFound(w, rq)
		return
	}
	db.DeleteTag(catName)
	http.Redirect(w, rq, "/", http.StatusSeeOther)
}

type dataSaveLink struct {
	*dataCommon

	// The following three fields can be non-empty, when set through URL parameters or when an erroneous request was made.

	URL             string
	Title           string
	Visibility      types.Visibility
	Description     string
	Tags            []types.Tag
	Another         bool
	ErrorInvalidURL bool
	ErrorNotFilled  bool
}

func handlerSaveLink(w http.ResponseWriter, rq *http.Request) {
	if !auth.AuthorizedFromRequest(rq) {
		log.Printf("Unauthorized attempt to access %s. %d.\n", rq.URL.Path, http.StatusUnauthorized)
		handlerUnauthorized(w, rq)
		return
	}
	common := emptyCommon()
	common.head = `<script defer src="/static/autocompletion.js"></script>`
	switch rq.Method {
	case http.MethodGet:
		// TODO: Document the param behaviour
		templateExec(w, templateSaveLink, dataSaveLink{
			URL:         rq.FormValue("url"),
			Title:       rq.FormValue("title"),
			Visibility:  types.VisibilityFromString(rq.FormValue("visibility")),
			Description: rq.FormValue("description"),
			Tags:        types.SplitTags(rq.FormValue("tags")),
			dataCommon:  common,
		}, rq)
	case http.MethodPost:
		var (
			addr        = rq.FormValue("url")
			title       = rq.FormValue("title")
			visibility  = types.VisibilityFromString(rq.FormValue("visibility"))
			description = rq.FormValue("description")
			tags        = types.SplitTags(rq.FormValue("tags"))
		)

		if addr == "" || title == "" {
			templateExec(w, templateSaveLink, dataSaveLink{
				URL:            addr,
				Title:          title,
				Visibility:     visibility,
				Description:    description,
				Tags:           tags,
				dataCommon:     common,
				ErrorNotFilled: true,
			}, rq)
			return
		}

		MixUpTitleLink(&title, &addr)

		if _, err := url.ParseRequestURI(addr); err != nil {
			templateExec(w, templateSaveLink, dataSaveLink{
				URL:             addr,
				Title:           title,
				Visibility:      visibility,
				Description:     description,
				Tags:            tags,
				dataCommon:      common,
				ErrorInvalidURL: true,
			}, rq)
			return
		}

		id := db.AddPost(types.Post{
			URL:         addr,
			Title:       title,
			Description: description,
			Visibility:  visibility,
			Tags:        tags,
		})

		another := rq.FormValue("another")
		if another == "true" {
			templateExec(w, templateSaveLink, dataSaveLink{
				dataCommon: common,
				Visibility: types.Public,
				Another:    true,
			}, rq)
			return
		}

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
		handlerNotFound(w, rq)
		return
	}

	post, found := db.PostForID(id)
	if !found {
		log.Println(err)
		handlerNotFound(w, rq)
		return
	}

	visibility := post.Visibility
	authed := auth.AuthorizedFromRequest(rq)
	if visibility == types.Private && !authed {
		log.Printf("Unauthorized attempt to access %s. %d.\n", rq.URL.Path, http.StatusUnauthorized)
		handlerUnauthorized(w, rq)
		return
	}

	log.Printf("Viewing post %d\n", id)

	post.Tags = db.TagsForPost(id)
	templateExec(w, templatePost, dataPost{
		Post:       post,
		dataCommon: emptyCommon(),
	}, rq)
}

func handlerPostLast(w http.ResponseWriter, rq *http.Request) {
	authed := auth.AuthorizedFromRequest(rq)
	post, found := db.LastPost(authed)
	if !found {
		log.Println("Can't reach the latest post")
		handlerNotFound(w, rq)
		return
	}
	log.Printf("Viewing the latest post %d\n", post.ID)
	post.Tags = db.TagsForPost(post.ID)
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
		handlerNotFound(w, rq)
		return
	}
	authed := auth.AuthorizedFromRequest(rq)
	common := emptyCommon()
	common.head = `
	<link rel="alternate" type="application/rss+xml" title="Daily digest" href="/digest-rss" />
`
	templateExec(w, templateFeed, dataFeed{
		AllPosts:        db.Posts(authed),
		SiteDescription: settings.SiteDescriptionHTML(),
		dataCommon:      common,
	}, rq)
}

func handlerGo(w http.ResponseWriter, rq *http.Request) {
	id, err := strconv.Atoi(strings.TrimPrefix(rq.URL.Path, "/go/"))
	if err != nil {
		log.Printf("%d: %s\n", http.StatusNotFound, rq.URL.Path)
		handlerNotFound(w, rq)
		return
	}

	var (
		authed      = auth.AuthorizedFromRequest(rq)
		post, found = db.PostForID(id)
	)
	switch {
	case !found:
		log.Printf("%d: %s\n", http.StatusNotFound, rq.URL.Path)
		handlerNotFound(w, rq)
	case !authed && post.Visibility == types.Private:
		log.Printf("Unauthorized attempt to access %s. %d.\n", rq.URL.Path, http.StatusUnauthorized)
		handlerUnauthorized(w, rq)
	default:
		http.Redirect(w, rq, post.URL, http.StatusSeeOther)
	}
}
