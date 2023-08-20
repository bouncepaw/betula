package web

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"git.sr.ht/~bouncepaw/betula/jobs"
	"git.sr.ht/~bouncepaw/betula/readpage"
	"git.sr.ht/~bouncepaw/betula/stricks"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"git.sr.ht/~bouncepaw/betula/auth"
	"git.sr.ht/~bouncepaw/betula/db"
	"git.sr.ht/~bouncepaw/betula/feeds"
	"git.sr.ht/~bouncepaw/betula/help"
	"git.sr.ht/~bouncepaw/betula/search"
	"git.sr.ht/~bouncepaw/betula/settings"
	"git.sr.ht/~bouncepaw/betula/types"
)

var (
	//go:embed views/*.gohtml *.css *.js
	fs embed.FS
	//go:embed bookmarklet.js
	bookmarkletScript string
	mux               = http.NewServeMux()
)

// Wrap handlers that only make sense for the admin with this thingy in init().
func adminOnly(next func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, rq *http.Request) {
		authed := auth.AuthorizedFromRequest(rq)
		if !authed {
			log.Printf("Unauthorized attempt to access %s. %d.\n", rq.URL.Path, http.StatusUnauthorized)
			handlerUnauthorized(w, rq)
			return
		}
		next(w, rq)
	}
}

func init() {
	mux.HandleFunc("/", handlerFeed)
	mux.HandleFunc("/repost", adminOnly(handlerRepost))
	mux.HandleFunc("/.well-known/mycoverse/inbox", handlerInbox)
	mux.HandleFunc("/help/en/", handlerEnglishHelp)
	mux.HandleFunc("/help", handlerHelp)
	mux.HandleFunc("/text/", handlerText)
	mux.HandleFunc("/digest-rss", handlerDigestRss)
	mux.HandleFunc("/posts-rss", handlerPostsRss)
	mux.HandleFunc("/save-link", adminOnly(handlerSaveLink))
	mux.HandleFunc("/edit-link/", adminOnly(handlerEditLink))
	mux.HandleFunc("/delete-link/", adminOnly(handlerDeleteLink))
	mux.HandleFunc("/post/", handlerPost)
	mux.HandleFunc("/last/", handlerPostLast)
	mux.HandleFunc("/go/", handlerGo)
	mux.HandleFunc("/about", handlerAbout)
	mux.HandleFunc("/tag/", handlerTag)
	mux.HandleFunc("/edit-tag/", adminOnly(handlerEditTag))
	mux.HandleFunc("/delete-tag/", adminOnly(handlerDeleteTag))
	mux.HandleFunc("/day/", handlerDay)
	mux.HandleFunc("/search", handlerSearch)
	mux.HandleFunc("/register", handlerRegister)
	mux.HandleFunc("/login", handlerLogin)
	mux.HandleFunc("/logout", handlerLogout)
	mux.HandleFunc("/settings", adminOnly(handlerSettings))
	mux.HandleFunc("/bookmarklet", adminOnly(handlerBookmarklet))
	mux.HandleFunc("/static/style.css", handlerStyle)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(fs))))
}

type dataRepost struct {
	*dataCommon

	ErrorInvalidURL,
	ErrorEmptyURL,
	ErrorImpossible,
	ErrorTimeout bool
	Err error

	FoundData  readpage.FoundData
	URL        string
	Visibility types.Visibility
	CopyTags   bool
}

func handlerRepost(w http.ResponseWriter, rq *http.Request) {
	repost := dataRepost{
		dataCommon: emptyCommon(),
		URL:        rq.FormValue("url"),
		Visibility: types.VisibilityFromString(rq.FormValue("visibility")),
		CopyTags:   rq.FormValue("copy-tags") == "true",
	}

	var foundData readpage.FoundData
	var err error

	// I call this a railway operator.
	switch ok := true; {
	case rq.Method == http.MethodGet:
		templateExec(w, templateRepost, repost, rq)
		return

	case repost.URL == "":
		repost.ErrorEmptyURL = true
		ok = false // Mark errors like that.
		fallthrough

	case ok && !stricks.ValidURL(repost.URL):
		repost.ErrorInvalidURL = true
		ok = false
		fallthrough

	case ok:
		foundData, err = readpage.FindRepostData(repost.URL)
		fallthrough

	case ok && errors.Is(err, readpage.ErrTimeout):
		repost.ErrorTimeout = true
		ok = false
		fallthrough

	case ok && err != nil:
		repost.Err = err
		ok = false
		fallthrough

	case ok && (foundData.IsHFeed || foundData.BookmarkOf == nil || foundData.PostName == ""):
		repost.ErrorImpossible = true
		ok = false
		fallthrough

	case !ok: // All errors end up here.
		w.WriteHeader(http.StatusBadRequest)
		templateExec(w, templateRepost, repost, rq)
		return
	}

	post := types.Post{
		URL:         foundData.BookmarkOf.String(),
		Title:       foundData.PostName,
		Description: foundData.Mycomarkup,
		Visibility:  repost.Visibility,
		RepostOf:    &repost.URL,
	}

	if repost.CopyTags {
		post.Tags = types.TagsFromStringSlice(foundData.Tags)
	}

	id := db.AddPost(post)

	go jobs.NotifyAboutMyRepost(int(id))
	http.Redirect(w, rq, fmt.Sprintf("/%d", id), http.StatusSeeOther)
}

func handlerInbox(w http.ResponseWriter, rq *http.Request) {
	if rq.Method != http.MethodPost {
		handlerNotFound(w, rq)
		return
	}

	data, err := io.ReadAll(rq.Body)
	if err != nil {
		log.Fatalln(err)
	}

	var dict = map[string]any{}
	err = json.Unmarshal(data, &dict)
	if err != nil {
		log.Printf("Invalid activity came in. JSON unmarshalling failed with: %w\n", err)
		// TODO: Handle somehow somewhat show an error dunno
		return
	}

	log.Printf("An acitivity with type ‘%s’ came in.\n", dict["type"])
	log.Printf("%q\n", dict)
}

type dataBookmarklet struct {
	*dataCommon
	Script string
}

func handlerBookmarklet(w http.ResponseWriter, rq *http.Request) {
	templateExec(w, templateBookmarklet, dataBookmarklet{
		dataCommon: emptyCommon(),
		Script:     fmt.Sprintf(bookmarkletScript, settings.SiteURL()),
	}, rq)
}

func handlerHelp(w http.ResponseWriter, rq *http.Request) {
	http.Redirect(w, rq, "/help/en/index", http.StatusSeeOther)
}

type dataHelp struct {
	*dataCommon
	This   help.Topic
	Topics []help.Topic
}

func handlerEnglishHelp(w http.ResponseWriter, rq *http.Request) {
	topicName := strings.TrimPrefix(rq.URL.Path, "/help/en/")
	if topicName == "/help/en" || topicName == "/" {
		topicName = "index"
	}
	topic, found := help.GetEnglishHelp(topicName)
	if !found {
		handlerNotFound(w, rq)
		return
	}

	templateExec(w, templateHelp, dataHelp{
		dataCommon: emptyCommon(),
		This:       topic,
		Topics:     help.Topics,
	}, rq)
}

func handlerStyle(w http.ResponseWriter, rq *http.Request) {
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	file, err := fs.Open("style.css")
	if err != nil {
		// We sure have problems if we can't read something from the embedded fs.
		log.Fatalln(fmt.Errorf("reading the built-in style: %w", err))
	}

	_, err = io.Copy(w, file)
	if err != nil {
		log.Fatalln(fmt.Errorf("writing to response: %w", err))
	}

	_, err = io.WriteString(w, settings.CustomCSS())
	if err != nil {
		log.Fatalln(fmt.Errorf("writing custom CSS: %w", err))
	}

	// Look at how detailed my error messages are! In a function that will
	// basically never fail!
}

type dataSearch struct {
	*dataCommon
	Query       string
	TotalPosts  uint
	PostsInPage []types.Post
}

var tagOnly = regexp.MustCompile(`^#([^?!:#@<>*|'"&%{}\\\s]+)\s*$`)

func handlerSearch(w http.ResponseWriter, rq *http.Request) {
	query := rq.FormValue("q")
	if query == "" {
		http.Redirect(w, rq, "/", http.StatusSeeOther)
		return
	}
	if tagOnly.MatchString(query) {
		tag := tagOnly.FindAllStringSubmatch(query, 1)[0][1]
		http.Redirect(w, rq, "/tag/"+tag, http.StatusSeeOther)
		return
	}
	authed := auth.AuthorizedFromRequest(rq)
	currentPage, err := strconv.Atoi(rq.FormValue("page"))
	if err != nil || currentPage <= 0 {
		currentPage = 1
	}
	posts, totalPosts := search.For(query, authed, uint(currentPage))

	common := emptyCommon()
	common.paginator = types.PaginatorFromURL(rq.URL, uint(currentPage), totalPosts)
	common.searchQuery = query
	log.Printf("Searching ‘%s’. Authorized: %v\n", query, authed)
	templateExec(w, templateSearch, dataSearch{
		dataCommon:  common,
		Query:       query,
		PostsInPage: posts,
		TotalPosts:  totalPosts,
	}, rq)
}

func handlerText(w http.ResponseWriter, rq *http.Request) {
	id, err := strconv.Atoi(strings.TrimPrefix(rq.URL.Path, "/text/"))
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

	log.Printf("Fetching text for post no. %d\n", id)

	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, post.Description)
}

func handlerPostsRss(w http.ResponseWriter, _ *http.Request) {
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

func handlerDigestRss(w http.ResponseWriter, _ *http.Request) {
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
	if rq.Method == http.MethodGet {
		templateExec(w, templateSettings, dataSettings{
			Settings: types.Settings{
				NetworkHost:               settings.NetworkHost(),
				NetworkPort:               settings.NetworkPort(),
				SiteName:                  settings.SiteName(),
				SiteURL:                   settings.SiteURL(),
				SiteTitle:                 settings.SiteTitle(),
				SiteDescriptionMycomarkup: settings.SiteDescriptionMycomarkup(),
				CustomCSS:                 settings.CustomCSS(),
			},
			dataCommon: emptyCommon(),
		}, rq)
		return
	}

	var newSettings = types.Settings{
		NetworkHost:               rq.FormValue("network-host"),
		SiteName:                  rq.FormValue("site-name"),
		SiteTitle:                 template.HTML(rq.FormValue("site-title")),
		SiteDescriptionMycomarkup: rq.FormValue("site-description"),
		SiteURL:                   rq.FormValue("site-url"),
		CustomCSS:                 rq.FormValue("custom-css"),
	}

	// If the port ≤ 0 or not really numeric, show error.
	if port, err := strconv.Atoi(rq.FormValue("network-port")); err != nil || port <= 0 {
		newSettings.NetworkPort = uint(port)
		templateExec(w, templateSettings, dataSettings{
			Settings:   newSettings,
			ErrBadPort: true,
			dataCommon: emptyCommon(),
		}, rq)
		return
	} else {
		newSettings.NetworkPort = settings.ValidatePortFromWeb(port)
	}

	oldPort := settings.NetworkPort()
	oldHost := settings.NetworkHost()
	settings.SetSettings(newSettings)
	if oldPort != settings.NetworkPort() || oldHost != settings.NetworkHost() {
		restartServer()
	}
	http.Redirect(w, rq, "/", http.StatusSeeOther)
}

func handlerDeleteLink(w http.ResponseWriter, rq *http.Request) {
	if rq.Method == http.MethodGet {
		handlerNotFound(w, rq)
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
	TotalPosts  uint
	PostsInPage []types.Post
}

func handlerTag(w http.ResponseWriter, rq *http.Request) {
	tagName := strings.TrimPrefix(rq.URL.Path, "/tag/")
	if tagName == "" {
		handlerTags(w, rq)
		return
	}
	currentPage, err := strconv.Atoi(rq.FormValue("page"))
	if err != nil || currentPage <= 0 {
		currentPage = 1
	}
	authed := auth.AuthorizedFromRequest(rq)

	posts, totalPosts := db.PostsWithTag(authed, tagName, uint(currentPage))

	common := emptyCommon()
	common.searchQuery = "#" + tagName
	common.paginator = types.PaginatorFromURL(rq.URL, uint(currentPage), totalPosts)
	templateExec(w, templateTag, dataTag{
		Tag: types.Tag{
			Name:        tagName,
			Description: db.DescriptionForTag(tagName),
		},
		PostsInPage: posts,
		TotalPosts:  totalPosts,
		dataCommon:  common,
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

func mixUpTitleLink(title *string, addr *string) {
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

type dataEditLink struct {
	errorTemplate
	*dataCommon
	types.Post
	ErrorEmptyURL      bool
	ErrorInvalidURL    bool
	ErrorTitleNotFound bool
}

func handlerEditLink(w http.ResponseWriter, rq *http.Request) {

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

	if rq.Method == http.MethodGet {
		post.Tags = db.TagsForPost(id)
		templateExec(w, templateEditLink, dataEditLink{
			Post:       post,
			dataCommon: common,
		}, rq)
		return
	}

	post.URL = rq.FormValue("url")
	post.Title = rq.FormValue("title")
	post.Visibility = types.VisibilityFromString(rq.FormValue("visibility"))
	post.Description = rq.FormValue("description")
	post.Tags = types.SplitTags(rq.FormValue("tags"))

	var viewData dataEditLink

	if post.URL == "" && post.Title == "" {
		viewData.emptyUrl(post, common, w, rq)
		return
	}

	mixUpTitleLink(&post.Title, &post.URL)

	if post.URL == "" {
		viewData.emptyUrl(post, common, w, rq)
		return
	}

	if post.Title == "" {
		if _, err := url.ParseRequestURI(post.URL); err != nil {
			viewData.invalidUrl(post, common, w, rq)
			return
		}
		newTitle, err := readpage.FindTitle(post.URL)
		if err != nil {
			log.Printf("Can't get HTML title from URL: %s\n", post.URL)
			viewData.titleNotFound(post, common, w, rq)
			return
		}
		post.Title = newTitle
	}

	if _, err := url.ParseRequestURI(post.URL); err != nil {
		log.Printf("Invalid URL was passed, asking again: %s\n", post.URL)
		viewData.invalidUrl(post, common, w, rq)
		return
	}

	db.EditPost(post)
	http.Redirect(w, rq, fmt.Sprintf("/%d", id), http.StatusSeeOther)
	log.Printf("Edited post no. %d\n", id)
}

type dataEditTag struct {
	*dataCommon
	types.Tag
	ErrorTakenName   bool
	ErrorNonExistent bool
}

func handlerEditTag(w http.ResponseWriter, rq *http.Request) {
	oldName := strings.TrimPrefix(rq.URL.Path, "/edit-tag/")
	oldTag := types.Tag{
		Name:        oldName,
		Description: db.DescriptionForTag(oldName),
	}

	if rq.Method == http.MethodGet {
		templateExec(w, templateEditTag, dataEditTag{
			Tag:        oldTag,
			dataCommon: emptyCommon(),
		}, rq)
		return
	}

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
	}

	if !db.TagExists(oldTag.Name) {
		log.Printf("Trying to rename a non-existent tag %s.\n", oldTag.Name)
		templateExec(w, templateEditTag, dataEditTag{
			Tag:              oldTag,
			ErrorNonExistent: true,
			dataCommon:       emptyCommon(),
		}, rq)
		return
	}

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

func handlerDeleteTag(w http.ResponseWriter, rq *http.Request) {
	if rq.Method == http.MethodGet {
		handlerNotFound(w, rq)
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
	errorTemplate
	*dataCommon
	types.Post
	Another bool

	// The following three fields can be non-empty, when set through URL parameters or when an erroneous request was made.
	ErrorEmptyURL      bool
	ErrorInvalidURL    bool
	ErrorTitleNotFound bool
}

func handlerSaveLink(w http.ResponseWriter, rq *http.Request) {
	var viewData dataSaveLink
	var post types.Post

	common := emptyCommon()
	common.head = `<script defer src="/static/autocompletion.js"></script>`

	if rq.Method == http.MethodGet {
		post.URL = rq.FormValue("url")
		post.Title = rq.FormValue("title")
		post.Visibility = types.VisibilityFromString(rq.FormValue("visibility"))
		post.Description = rq.FormValue("description")
		post.Tags = types.SplitTags(rq.FormValue("tags"))
		// TODO: Document the param behaviour
		templateExec(w, templateSaveLink, dataSaveLink{
			Post:       post,
			dataCommon: common,
		}, rq)
		return
	}

	post.URL = rq.FormValue("url")
	post.Title = rq.FormValue("title")
	post.Visibility = types.VisibilityFromString(rq.FormValue("visibility"))
	post.Description = rq.FormValue("description")
	post.Tags = types.SplitTags(rq.FormValue("tags"))

	if post.URL == "" && post.Title == "" {
		viewData.emptyUrl(post, common, w, rq)
		return
	}

	mixUpTitleLink(&post.Title, &post.URL)

	if post.URL == "" {
		viewData.emptyUrl(post, common, w, rq)
		return
	}

	if post.Title == "" {
		if _, err := url.ParseRequestURI(post.URL); err != nil {
			viewData.invalidUrl(post, common, w, rq)
			return
		}
		newTitle, err := readpage.FindTitle(post.URL)
		if err != nil {
			viewData.titleNotFound(post, common, w, rq)
			return
		}
		post.Title = newTitle
	}

	if _, err := url.ParseRequestURI(post.URL); err != nil {
		viewData.invalidUrl(post, common, w, rq)
		return
	}

	id := db.AddPost(post)

	another := rq.FormValue("another")
	if another == "true" {
		var anotherPost types.Post
		anotherPost.Visibility = types.Public
		templateExec(w, templateSaveLink, dataSaveLink{
			dataCommon: common,
			Post:       anotherPost,
			Another:    true,
		}, rq)
		return
	}

	http.Redirect(w, rq, fmt.Sprintf("/%d", id), http.StatusSeeOther)

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
	TotalPosts      uint
	PostsInPage     []types.Post
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
	<link rel="alternate" type="application/rss+xml" title="Daily digest (recommended)" href="/digest-rss">
	<link rel="alternate" type="application/rss+xml" title="Individual posts" href="/posts-rss">
`
	var currentPage uint
	if page, err := strconv.Atoi(rq.FormValue("page")); err != nil || page == 0 {
		currentPage = 1
	} else {
		currentPage = uint(page)
	}

	posts, totalPosts := db.Posts(authed, currentPage)
	common.paginator = types.PaginatorFromURL(rq.URL, currentPage, totalPosts)

	templateExec(w, templateFeed, dataFeed{
		TotalPosts:      totalPosts,
		PostsInPage:     posts,
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
