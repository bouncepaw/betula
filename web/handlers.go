// SPDX-FileCopyrightText: 2023 Danila Gorelko
// SPDX-FileCopyrightText: 2023 ninedraft
// SPDX-FileCopyrightText: 2023 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2023 Umar Getagazov
// SPDX-FileCopyrightText: 2024 Danila Gorelko
// SPDX-FileCopyrightText: 2024 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2025 Danila Gorelko
// SPDX-FileCopyrightText: 2025 Guilherme Puida Moreira
// SPDX-FileCopyrightText: 2025 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package web

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"git.sr.ht/~bouncepaw/betula/feeds"
	"git.sr.ht/~bouncepaw/betula/gateways/activitypub"
	"git.sr.ht/~bouncepaw/betula/pkg/rss"
	apports "git.sr.ht/~bouncepaw/betula/ports/activitypub"
	archivingports "git.sr.ht/~bouncepaw/betula/ports/archiving"
	likingports "git.sr.ht/~bouncepaw/betula/ports/liking"
	"git.sr.ht/~bouncepaw/betula/ports/notif"
	"git.sr.ht/~bouncepaw/betula/svc/archiving"
	likingsvc "git.sr.ht/~bouncepaw/betula/svc/liking"
	"git.sr.ht/~bouncepaw/betula/svc/notif"
	notiftypes "git.sr.ht/~bouncepaw/betula/types/notif"

	"git.sr.ht/~bouncepaw/betula/fediverse"
	"git.sr.ht/~bouncepaw/betula/fediverse/activities"
	"git.sr.ht/~bouncepaw/betula/jobs"
	"git.sr.ht/~bouncepaw/betula/jobs/jobtype"
	"git.sr.ht/~bouncepaw/betula/readpage"

	"git.sr.ht/~bouncepaw/betula/auth"
	"git.sr.ht/~bouncepaw/betula/db"
	"git.sr.ht/~bouncepaw/betula/help"
	"git.sr.ht/~bouncepaw/betula/search"
	"git.sr.ht/~bouncepaw/betula/settings"
	"git.sr.ht/~bouncepaw/betula/types"
)

var (
	//go:embed views/*.gohtml *.css *.js pix/*
	fs embed.FS
	//go:embed bookmarklet.js
	bookmarkletScript string
	mux               = http.NewServeMux()

	// One day, all shall be in services!
	svcNotif     notifports.Service     = notifsvc.New(repoNotif)
	svcArchiving archivingports.Service = archivingsvc.New(
		archivingsvc.NewObeliskFetcher(), db.NewArchivesRepo())
	svcLiking likingports.Service = likingsvc.New(
		repoLike,
		repoLikeCollection,
		repoLocalBookmark,
		repoNotif,
		activityPub)

	repoLike           = db.NewLikeRepo()
	repoLikeCollection = db.NewLikeCollectionRepo()
	repoNotif          = db.New()
	repoActor          = db.NewActorRepo()
	repoLocalBookmark  = db.NewLocalBookmarksRepo()
	repoRemoteBookmark = db.NewRemoteBookmarkRepo()

	activityPub = apgw.NewActivityPub(repoActor, repoRemoteBookmark)
)

func init() {
	mux.HandleFunc("/", handlerNotFound)

	mux.HandleFunc("GET /random", getRandom)

	mux.HandleFunc("GET /{$}", getIndex)
	mux.HandleFunc("GET /{id}", fediverseWebFork(getBookmarkFedi, getBookmarkWeb))

	mux.HandleFunc("GET /help/en/", getEnglishHelp)
	mux.HandleFunc("GET /help", getHelp)
	mux.HandleFunc("GET /text/{id}", getText)
	mux.HandleFunc("GET /digest-rss", getDigestRss)
	mux.HandleFunc("GET /posts-rss", getPostsRss)
	mux.HandleFunc("GET /go/{id}", getGo)
	mux.HandleFunc("GET /about", getAbout)

	mux.HandleFunc("GET /tag", handlerTags)
	mux.HandleFunc("GET /tag/{name}", getTag)

	mux.HandleFunc("GET /day/{dayStamp}", getDay)
	mux.HandleFunc("GET /search", getSearch)
	mux.HandleFunc("GET /static/style.css", getStyle)
	mux.HandleFunc("GET /static/private.js", getPrivateCustomJS)
	mux.HandleFunc("GET /static/public.js", getPublicCustomJS)

	mux.HandleFunc("POST /register", postRegister)

	mux.HandleFunc("GET /login", getLogin)
	mux.HandleFunc("POST /login", postLogin)

	mux.HandleFunc("GET /logout", getLogout)
	mux.HandleFunc("POST /logout", postLogout)

	mux.HandleFunc("GET /settings", adminOnly(getSettings))
	mux.HandleFunc("POST /settings", adminOnly(postSettings))

	mux.HandleFunc("GET /sessions", adminOnly(getSessions))
	mux.HandleFunc("POST /delete-session/{token}", adminOnly(deleteSession))
	mux.HandleFunc("POST /delete-sessions/", adminOnly(deleteSessions))

	mux.HandleFunc("GET /bookmarklet", adminOnly(getBookmarklet))

	// Create & Modify
	mux.HandleFunc("GET /repost", adminOnly(getRepost))
	mux.HandleFunc("POST /repost", adminOnly(postRepost))

	mux.HandleFunc("POST /unrepost/{id}", adminOnly(postUnrepost))

	mux.HandleFunc("GET /save-link", adminOnly(getSaveBookmark))
	mux.HandleFunc("POST /save-link", adminOnly(postSaveBookmark))

	mux.HandleFunc("GET /edit-link/{id}", adminOnly(getEditBookmark))
	mux.HandleFunc("POST /edit-link/{id}", adminOnly(postEditBookmark))

	mux.HandleFunc("POST /edit-link-tags/{id}", adminOnly(postEditBookmarkTags))
	mux.HandleFunc("POST /delete-link/{id}", adminOnly(postDeleteBookmark))

	mux.HandleFunc("GET /edit-tag/{name}", adminOnly(getEditTag))
	mux.HandleFunc("POST /edit-tag/{name}", adminOnly(postEditTag))
	mux.HandleFunc("POST /delete-tag/{name}", adminOnly(postDeleteTag))

	// Notifications
	mux.HandleFunc("GET /notifications", adminOnly(federatedOnly(getNotifications)))
	mux.HandleFunc("POST /notifications/read", adminOnly(federatedOnly(postAllNotificationsRead)))

	// Archives
	mux.HandleFunc("POST /make-new-archive/{id}", adminOnly(postMakeNewArchive))
	mux.HandleFunc("GET /artifact/{slug}", adminOnly(getArtifact))
	mux.HandleFunc("POST /delete-archive", adminOnly(postDeleteArchive))

	// Federation interface
	mux.HandleFunc("POST /follow", adminOnly(federatedOnly(postFollow)))
	mux.HandleFunc("POST /unfollow", adminOnly(federatedOnly(postUnfollow)))
	mux.HandleFunc("GET /following", fediverseWebFork(nil, getFollowingWeb))
	mux.HandleFunc("GET /followers", fediverseWebFork(nil, getFollowersWeb))
	mux.HandleFunc("GET /timeline", adminOnly(federatedOnly(getTimeline)))
	mux.HandleFunc("POST /like", adminOnly(federatedOnly(postLike)))
	mux.HandleFunc("POST /unlike", adminOnly(federatedOnly(postUnlike)))

	// Federated search
	mux.HandleFunc("GET /fedisearch", adminOnly(federatedOnly(handlerFediSearch)))
	mux.HandleFunc("POST /fedisearch", adminOnly(federatedOnly(handlerFediSearch)))
	mux.HandleFunc("POST /.well-known/betula-federated-search", federatedOnly(postFediSearchAPI))

	// ActivityPub
	mux.HandleFunc("POST /inbox", federatedOnly(postInbox))

	// NodeInfo
	mux.HandleFunc("GET /.well-known/nodeinfo", getWellKnownNodeInfo)
	mux.HandleFunc("GET /nodeinfo/2.0", getNodeInfo)

	// WebFinger
	mux.HandleFunc("GET /.well-known/webfinger", federatedOnly(getWebFinger))

	// Static files
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(fs))))

	// The service worker needs to be served from the page root to be registered with the correct scope.
	mux.HandleFunc("GET /service-worker.js", adminOnly(getServiceWorker))
	mux.HandleFunc("GET /manifest.json", getManifest)
}

// Handlers directly related to federation go to handlers_federated.go.
// Others go to this file.

func postAllNotificationsRead(w http.ResponseWriter, rq *http.Request) {
	err := svcNotif.MarkAllAsRead()
	if err != nil {
		slog.Error("Failed to mark all bookmarks as read", "err", err)
		handlerBadRequest(w, rq)
		return
	}
	w.WriteHeader(http.StatusOK)
}

type dataNotifications struct {
	*dataCommon

	Count  int
	Groups []notiftypes.NotificationGroup
}

func getNotifications(w http.ResponseWriter, rq *http.Request) {
	groups, err := svcNotif.GetAll()
	if err != nil {
		slog.Error("Failed to get all notifications", "err", err)
		handlerBadRequest(w, rq)
		return
	}

	var count int
	for _, g := range groups {
		count += len(g.Notifications)
	}

	templateExec(w, rq, templateNotifications, dataNotifications{
		dataCommon: emptyCommon(),
		Count:      count,
		Groups:     groups,
	})
}

func postDeleteArchive(w http.ResponseWriter, rq *http.Request) {
	// This turned out to be much more complex than I wanted it to.
	// Corner cases be damned.
	var (
		err                 error
		archiveIDParameter  = rq.FormValue("archive-id")
		bookmarkIDParameter = rq.FormValue("bookmark-id")
	)
	slog.Info("Request archive deletion",
		"archiveID", archiveIDParameter, "bookmarkID", bookmarkIDParameter)

	var bookmarkID int64
	bookmarkID, err = strconv.ParseInt(bookmarkIDParameter, 10, 64)
	if err != nil {
		slog.Warn("Failed to parse bookmark id",
			"bookmarkID", bookmarkIDParameter, "err", err)
		handlerNotFound(w, rq)
		return
	}

	var bookmark, found = db.GetBookmarkByID(int(bookmarkID))
	if !found {
		slog.Info("Bookmark not found", "bookmarkID", bookmarkID)
		handlerNotFound(w, rq)
		return
	}

	var templateData dataBookmark

	var archiveID int64
	archiveID, err = strconv.ParseInt(archiveIDParameter, 10, 64)
	if err != nil {
		templateData = renderBookmark(bookmark, w, rq, false)
		templateData.Notifications = append(templateData.Notifications,
			SystemNotification{
				Category: NotificationFailure,
				Body: template.HTML(fmt.Sprintf(
					"Failed to parse archive id. ID = %s, err = %s",
					archiveIDParameter, err)),
			})
		slog.Warn("Failed to parse archive id for deletion", "id", archiveID, "err", err)
	} else if err = db.NewArchivesRepo().DeleteArchive(archiveID); err != nil {
		templateData = renderBookmark(bookmark, w, rq, false)
		templateData.Notifications = append(templateData.Notifications,
			SystemNotification{
				Category: NotificationFailure,
				Body: template.HTML(fmt.Sprintf(
					"Failed to delete archive: %s",
					err)),
			})
		slog.Warn("Failed to delete archive", "id", archiveID, "err", err)
	} else {
		templateData = renderBookmark(bookmark, w, rq, false)
		templateData.Notifications = append(templateData.Notifications,
			SystemNotification{
				Category: NotificationSuccess,
				Body:     "Archive deleted",
			})
		slog.Info("Deleted archive", "id", archiveID, "bookmarkID", bookmarkID)
	}

	templateExec(w, rq, templatePost, templateData)
}

func getArtifact(w http.ResponseWriter, rq *http.Request) {
	// TODO: when uses of artifacts other than archives emerge,
	// implement access restrictions. Artifacts that belong to
	// archives shall remain private. That would probably
	// involve changing the database scheme.
	var slug = rq.PathValue("slug")
	var artifact, err = db.NewArtifactsRepo().Fetch(slug)
	if err != nil {
		slog.Warn("Requested artifact does not exist", "id", slug)
		handlerNotFound(w, rq)
		return
	}

	slog.Info("Request artifact", "id", slug, "mime", artifact.MimeType)
	if !artifact.IsGzipped {
		w.Header().Add("Content-Type", artifact.MimeType)
		_, err = w.Write(artifact.Data)
		if err != nil {
			slog.Error("Failed to write artifact data",
				"err", err, "id", slug)
		}
		return
	}

	// TODO: maybe support clients that do not support gzip encoding.
	w.Header().Add("Content-Type", artifact.MimeType)
	w.Header().Add("Content-Encoding", "gzip")
	_, err = w.Write(artifact.Data)
	if err != nil {
		slog.Error("Failed to write artifact data",
			"err", err, "id", slug)
	}
}

func postMakeNewArchive(w http.ResponseWriter, rq *http.Request) {
	var bookmark, ok = extractBookmark(w, rq)
	if !ok {
		return
	}
	slog.Info("Requesting to make a new archive", "bookmarkID", bookmark.ID)

	archiveID, err := svcArchiving.Archive(*bookmark)
	if err != nil {
		handlerBadRequest(w, rq)
		return
	}

	var addr = fmt.Sprintf("/%d?highlight-archive=%d", bookmark.ID, archiveID)
	http.Redirect(w, rq, addr, http.StatusSeeOther)
}

func getRandom(w http.ResponseWriter, rq *http.Request) {
	authed := auth.AuthorizedFromRequest(rq)
	common := emptyCommon()

	bookmarks, totalBookmarks := db.RandomBookmarks(authed, 20)
	groups := types.GroupLocalBookmarksByDate(types.RenderLocalBookmarks(bookmarks))

	templateExec(w, rq, templateFeed, dataFeed{
		Random:               true,
		TotalBookmarks:       totalBookmarks,
		BookmarkGroupsInPage: groups,
		SiteDescription:      settings.SiteDescriptionHTML(),
		dataCommon:           common,
	})
}

type dataAt struct {
	*dataCommon

	Account              types.Actor
	BookmarkGroupsInPage []types.RemoteBookmarkGroup
	TotalBookmarks       uint
}

func handlerAt(w http.ResponseWriter, rq *http.Request) {
	/*
		Show profile. Imagine this Betula's author username is goremyka. Then:

			* /@goremyka resolves to their profile. It is public.
			* /@anything is 404 for everyone.
			* /@boris@godun.ov is not available for strangers, but the Boris's profile for the admin.

		This endpoint is available in both HTML and Activity form.

			* The HTML form shows what you expect. Some bookmarks in the future, maybe. Available for both local profile and remote profiles.
			* The Activity form shows the Actor object. Available for the local profile only.
	*/
	var (
		accept               = rq.Header.Get("Accept")
		wantsActivity        = strings.Contains(accept, types.ActivityType) || strings.Contains(accept, types.OtherActivityType)
		userAtHost           = strings.TrimPrefix(rq.URL.Path, "/@")
		user, host, isRemote = strings.Cut(userAtHost, "@")
		authed               = auth.AuthorizedFromRequest(rq)
		ourUsername          = settings.AdminUsername()
	)

	switch {
	case isRemote && !authed:
		log.Printf("Unauthorized request of remote profile @%s, rejecting\n", userAtHost)
		handlerUnauthorized(w, rq)

	case isRemote && wantsActivity:
		w.Header().Set("Content-Type", types.ActivityType)
		log.Printf("Request remote user %s@%s as an activity, rejecting (HTML only)\n", user, host)
		handlerNotFound(w, rq)

	case isRemote && !wantsActivity:
		log.Printf("Request remote user @%s@%s as a page\n", user, host)

		actor, err := fediverse.RequestActorByNickname(fmt.Sprintf("%s@%s", user, host))
		if err != nil {
			log.Printf("While fetching %s@%s profile, got the error: %s\n", user, host, err)
			handlerNotFound(w, rq)
			return
		}
		actor.SubscriptionStatus = db.SubscriptionStatus(actor.ID)

		currentPage := extractPage(rq)
		bookmarks, total := db.GetRemoteBookmarksBy(actor.ID, currentPage)

		renderedBookmarks := fediverse.RenderRemoteBookmarks(bookmarks)
		if err := svcLiking.FillLikes(rq.Context(), nil, renderedBookmarks); err != nil {
			slog.Error("Failed to fill likes for remote bookmarks", "err", err)
		}

		common := emptyCommon()
		common.searchQuery = actor.Acct()
		common.paginator = types.PaginatorFromURL(rq.URL, currentPage, total)
		templateExec(w, rq, templateRemoteProfile, dataAt{
			dataCommon:           common,
			Account:              *actor,
			BookmarkGroupsInPage: types.GroupRemoteBookmarksByDate(renderedBookmarks),
			TotalBookmarks:       total,
		})

	case !isRemote && userAtHost != ourUsername:
		log.Printf("Request local user @%s, not found\n", userAtHost)
		handlerNotFound(w, rq)
	case !isRemote && wantsActivity:
		log.Printf("Request info about you as an activity\n")
		w.Header().Set("Content-Type", types.ActivityType)
		handlerActor(w, rq)
	case !isRemote && !wantsActivity:
		log.Println("Viewing your profile")
		getMyProfile(w, rq)
	}
}

type dataMyProfile struct {
	*dataCommon

	Nickname                       string
	Summary                        template.HTML
	LinkCount, TagCount            uint
	FollowingCount, FollowersCount uint
	OldestTime, NewestTime         *time.Time
}

func getMyProfile(w http.ResponseWriter, rq *http.Request) {
	authed := auth.AuthorizedFromRequest(rq)
	templateExec(w, rq, templateMyProfile, dataMyProfile{
		dataCommon: emptyCommon(),

		Nickname:       fmt.Sprintf("@%s@%s", settings.AdminUsername(), settings.SiteDomain()),
		Summary:        settings.SiteDescriptionHTML(),
		LinkCount:      db.BookmarkCount(authed),
		TagCount:       db.TagCount(authed),
		FollowingCount: db.CountFollowing(),
		FollowersCount: db.CountFollowers(),
		OldestTime:     db.OldestTime(authed),
		NewestTime:     db.NewestTime(authed),
	})
}

type dataBookmarklet struct {
	*dataCommon
	Script string
}

func getBookmarklet(w http.ResponseWriter, rq *http.Request) {
	templateExec(w, rq, templateBookmarklet, dataBookmarklet{
		dataCommon: emptyCommon(),
		Script:     fmt.Sprintf(bookmarkletScript, settings.SiteURL()),
	})
}

func getHelp(w http.ResponseWriter, rq *http.Request) {
	http.Redirect(w, rq, "/help/en/index", http.StatusSeeOther)
}

type dataHelp struct {
	*dataCommon
	This   help.Topic
	Topics []help.Topic
}

func getEnglishHelp(w http.ResponseWriter, rq *http.Request) {
	topicName := strings.TrimPrefix(rq.URL.Path, "/help/en/")
	if topicName == "/help/en" || topicName == "/" {
		topicName = "index"
	}
	topic, found := help.GetEnglishHelp(topicName)
	if !found {
		handlerNotFound(w, rq)
		return
	}

	templateExec(w, rq, templateHelp, dataHelp{
		dataCommon: emptyCommon(),
		This:       topic,
		Topics:     help.Topics,
	})
}

func getPrivateCustomJS(w http.ResponseWriter, rq *http.Request) {
	var js = settings.PrivateCustomJS()
	if js == "" {
		slog.Info("No custom private JS found; 404")
		http.NotFound(w, rq)
		return
	}

	w.Header().Set("Content-Type", "text/javascript")
	var _, err = io.WriteString(w, js)
	if err != nil {
		slog.Error("Failed to serve private custom JS", "err", err)
	}
}

func getPublicCustomJS(w http.ResponseWriter, rq *http.Request) {
	var js = settings.PublicCustomJS()
	if js == "" {
		slog.Info("No custom public JS found; 404")
		http.NotFound(w, rq)
		return
	}

	w.Header().Set("Content-Type", "text/javascript")
	var _, err = io.WriteString(w, js)
	if err != nil {
		slog.Error("Failed to serve public custom JS", "err", err)
	}
}

func getStyle(w http.ResponseWriter, rq *http.Request) {
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
	Query                string
	TotalBookmarks       uint
	BookmarkGroupsInPage []types.LocalBookmarkGroup
}

var tagOnly = regexp.MustCompile(`^#([^?!:#@<>*|'"&%{}\\\s]+)\s*$`)
var usernameOnly = regexp.MustCompile(`^@.*@.*$`)

func getSearch(w http.ResponseWriter, rq *http.Request) {
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

	if usernameOnly.MatchString(query) {
		http.Redirect(w, rq, query, http.StatusSeeOther)
		return
	}

	authed := auth.AuthorizedFromRequest(rq)
	currentPage := extractPage(rq)
	bookmarks, totalBookmarks := search.For(query, authed, currentPage)

	renderedBookmarks := types.RenderLocalBookmarks(bookmarks)
	if err := svcLiking.FillLikes(rq.Context(), renderedBookmarks, nil); err != nil {
		slog.Error("Failed to fill likes for local bookmarks", "err", err)
	}
	groups := types.GroupLocalBookmarksByDate(renderedBookmarks)

	common := emptyCommon()
	common.paginator = types.PaginatorFromURL(rq.URL, currentPage, totalBookmarks)
	common.searchQuery = query
	log.Printf("Searching ‘%s’. Authorized: %v\n", query, authed)
	templateExec(w, rq, templateSearch, dataSearch{
		dataCommon:           common,
		Query:                query,
		BookmarkGroupsInPage: groups,
		TotalBookmarks:       totalBookmarks,
	})
}

func getText(w http.ResponseWriter, rq *http.Request) {
	bookmark, ok := extractBookmark(w, rq)
	if !ok {
		return
	}

	log.Printf("Fetching text for bookmark no. %d\n", bookmark.ID)

	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, bookmark.Description)
}

func writeFeed(fd *rss.Feed, w http.ResponseWriter) {
	err := fd.Write(w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, err.Error())
	}
}

func getPostsRss(w http.ResponseWriter, _ *http.Request) {
	writeFeed(feeds.Posts(), w)
}

func getDigestRss(w http.ResponseWriter, _ *http.Request) {
	writeFeed(feeds.Digest(), w)
}

var dayStampRegex = regexp.MustCompile("^[0-9]{4}-[0-9]{2}-[0-9]{2}$")

type dataDay struct {
	*dataCommon
	DayStamp  string
	Bookmarks []types.RenderedLocalBookmark
}

func getDay(w http.ResponseWriter, rq *http.Request) {
	authed := auth.AuthorizedFromRequest(rq)
	dayStamp := rq.PathValue("dayStamp")
	// If no day given, default to today.
	if dayStamp == "" {
		now := time.Now()
		dayStamp = fmt.Sprintf("%d-%02d-%02d", now.Year(), now.Month(), now.Day())
	} else if !dayStampRegex.MatchString(dayStamp) {
		handlerNotFound(w, rq)
		return
	}

	bookmarks := types.RenderLocalBookmarks(db.BookmarksForDay(authed, dayStamp))
	if err := svcLiking.FillLikes(rq.Context(), bookmarks, nil); err != nil {
		slog.Error("Failed to fill likes for local bookmarks", "err", err)
	}

	templateExec(w, rq, templateDay, dataDay{
		dataCommon: emptyCommon(),
		DayStamp:   dayStamp,
		Bookmarks:  bookmarks,
	})
}

type dataSettings struct {
	types.Settings
	*dataCommon
	ErrBadPort  bool
	FirstRun    bool
	RequestHost string
}

func getSettings(w http.ResponseWriter, rq *http.Request) {
	isFirstRun := rq.FormValue("first-run") == "true"
	templateExec(w, rq, templateSettings, dataSettings{
		Settings: types.Settings{
			NetworkHost:               settings.NetworkHost(),
			NetworkPort:               settings.NetworkPort(),
			SiteName:                  settings.SiteName(),
			SiteTitle:                 settings.SiteTitle(),
			SiteDescriptionMycomarkup: settings.SiteDescriptionMycomarkup(),
			SiteURL:                   settings.SiteURL(),
			CustomCSS:                 settings.CustomCSS(),
			FederationEnabled:         settings.FederationEnabled(),
			PublicCustomJS:            settings.PublicCustomJS(),
			PrivateCustomJS:           settings.PrivateCustomJS(),
		},
		dataCommon:  emptyCommon(),
		FirstRun:    isFirstRun,
		RequestHost: rq.Host,
	})
	return
}

func postSettings(w http.ResponseWriter, rq *http.Request) {
	isFirstRun := rq.FormValue("first-run") == "true"
	var newSettings = types.Settings{
		NetworkHost:               rq.FormValue("network-host"),
		SiteName:                  rq.FormValue("site-name"),
		SiteTitle:                 template.HTML(rq.FormValue("site-title")),
		SiteDescriptionMycomarkup: rq.FormValue("site-description"),
		SiteURL:                   rq.FormValue("site-url"),
		CustomCSS:                 rq.FormValue("custom-css"),
		FederationEnabled:         rq.FormValue("enable-federation") == "true",
		PublicCustomJS:            rq.FormValue("public-custom-js"),
		PrivateCustomJS:           rq.FormValue("private-custom-js"),
	}

	// If the port ≤ 0 or not really numeric, show error.
	if port, err := strconv.Atoi(rq.FormValue("network-port")); err != nil || port <= 0 {
		newSettings.NetworkPort = uint(port)
		templateExec(w, rq, templateSettings, dataSettings{
			Settings:   newSettings,
			ErrBadPort: true,
			dataCommon: emptyCommon(),
		})
		return
	} else {
		newSettings.NetworkPort = settings.ValidatePortFromWeb(port)
	}

	oldPort := settings.NetworkPort()
	oldHost := settings.NetworkHost()
	settings.SetSettings(newSettings)
	activities.GenerateBetulaActor()
	if oldPort != settings.NetworkPort() || oldHost != settings.NetworkHost() {
		restartServer()
	}
	if isFirstRun {
		http.Redirect(w, rq, "/", http.StatusSeeOther)
	} else {
		http.Redirect(w, rq, "/settings", http.StatusSeeOther)
	}
}

func postDeleteBookmark(w http.ResponseWriter, rq *http.Request) {
	id, ok := extractBookmarkID(w, rq)
	if !ok {
		return
	}

	if confirmed := rq.FormValue("confirmed"); confirmed != "true" {
		http.Redirect(w, rq, fmt.Sprintf("/edit-link/%d", id), http.StatusSeeOther)
		return
	}

	bookmark, found := db.GetBookmarkByID(id)

	if !found {
		log.Println("Trying to delete a non-existent bookmark.")
		handlerNotFound(w, rq)
		return
	}

	db.DeleteBookmark(id)
	http.Redirect(w, rq, "/", http.StatusSeeOther)

	if settings.FederationEnabled() {
		go func(bookmark types.Bookmark) {
			if bookmark.Visibility != types.Public {
				return
			}
			data, err := activities.DeleteNote(bookmark.ID)
			if err != nil {
				log.Printf("When creating Delete{Note} activity for bookmark no. %d: %s\n", id, err)
				return
			}
			jobs.ScheduleDatum(jobtype.SendDeleteNote, data)
		}(bookmark)
	}
}

type dataSessions struct {
	Sessions []types.Session
	*dataCommon
}

func getSessions(w http.ResponseWriter, rq *http.Request) {
	currentToken, err := auth.Token(rq)
	if err != nil {
		handlerUnauthorized(w, rq)
		return
	}
	sessions := auth.MarkCurrentSession(currentToken, auth.Sessions())
	templateExec(w, rq, templateSessions, dataSessions{
		Sessions:   sessions,
		dataCommon: emptyCommon(),
	})
	return
}

func deleteSession(w http.ResponseWriter, rq *http.Request) {
	token := rq.PathValue("token")
	db.StopSession(token)
	http.Redirect(w, rq, "/sessions", http.StatusSeeOther)
}

func deleteSessions(w http.ResponseWriter, rq *http.Request) {
	token, err := auth.Token(rq)
	if err != nil {
		handlerUnauthorized(w, rq)
		return
	}
	db.StopAllSessions(token)
	http.Redirect(w, rq, "/sessions", http.StatusSeeOther)
}

func handlerBadRequest(w http.ResponseWriter, rq *http.Request) {
	slog.Error("400 Bad Request", "url", rq.URL.Path)
	w.WriteHeader(http.StatusBadRequest)
	templateExec(w, rq, templateStatus, dataAuthorized{
		dataCommon: emptyCommon(),
		Status:     http.StatusText(http.StatusBadRequest),
	})
}

func handlerNotFound(w http.ResponseWriter, rq *http.Request) {
	log.Printf("404 Not found: %s\n", rq.URL.Path)
	w.WriteHeader(http.StatusNotFound)
	templateExec(w, rq, templateStatus, dataAuthorized{
		dataCommon: emptyCommon(),
		Status:     http.StatusText(http.StatusNotFound),
	})
}

func handlerUnauthorized(w http.ResponseWriter, rq *http.Request) {
	log.Printf("401 Unauthorized: %s\n", rq.URL.Path)
	w.WriteHeader(http.StatusUnauthorized)
	templateExec(w, rq, templateStatus, dataAuthorized{
		dataCommon: emptyCommon(),
		Status:     http.StatusText(http.StatusUnauthorized),
	})
}

func handlerNotFederated(w http.ResponseWriter, rq *http.Request) {
	// TODO: a proper separate error page!
	log.Printf("404 Not found + Not federated: %s\n", rq.URL.Path)
	w.WriteHeader(http.StatusNotFound)
	templateExec(w, rq, templateStatus, dataAuthorized{
		dataCommon: emptyCommon(),
		Status:     "Not federated",
	})
}

func getLogout(w http.ResponseWriter, rq *http.Request) {
	templateExec(w, rq, templateLogoutForm, dataAuthorized{
		dataCommon: emptyCommon(),
	})
	return
}

func postLogout(w http.ResponseWriter, rq *http.Request) {
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

func getLogin(w http.ResponseWriter, rq *http.Request) {
	templateExec(w, rq, templateLoginForm, dataLogin{
		dataCommon: emptyCommon(),
	})
	return
}

func postLogin(w http.ResponseWriter, rq *http.Request) {
	var (
		name      = rq.FormValue("name")
		pass      = rq.FormValue("pass")
		userAgent = rq.Header.Get("User-Agent")
	)

	if !auth.CredentialsMatch(name, pass) {
		// If incorrect password, ask the client to try again.
		w.WriteHeader(http.StatusBadRequest)
		templateExec(w, rq, templateLoginForm, dataLogin{
			Name:       name,
			Pass:       pass,
			Incorrect:  true,
			dataCommon: emptyCommon(),
		})
		return
	}

	auth.LogInResponse(userAgent, w)
	// TODO: Redirect to the previous (?) location, whatever it is
	http.Redirect(w, rq, "/", http.StatusSeeOther)
}

func postRegister(w http.ResponseWriter, rq *http.Request) {
	log.Println("/register")
	if auth.Ready() {
		// TODO: Let admin change credentials.
		log.Println("Cannot reregister")
		return
	}
	var (
		name      = rq.FormValue("name")
		pass      = rq.FormValue("pass")
		userAgent = rq.Header.Get("User-Agent")
	)
	auth.SetCredentials(name, pass)
	auth.LogInResponse(userAgent, w)
	http.Redirect(w, rq, "/settings?first-run=true", http.StatusSeeOther)
}

type dataTags struct {
	*dataCommon
	Tags []types.Tag
}

func handlerTags(w http.ResponseWriter, rq *http.Request) {
	authed := auth.AuthorizedFromRequest(rq)
	templateExec(w, rq, templateTags, dataTags{
		Tags:       db.Tags(authed),
		dataCommon: emptyCommon(),
	})
}

type dataTag struct {
	*dataCommon
	types.Tag
	TotalBookmarks       uint
	BookmarkGroupsInPage []types.LocalBookmarkGroup
}

func getTag(w http.ResponseWriter, rq *http.Request) {
	tagName := rq.PathValue("name")
	currentPage := extractPage(rq)
	authed := auth.AuthorizedFromRequest(rq)

	bookmarks, totalBookmarks := db.BookmarksWithTag(authed, tagName, currentPage)
	renderedBookmarks := types.RenderLocalBookmarks(bookmarks)
	if err := svcLiking.FillLikes(rq.Context(), renderedBookmarks, nil); err != nil {
		slog.Error("Failed to fill likes for local bookmarks", "err", err)
	}
	groups := types.GroupLocalBookmarksByDate(renderedBookmarks)

	common := emptyCommon()
	common.searchQuery = "#" + tagName
	common.paginator = types.PaginatorFromURL(rq.URL, currentPage, totalBookmarks)
	templateExec(w, rq, templateTag, dataTag{
		Tag: types.Tag{
			Name:        tagName,
			Description: db.DescriptionForTag(tagName),
		},
		BookmarkGroupsInPage: groups,
		TotalBookmarks:       totalBookmarks,
		dataCommon:           common,
	})
}

type dataAbout struct {
	*dataCommon
	SiteDescription template.HTML
}

func getAbout(w http.ResponseWriter, rq *http.Request) {
	templateExec(w, rq, templateAbout, dataAbout{
		dataCommon:      emptyCommon(),
		SiteDescription: settings.SiteDescriptionHTML(),
	})
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

func postEditBookmarkTags(w http.ResponseWriter, rq *http.Request) {
	id, ok := extractBookmarkID(w, rq)
	if !ok {
		return
	}

	tags := types.SplitTags(rq.FormValue("tags"))
	db.SetTagsFor(id, tags)

	next := rq.FormValue("next")
	http.Redirect(w, rq, next, http.StatusSeeOther)

	if settings.FederationEnabled() {
		go func(id int) {
			// the handler never modifies the bookmark visibility, so we don't care about the past, so we only look at the current value
			bookmark, found := db.GetBookmarkByID(id)
			if !found {
				log.Printf("When federating bookmark no. %d: bookmark not found\n", bookmark.ID)
				return
			}
			if bookmark.Visibility != types.Public {
				return
			}

			// The bookmark remains public
			data, err := activities.UpdateNote(bookmark)
			if err != nil {
				log.Printf("When creating Update{Note} activity for bookmark no. %d: %s\n", bookmark.ID, err)
				return
			}
			jobs.ScheduleDatum(jobtype.SendUpdateNote, data)
		}(id)
	}
}

type dataEditLink struct {
	errorTemplate
	*dataCommon
	types.Bookmark
	ErrorEmptyURL      bool
	ErrorInvalidURL    bool
	ErrorTitleNotFound bool

	DuplicateBookmarkID int
}

func getEditBookmark(w http.ResponseWriter, rq *http.Request) {
	bookmark, ok := extractBookmark(w, rq)
	if !ok {
		return
	}

	bookmark.Tags = db.TagsForBookmarkByID(bookmark.ID)
	templateExec(w, rq, templateEditLink, dataEditLink{
		Bookmark:   *bookmark,
		dataCommon: commonWithAutoCompletion(),
	})
	return
}

func postEditBookmark(w http.ResponseWriter, rq *http.Request) {
	bookmark, ok := extractBookmark(w, rq)
	if !ok {
		return
	}

	common := commonWithAutoCompletion()

	oldVisibility := bookmark.Visibility

	if rq.Method == http.MethodGet {
		bookmark.Tags = db.TagsForBookmarkByID(bookmark.ID)
		templateExec(w, rq, templateEditLink, dataEditLink{
			Bookmark:   *bookmark,
			dataCommon: common,
		})
		return
	}

	oldURL := bookmark.URL
	newURL := rq.FormValue("url")
	bookmark.URL = newURL
	bookmark.Title = rq.FormValue("title")
	bookmark.Visibility = types.VisibilityFromString(rq.FormValue("visibility"))
	bookmark.Description = rq.FormValue("description")

	// If this is true, a user can edit a bookmark with the URL of another bookmark next time they click 'Save' button.
	saveDuplicate := rq.FormValue("duplicate") == "true"

	if oldURL != newURL && !saveDuplicate {
		existingBookmarkID, found := db.GetBookmarkIDByURL(newURL)
		if found {
			templateExec(w, rq, templateEditLink, dataEditLink{
				Bookmark: *bookmark,
				dataCommon: commonWithAutoCompletion().
					withSystemNotifications(
						SystemNotification{
							Category: NotificationClarification,
							Body:     template.HTML(fmt.Sprintf(`A bookmark with this URL <a href="%d">already exists</a>.`, existingBookmarkID)),
						}),
				DuplicateBookmarkID: existingBookmarkID,
			})
			return
		}
	}

	bookmark.Tags = types.SplitTags(rq.FormValue("tags"))

	var viewData dataEditLink

	if bookmark.URL == "" && bookmark.Title == "" {
		viewData.emptyUrl(*bookmark, common, w, rq)
		return
	}

	mixUpTitleLink(&bookmark.Title, &bookmark.URL)

	if bookmark.URL == "" {
		viewData.emptyUrl(*bookmark, common, w, rq)
		return
	}

	if bookmark.Title == "" {
		if _, err := url.ParseRequestURI(bookmark.URL); err != nil {
			viewData.invalidUrl(*bookmark, common, w, rq)
			return
		}
		newTitle, err := readpage.FindTitle(bookmark.URL)
		if err != nil {
			log.Printf("Can't get HTML title from URL: %s\n", bookmark.URL)
			viewData.titleNotFound(*bookmark, common, w, rq)
			return
		}
		bookmark.Title = newTitle
	}

	if _, err := url.ParseRequestURI(bookmark.URL); err != nil {
		log.Printf("Invalid URL was passed, asking again: %s\n", bookmark.URL)
		viewData.invalidUrl(*bookmark, common, w, rq)
		return
	}

	db.EditBookmark(*bookmark)
	http.Redirect(w, rq, fmt.Sprintf("/%d", bookmark.ID), http.StatusSeeOther)
	log.Printf("Edited bookmark no. %d\n", bookmark.ID)

	if settings.FederationEnabled() {
		go func(post types.Bookmark, oldVisibility types.Visibility) {
			wasPublic := oldVisibility == types.Public
			isPublic := post.Visibility == types.Public

			// The post remains private.
			if !wasPublic && !isPublic {
				return
			}

			// The post was hidden by the author. Let's broadcast Delete.
			if wasPublic && !isPublic {
				data, err := activities.DeleteNote(post.ID)
				if err != nil {
					log.Printf("When creating Delete{Note} activity for post no. %d: %s\n", post.ID, err)
					return
				}
				jobs.ScheduleDatum(jobtype.SendDeleteNote, data)
				return
			}

			post.CreationTime = time.Now().UTC().Format(types.TimeLayout) // It shall match the one generated in DB

			// The post was unpublic, but became public. Let's broadcast Create.
			if !wasPublic && isPublic {
				data, err := activities.CreateNote(post)
				if err != nil {
					log.Printf("When creating Create{Note} activity for post no. %d: %s\n", post.ID, err)
					return
				}
				jobs.ScheduleDatum(jobtype.SendCreateNote, data)
				return
			}

			// The post remains public
			data, err := activities.UpdateNote(post)
			if err != nil {
				log.Printf("When creating Update{Note} activity for post no. %d: %s\n", post.ID, err)
				return
			}
			jobs.ScheduleDatum(jobtype.SendUpdateNote, data)
		}(*bookmark, oldVisibility)
	}
}

type dataEditTag struct {
	*dataCommon
	types.Tag
	ErrorTakenName   bool
	ErrorNonExistent bool
}

func oldTag(rq *http.Request) types.Tag {
	oldName := rq.PathValue("name")
	return types.Tag{
		Name:        oldName,
		Description: db.DescriptionForTag(oldName),
	}
}

func getEditTag(w http.ResponseWriter, rq *http.Request) {
	templateExec(w, rq, templateEditTag, dataEditTag{
		Tag:        oldTag(rq),
		dataCommon: emptyCommon(),
	})
}

func postEditTag(w http.ResponseWriter, rq *http.Request) {
	var newTag types.Tag
	newName := types.CanonicalTagName(rq.FormValue("new-name"))
	newTag.Name = newName
	newTag.Description = strings.TrimSpace(rq.FormValue("description"))

	merge := rq.FormValue("merge")

	oldTag := oldTag(rq)

	if db.TagExists(newTag.Name) && merge != "true" && newTag.Name != oldTag.Name {
		log.Printf("Trying to rename a tag %s to a taken name %s.\n", oldTag.Name, newTag.Name)
		templateExec(w, rq, templateEditTag, dataEditTag{
			Tag:            oldTag,
			ErrorTakenName: true,
			dataCommon:     emptyCommon(),
		})
		return
	}

	if !db.TagExists(oldTag.Name) {
		log.Printf("Trying to rename a non-existent tag %s.\n", oldTag.Name)
		templateExec(w, rq, templateEditTag, dataEditTag{
			Tag:              oldTag,
			ErrorNonExistent: true,
			dataCommon:       emptyCommon(),
		})
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

func postDeleteTag(w http.ResponseWriter, rq *http.Request) {
	catName := rq.PathValue("name")
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
	types.Bookmark
	Another bool

	// The following three fields can be non-empty, when an erroneous request was made.
	ErrorEmptyURL      bool
	ErrorInvalidURL    bool
	ErrorTitleNotFound bool

	DuplicateBookmarkID int
}

func getSaveBookmark(w http.ResponseWriter, rq *http.Request) {
	var bookmark types.Bookmark

	bookmark.URL = rq.FormValue("url")
	bookmark.Title = rq.FormValue("title")
	bookmark.Visibility = types.VisibilityFromString(rq.FormValue("visibility"))
	bookmark.Description = rq.FormValue("description")
	bookmark.Tags = types.SplitTags(rq.FormValue("tags"))

	// When sharing a web page via the web-share API on Chrome or Firefox, the URL of the shared page
	// is placed on the "description" query parameter instead of the "url" parameter.
	// If "url" is empty and "description" starts with "http", we can reasonably assume that the
	// "description" is the actual URL.
	if bookmark.URL == "" && strings.HasPrefix(bookmark.Description, "http") {
		bookmark.URL = bookmark.Description
		bookmark.Description = ""
	}

	// TODO: Document the param behaviour
	templateExec(w, rq, templateSaveLink, dataSaveLink{
		Bookmark:   bookmark,
		dataCommon: commonWithAutoCompletion(),
	})
	return
}

func postSaveBookmark(w http.ResponseWriter, rq *http.Request) {
	var viewData dataSaveLink
	var bookmark types.Bookmark

	common := commonWithAutoCompletion()

	bookmark.URL = rq.FormValue("url")
	bookmark.Title = rq.FormValue("title")
	bookmark.Visibility = types.VisibilityFromString(rq.FormValue("visibility"))
	bookmark.Description = rq.FormValue("description")
	bookmark.Tags = types.SplitTags(rq.FormValue("tags"))

	// If this is true, a user can save a duplicate next time they click 'Save' button.
	saveDuplicate := rq.FormValue("duplicate") == "true"

	if bookmark.URL == "" && bookmark.Title == "" {
		viewData.emptyUrl(bookmark, common, w, rq)
		return
	}

	mixUpTitleLink(&bookmark.Title, &bookmark.URL)

	if bookmark.URL == "" {
		viewData.emptyUrl(bookmark, common, w, rq)
		return
	}

	if bookmark.Title == "" {
		if _, err := url.ParseRequestURI(bookmark.URL); err != nil {
			viewData.invalidUrl(bookmark, common, w, rq)
			return
		}
		newTitle, err := readpage.FindTitle(bookmark.URL)
		if err != nil {
			viewData.titleNotFound(bookmark, common, w, rq)
			return
		}
		bookmark.Title = newTitle
	}

	if _, err := url.ParseRequestURI(bookmark.URL); err != nil {
		viewData.invalidUrl(bookmark, common, w, rq)
		return
	}

	// Check if a bookmark with this URL already exists.
	if !saveDuplicate {
		existingBookmarkID, found := db.GetBookmarkIDByURL(bookmark.URL)
		if found {
			bookmark.ID = existingBookmarkID
			templateExec(w, rq, templateSaveLink, dataSaveLink{
				Bookmark: bookmark,
				dataCommon: commonWithAutoCompletion().
					withSystemNotifications(
						SystemNotification{
							Category: NotificationClarification,
							Body:     template.HTML(fmt.Sprintf(`A bookmark with this URL <a href="%d">already exists</a>.`, existingBookmarkID)),
						}),
				DuplicateBookmarkID: existingBookmarkID,
			})
			return
		}
	}

	// No duplicate found, insert a new bookmark
	id := db.InsertBookmark(bookmark)
	bookmark.ID = int(id)

	another := rq.FormValue("another")
	if another == "true" {
		var anotherPost types.Bookmark
		anotherPost.Visibility = types.Public
		templateExec(w, rq, templateSaveLink, dataSaveLink{
			dataCommon: common,
			Bookmark:   anotherPost,
			Another:    true,
		})
		return
	}

	http.Redirect(w, rq, fmt.Sprintf("/%d", id), http.StatusSeeOther)

	if settings.FederationEnabled() {
		go func(bookmark types.Bookmark) {
			if bookmark.Visibility != types.Public {
				return
			}
			bookmark.CreationTime = time.Now().UTC().Format(types.TimeLayout) // It shall match the one generated in DB
			data, err := activities.CreateNote(bookmark)
			if err != nil {
				log.Printf("When creating Create{Note} activity for post no. %d: %s\n", id, err)
				return
			}
			jobs.ScheduleDatum(jobtype.SendCreateNote, data)
		}(bookmark)
	}
}

type dataBookmark struct {
	Bookmark types.Bookmark
	Reposts  []types.RepostInfo

	LikeCounter int
	LikedByUs   bool
	Likes       []apports.Actor

	Archives         []types.Archive
	HighlightArchive int64
	*dataCommon

	Notifications []SystemNotification
}

func getBookmarkWeb(w http.ResponseWriter, rq *http.Request) {
	bookmark, ok := extractBookmark(w, rq)
	if !ok {
		return
	}
	log.Printf("Get bookmark page no. %d\n", bookmark.ID)
	var data = renderBookmark(*bookmark, w, rq, true)
	templateExec(w, rq, templatePost, data)
}

func renderBookmark(
	bookmark types.Bookmark,
	w http.ResponseWriter,
	rq *http.Request,
	includeLikes bool,
) dataBookmark {
	var notifications []SystemNotification

	// TODO: do not fetch for the unauthorized
	archivesRepo := db.NewArchivesRepo()
	archives, err := archivesRepo.FetchForBookmark(int64(bookmark.ID))
	if err != nil {
		slog.Error("Failed to fetch archives for bookmark",
			"bookmarkID", bookmark.ID,
			"err", err)
		notifications = append(notifications,
			SystemNotification{
				Category: NotificationFailure,
				Body: template.HTML(fmt.Sprintf(
					"Failed to fetch archives: %s", err)),
			})
	}

	common := emptyCommon()
	common.head = template.HTML(fmt.Sprintf(`<link rel="alternate" type="text/mycomarkup" href="/text/%d">`, bookmark.ID))
	if bookmark.RepostOf == nil {
		common.head += template.HTML(fmt.Sprintf(`
<link rel="alternate" type="%s" href="/%d"'>`, types.OtherActivityType, bookmark.ID))
	}

	var highlightArchive int64
	{
		var val = rq.FormValue("highlight-archive")
		if val != "" {
			var n, err = strconv.Atoi(val)
			if err != nil {
				slog.Warn("Invalid value for highlight-archive",
					"val", val, "err", err)
			} else {
				highlightArchive = int64(n)
			}
		}
	}

	bookmark.Tags = db.TagsForBookmarkByID(bookmark.ID)

	var reposts []types.RepostInfo
	if r, err := db.RepostsOf(bookmark.ID); err != nil {
		slog.Warn("Failed to fetch reposts for bookmark", "bookmarkID", bookmark.ID, "err", err)
		notifications = append(notifications,
			SystemNotification{
				Category: NotificationFailure,
				Body: template.HTML(fmt.Sprintf(
					"Failed to fetch reposts: %s", err)),
			})
	} else {
		reposts = r
	}

	var (
		likes       []apports.Actor
		likedByUs   bool
		likeCounter int
	)
	if includeLikes {
		likes, likedByUs, err = svcLiking.ActorsThatLiked(rq.Context(), bookmark.ID)
		if err != nil {
			slog.Warn("Failed to fetch likes for bookmark",
				"bookmarkID", bookmark.ID, "err", err)
			notifications = append(notifications,
				SystemNotification{
					Category: NotificationFailure,
					Body: template.HTML(fmt.Sprintf(
						"Failed to fetch likes: %s", err)),
				})
		}

		likeCounter = len(likes)
		if likedByUs {
			likeCounter++
		}
	}

	return dataBookmark{
		Bookmark:         bookmark,
		Reposts:          reposts,
		Archives:         archives,
		HighlightArchive: highlightArchive,
		dataCommon:       common,
		Notifications:    notifications,

		LikeCounter: likeCounter,
		LikedByUs:   likedByUs,
		Likes:       likes,
	}
}

type dataFeed struct {
	Random               bool
	TotalBookmarks       uint
	BookmarkGroupsInPage []types.LocalBookmarkGroup
	SiteDescription      template.HTML
	*dataCommon
}

func getIndex(w http.ResponseWriter, rq *http.Request) {
	authed := auth.AuthorizedFromRequest(rq)
	common := emptyCommon()
	common.head = `
	<link rel="alternate" type="application/rss+xml" title="Daily digest (recommended)" href="/digest-rss">
	<link rel="alternate" type="application/rss+xml" title="Individual posts" href="/posts-rss">
`

	currentPage := extractPage(rq)
	bookmarks, totalBookmarks := db.Bookmarks(authed, currentPage)
	renderedBookmarks := types.RenderLocalBookmarks(bookmarks)
	if err := svcLiking.FillLikes(rq.Context(), renderedBookmarks, nil); err != nil {
		slog.Error("Failed to fill likes for local bookmarks", "err", err)
	}
	groups := types.GroupLocalBookmarksByDate(renderedBookmarks)

	common.paginator = types.PaginatorFromURL(rq.URL, currentPage, totalBookmarks)

	templateExec(w, rq, templateFeed, dataFeed{
		TotalBookmarks:       totalBookmarks,
		BookmarkGroupsInPage: groups,
		SiteDescription:      settings.SiteDescriptionHTML(),
		dataCommon:           common,
	})
}

func getServiceWorker(w http.ResponseWriter, r *http.Request) {
	http.ServeFileFS(w, r, fs, "service-worker.js")
}

func getManifest(w http.ResponseWriter, r *http.Request) {
	manifest, err := json.Marshal(map[string]any{
		"name":      settings.SiteName(),
		"display":   "standalone",
		"start_url": "/",
		"share_target": map[string]any{
			"action": "/save-link",
			"method": "GET",
			"params": map[string]any{
				"url":   "url",
				"title": "title",
				"text":  "description",
			},
		},
		"icons": []any{
			map[string]any{
				"src":   "static/pix/icon-512.png",
				"type":  "image/png",
				"sizes": "512x512",
			},
			map[string]any{
				"src":   "static/pix/icon-192.png",
				"type":  "image/png",
				"sizes": "192x192",
			},
		},
	})

	if err != nil {
		log.Printf("When marhalling manifest.json: %s", err)
		handlerNotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(manifest); err != nil {
		log.Printf("When serving manifest.json: %s", err)
	}
}

func getGo(w http.ResponseWriter, rq *http.Request) {
	bookmark, ok := extractBookmark(w, rq)
	if !ok {
		return
	}
	http.Redirect(w, rq, bookmark.URL, http.StatusSeeOther)
}
