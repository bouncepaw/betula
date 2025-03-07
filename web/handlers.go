package web

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"git.sr.ht/~bouncepaw/betula/archiving"
	"git.sr.ht/~bouncepaw/betula/fediverse"
	"git.sr.ht/~bouncepaw/betula/fediverse/activities"
	"git.sr.ht/~bouncepaw/betula/fediverse/signing"
	"git.sr.ht/~bouncepaw/betula/jobs"
	"git.sr.ht/~bouncepaw/betula/jobs/jobtype"
	"git.sr.ht/~bouncepaw/betula/readpage"
	"git.sr.ht/~bouncepaw/betula/stricks"
	"html/template"
	"humungus.tedunangst.com/r/webs/rss"
	"io"
	"log"
	"log/slog"
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
	//go:embed views/*.gohtml *.css *.js pix/*
	fs embed.FS
	//go:embed bookmarklet.js
	bookmarkletScript string
	mux               = http.NewServeMux()
)

func init() {
	mux.HandleFunc("/", handlerNotFound)

	mux.HandleFunc("GET /random", getRandom)

	mux.HandleFunc("GET /{$}", getIndex)
	mux.HandleFunc("GET /{id}", fediverseWebFork(getBookmarkFedi, getBookmarkWeb))

	mux.HandleFunc("GET /reposts-of/{id}", getRepostsOf)
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

	// Archives
	mux.HandleFunc("POST /make-new-archive/{id}", adminOnly(postMakeNewArchive))
	mux.HandleFunc("GET /artifact/{slug}", adminOnly(getArtifact))

	// Federation interface
	mux.HandleFunc("POST /follow", adminOnly(federatedOnly(postFollow)))
	mux.HandleFunc("POST /unfollow", adminOnly(federatedOnly(postUnfollow)))
	mux.HandleFunc("GET /following", fediverseWebFork(nil, getFollowingWeb))
	mux.HandleFunc("GET /followers", fediverseWebFork(nil, getFollowersWeb))
	mux.HandleFunc("GET /timeline", adminOnly(federatedOnly(getTimeline)))

	// ActivityPub
	mux.HandleFunc("POST /inbox", federatedOnly(postInbox))

	// NodeInfo
	mux.HandleFunc("GET /.well-known/nodeinfo", getWellKnownNodeInfo)
	mux.HandleFunc("GET /nodeinfo/2.0", getNodeInfo)

	// WebFinger
	mux.HandleFunc("GET /.well-known/webfinger", federatedOnly(getWebFinger))

	// Static files
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(fs))))
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

	var bytes, mime, err = archiving.NewObeliskArchiver().Fetch(bookmark.URL)
	if err != nil {
		slog.Error("Obelisk failed to fetch an archive of the page",
			"url", bookmark.URL, "err", err)
		handlerNotFound(w, rq)
		return
	}

	artifact, err := types.NewCompressedDocumentArtifact(bytes, mime)
	if err != nil {
		slog.Error("Failed to compress the new archive",
			"url", bookmark.URL, "err", err)
		handlerNotFound(w, rq)
		return
	}

	archiveID, err := db.NewArchivesRepo().Store(int64(bookmark.ID), artifact)
	if err != nil {
		slog.Error("Failed to store the new archive",
			"url", bookmark.URL, "err", err)
		handlerNotFound(w, rq)
		return
	}

	var addr = fmt.Sprintf("/%d?highlight-archive=%d", bookmark.ID, archiveID)
	http.Redirect(w, rq, addr, http.StatusSeeOther)
}

func getRandom(w http.ResponseWriter, rq *http.Request) {
	authed := auth.AuthorizedFromRequest(rq)
	common := emptyCommon()

	bookmarks, totalBookmarks := db.RandomBookmarks(authed, 20)

	templateExec(w, rq, templateFeed, dataFeed{
		Random:               true,
		TotalBookmarks:       totalBookmarks,
		BookmarkGroupsInPage: types.GroupLocalBookmarksByDate(bookmarks),
		SiteDescription:      settings.SiteDescriptionHTML(),
		dataCommon:           common,
	})
}

type dataTimeline struct {
	*dataCommon

	Following            uint
	TotalBookmarks       uint
	BookmarkGroupsInPage []types.RemoteBookmarkGroup
}

func getTimeline(w http.ResponseWriter, rq *http.Request) {
	log.Println("You viewed the Timeline")

	common := emptyCommon()

	currentPage := extractPage(rq)
	bookmarks, total := db.GetRemoteBookmarks(currentPage)
	common.paginator = types.PaginatorFromURL(rq.URL, currentPage, total)

	templateExec(w, rq, templateTimeline, dataTimeline{
		dataCommon:           common,
		TotalBookmarks:       total,
		Following:            db.CountFollowing(),
		BookmarkGroupsInPage: types.GroupRemoteBookmarksByDate(fediverse.RenderRemoteBookmarks(bookmarks)),
	})
}

type dataActorList struct {
	*dataCommon

	Actors []types.Actor
}

func getFollowersWeb(w http.ResponseWriter, rq *http.Request) {
	templateExec(w, rq, templateFollowers, dataActorList{
		dataCommon: emptyCommon(),
		Actors:     db.GetFollowers(),
	})
}

func getFollowingWeb(w http.ResponseWriter, rq *http.Request) {
	templateExec(w, rq, templateFollowing, dataActorList{
		dataCommon: emptyCommon(),
		Actors:     db.GetFollowing(),
	})
}

// postUnfollow is similar to postFollow excepts it's unfollow
func postUnfollow(w http.ResponseWriter, rq *http.Request) {
	var (
		account = rq.FormValue("account")
		next    = rq.FormValue("next")
	)

	if account == "" || next == "" {
		log.Println("/unfollow: required parameters were not passed")
		handlerNotFound(w, rq)
		return
	}

	actor, err := fediverse.RequestActorByNickname(account)
	if err != nil {
		log.Printf("/unfollow: %s\n", err)
		handlerNotFound(w, rq)
		return
	}

	activity, err := activities.NewUndoFollowFromUs(actor.ID)
	if err != nil {
		log.Printf("When creating Undo{Follow} activity: %s\n", err)
		return
	}
	if err = jobs.SendActivityToInbox(activity, actor.Inbox); err != nil {
		log.Printf("When sending activity: %s\n", err)
		return
	}
	db.StopFollowing(actor.ID)
	http.Redirect(w, rq, next, http.StatusSeeOther)
}

// postFollow follows the account specified and redirects next if successful, shows an error if not.
// Both parameters are required.
//
//	/follow?account=@bouncepaw@links.bouncepaw.com&next=/@bouncepaw@links.bouncepaw.com
func postFollow(w http.ResponseWriter, rq *http.Request) {
	var (
		account = rq.FormValue("account")
		next    = rq.FormValue("next")
	)

	if account == "" || next == "" {
		log.Println("/follow: required parameters were not passed")
		handlerNotFound(w, rq)
		return
	}

	actor, err := fediverse.RequestActorByNickname(account)
	if err != nil {
		log.Printf("/follow: %s\n", err)
		handlerNotFound(w, rq)
		return
	}

	activity, err := activities.NewFollowFromUs(actor.ID)
	if err != nil {
		log.Printf("When creating Follow activity: %s\n", err)
		return
	}
	if err = jobs.SendActivityToInbox(activity, actor.Inbox); err != nil {
		log.Printf("When sending activity: %s\n", err)
		return
	}
	db.AddPendingFollowing(actor.ID)
	http.Redirect(w, rq, next, http.StatusSeeOther)
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

		common := emptyCommon()
		common.searchQuery = actor.Acct()
		common.paginator = types.PaginatorFromURL(rq.URL, currentPage, total)
		templateExec(w, rq, templateRemoteProfile, dataAt{
			dataCommon:           common,
			Account:              *actor,
			BookmarkGroupsInPage: types.GroupRemoteBookmarksByDate(fediverse.RenderRemoteBookmarks(bookmarks)),
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

func getWebFinger(w http.ResponseWriter, rq *http.Request) {
	adminUsername := settings.AdminUsername()

	resource := rq.FormValue("resource")
	expected := fmt.Sprintf("acct:%s@%s", adminUsername, types.CleanerLink(settings.SiteURL()))
	if resource != expected {
		log.Printf("WebFinger: Unexpected resource %s\n", resource)
		handlerNotFound(w, rq)
		return
	}
	doc := fmt.Sprintf(`{
  "subject":"%s",
  "links":[
    {
      "rel":"self",
      "type":"application/activity+json",
      "href":"%s/@%s"
    }
  ]
}`, expected, settings.SiteURL(), adminUsername)
	w.Header().Set("Content-Type", "application/ld+json; profile=\"https://www.w3.org/ns/activitystreams\"")
	if _, err := fmt.Fprintf(w, doc); err != nil {
		log.Printf("Error when serving WebFinger: %s\n", err)
	}
}

func handlerActor(w http.ResponseWriter, rq *http.Request) {
	var (
		siteURL       = settings.SiteURL()
		adminUsername = settings.AdminUsername()
	)

	doc, err := json.Marshal(map[string]any{
		"@context":          []string{"https://www.w3.org/ns/activitystreams", "https://w3id.org/security/v1"},
		"type":              "Person",
		"id":                fediverse.OurID(),
		"preferredUsername": adminUsername,
		"name":              settings.SiteName(),
		"inbox":             siteURL + "/inbox",
		"summary":           settings.SiteDescriptionMycomarkup(), // TODO: Think about it
		"publicKey": map[string]string{
			"id":           fediverse.OurID() + "#main-key",
			"owner":        fediverse.OurID(),
			"publicKeyPem": signing.PublicKey(),
		},
		"followers": siteURL + "/followers",
		"following": siteURL + "/following",
		"outbox":    siteURL + "/outbox",
		"url":       fediverse.OurID(),
	})
	if err != nil {
		log.Printf("When marshaling actor activity: %s\n", err)
		handlerNotFound(w, rq)
		return
	}

	w.Header().Set("Content-Type", types.OtherActivityType)
	if _, err := w.Write(doc); err != nil {
		log.Printf("Error when serving Actor: %s\n", err)
	}
}

func getNodeInfo(w http.ResponseWriter, rq *http.Request) {
	// See:
	// => https://github.com/jhass/nodeinfo/blob/main/schemas/2.0/example.json
	// => https://mastodon.social/nodeinfo/2.0
	doc, err := json.Marshal(map[string]any{
		"version": "2.0",
		"software": map[string]string{
			"name":    "betula",
			"version": "1.4.0-rc1",
		},
		"protocols": []string{"activitypub"},
		"services": map[string][]string{
			"inbound":  {},
			"outbound": {"rss2.0"},
		},
		"openRegistrations": false,
		"usage": map[string]any{
			"users": map[string]int{
				"total":          1,
				"activeHalfyear": 1,
				"activeMonth":    1,
			},
			"localPosts":    db.BookmarkCount(false),
			"localComments": 0,
		},
		"metadata": map[string]string{
			"nodeName":        settings.SiteName(),
			"nodeDescription": settings.SiteDescriptionMycomarkup(),
		},
	})
	if err != nil {
		log.Printf("When marshaling /nodeinfo/2.0: %s\n", err)
		return
	}
	w.Header().Set("Content-Type", "application/json; profile=\"http://nodeinfo.diaspora.software/ns/schema/2.0#\"")

	if _, err = w.Write(doc); err != nil {
		log.Printf("Error when serving /nodeinfo/2.0: %s\n", err)
	}
}

func getWellKnownNodeInfo(w http.ResponseWriter, rq *http.Request) {
	// See:
	// => https://github.com/jhass/nodeinfo/blob/main/PROTOCOL.md
	// => https://docs.joinmastodon.org/dev/routes/#nodeinfo
	doc := `{
		"links": [
			{
				"rel": "http://nodeinfo.diaspora.software/ns/schema/2.0",
				"href": "%s/nodeinfo/2.0"
			}
		]
	}`
	w.Header().Set("Content-Type", "application/json")
	if _, err := fmt.Fprintf(w, fmt.Sprintf(doc, settings.SiteURL())); err != nil {
		log.Printf("Error when serving /.well-known/nodeinfo: %s\n", err)
	}
}

func postUnrepost(w http.ResponseWriter, rq *http.Request) {
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
		log.Printf("Trying to unrepost non-existent post no. %d\n", id)
		handlerNotFound(w, rq)
		return
	}
	if bookmark.RepostOf == nil {
		log.Printf("Trying to unrepost a non-repost post no. %d\n", id)
		handlerNotFound(w, rq)
		return
	}

	originalPage := *bookmark.RepostOf
	bookmark.RepostOf = nil
	db.EditBookmark(bookmark)
	http.Redirect(w, rq, fmt.Sprintf("/%d", id), http.StatusSeeOther)

	if settings.FederationEnabled() {
		jobs.ScheduleJSON(jobtype.SendUndoAnnounce, activities.UndoAnnounceReport{
			AnnounceReport: activities.AnnounceReport{
				ReposterUsername: settings.AdminUsername(),
				RepostPage:       fmt.Sprintf("%s/%d", settings.SiteURL(), bookmark.ID),
				OriginalPage:     originalPage,
			},
		})
	}
}

type dataRepostsOf struct {
	*dataCommon

	types.Bookmark
	Reposts []types.RepostInfo
}

func getRepostsOf(w http.ResponseWriter, rq *http.Request) {
	bookmark, ok := extractBookmark(w, rq)
	if !ok {
		return
	}

	reposts, err := db.RepostsOf(bookmark.ID)
	if err != nil {
		// time parsing issues! whatever
		handlerNotFound(w, rq)
		return
	}
	templateExec(w, rq, templateRepostsFor, dataRepostsOf{
		dataCommon: emptyCommon(),
		Bookmark:   *bookmark,
		Reposts:    reposts,
	})

	log.Printf("Show %d reposts for bookmark no. %d\n", len(reposts), bookmark.ID)
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

func repostFormData(rq *http.Request) dataRepost {
	return dataRepost{
		dataCommon: emptyCommon(),
		URL:        rq.FormValue("url"),
		Visibility: types.VisibilityFromString(rq.FormValue("visibility")),
		CopyTags:   rq.FormValue("copy-tags") == "true",
	}
}

func getRepost(w http.ResponseWriter, rq *http.Request) {
	templateExec(w, rq, templateRepost, repostFormData(rq))
}

func postRepost(w http.ResponseWriter, rq *http.Request) {
	formData := repostFormData(rq)
	// Input validation
	if formData.URL == "" {
		formData.ErrorEmptyURL = true
	} else if !stricks.ValidURL(formData.URL) {
		formData.ErrorInvalidURL = true
	} else {
		goto fetchRemoteBookmark
	}
	w.WriteHeader(http.StatusBadRequest)
	templateExec(w, rq, templateRepost, formData)
	return

fetchRemoteBookmark:
	bookmark, err := fediverse.FetchBookmarkAsRepost(formData.URL)
	if errors.Is(err, fediverse.ErrNotBookmark) {
		formData.ErrorImpossible = true
	} else if errors.Is(err, readpage.ErrTimeout) {
		formData.ErrorTimeout = true
	} else if err != nil {
		formData.Err = err
	} else {
		goto reposting
	}
	w.WriteHeader(http.StatusBadRequest)
	templateExec(w, rq, templateRepost, formData)
	return

reposting:
	if !formData.CopyTags {
		bookmark.Tags = nil // 🐸
	}

	id := db.InsertBookmark(*bookmark)

	http.Redirect(w, rq, fmt.Sprintf("/%d", id), http.StatusSeeOther)
	if settings.FederationEnabled() {
		jobs.ScheduleDatum(jobtype.SendAnnounce, id)
	}
}

func postInbox(w http.ResponseWriter, rq *http.Request) {
	data, err := io.ReadAll(io.LimitReader(rq.Body, 32*1000*1000)) // Read no more than 32 KB.
	if err != nil {
		log.Fatalln(err)
	}

	report, err := activities.Guess(data)
	if err != nil {
		log.Printf("Error while parsing incoming activity: %v\n", err)
		return
	}
	if report == nil {
		// Ignored
		return
	}

	switch report := report.(type) {
	case activities.CreateNoteReport:
		if !db.SubscriptionStatus(report.Bookmark.ActorID).WeFollowThem() {
			log.Printf("%s sent us a bookmark %s, but we don't follow them. Contents: %q. Ignoring.\n",
				report.Bookmark.ActorID, report.Bookmark.ID, report.Bookmark.DescriptionHTML)
			return
		}

		log.Printf("%s sent us a bookmark %s. Contents: %q\n",
			report.Bookmark.ActorID, report.Bookmark.ID, report.Bookmark.DescriptionMycomarkup.String)
		db.InsertRemoteBookmark(report.Bookmark)

	case activities.UpdateNoteReport:
		if !db.RemoteBookmarkIsStored(report.Bookmark.ID) {
			// TODO: maybe store them?
			log.Printf("%s updated the bookmark %s, but we don't have it. Contents: %q. Ignoring.\n",
				report.Bookmark.ActorID, report.Bookmark.ID, report.Bookmark.DescriptionHTML)
			return
		}

		log.Printf("%s updated the bookmark %s. Contents: %q\n",
			report.Bookmark.ActorID, report.Bookmark.ID, report.Bookmark.DescriptionMycomarkup.String)
		db.UpdateRemoteBookmark(report.Bookmark)

	case activities.DeleteNoteReport:
		log.Printf("%s deleted the bookmark %s.\n", report.ActorID, report.BookmarkID)
		db.DeleteRemoteBookmark(report.BookmarkID)

	case activities.UndoAnnounceReport:
		log.Printf("%s revoked their repost of %s at %s\n", report.ReposterUsername, report.OriginalPage, report.RepostPage)
		jobs.ScheduleJSON(jobtype.ReceiveUndoAnnounce, report)

	case activities.AnnounceReport:
		log.Printf("%s reposted %s at %s\n", report.ReposterUsername, report.OriginalPage, report.RepostPage)
		jobs.ScheduleJSON(jobtype.ReceiveAnnounce, report)

	case activities.UndoFollowReport:
		// We'll schedule no job because we are making no network request to handle this.
		if report.ObjectID != fediverse.OurID() {
			log.Printf("%s asked to unfollow %s, and that's not us; ignoring.\n", report.ActorID, report.ObjectID)
			return
		}
		if !db.SubscriptionStatus(report.ActorID).TheyFollowUs() {
			log.Printf("%s asked to unfollow us, but they don't follow us; ignoring.\n", report.ActorID)
			return
		}
		log.Printf("%s asked to unfollow us. Thanks for being with us, goodbye!\n", report.ActorID)
		db.RemoveFollower(report.ActorID)

	case activities.FollowReport:
		_, err := fediverse.RequestActorByID(report.ActorID)
		if err != nil {
			log.Printf("Couldn't fetch actor: %s\n", err)
			return
		}
		if signedOK := fediverse.VerifyRequest(rq, data); !signedOK {
			log.Printf("Couldn't verify the signature from %s\n", report.ActorID)
			return
		}

		if report.ObjectID == fediverse.OurID() {
			log.Printf("%s asked to follow us\n", report.ActorID)
			jobs.ScheduleJSON(jobtype.SendAcceptFollow, report)
		} else {
			log.Printf("%s asked to follow %s, which is not us\n", report.ActorID, report.ObjectID)
			jobs.ScheduleJSON(jobtype.ReceiveRejectFollow, report)
		}

	case activities.AcceptReport:
		_, err := fediverse.RequestActorByID(report.ActorID)
		if err != nil {
			log.Printf("Couldn't fetch actor: %s\n", err)
			return
		}
		if signedOK := fediverse.VerifyRequest(rq, data); !signedOK {
			log.Printf("Couldn't verify the signature from %s\n", report.ActorID)
			return
		}

		switch report.Object["type"] {
		case "Follow":
			report := activities.FollowReport{
				ActorID:          stricks.StringifyAnything(report.Object["actor"]),
				ObjectID:         stricks.StringifyAnything(report.Object["object"]),
				OriginalActivity: report.Object,
			}
			jobs.ScheduleJSON(jobtype.ReceiveAcceptFollow, report)
		}

	case activities.RejectReport:
		_, err := fediverse.RequestActorByID(report.ActorID)
		if err != nil {
			log.Printf("Couldn't fetch actor: %s\n", err)
			return
		}
		if signedOK := fediverse.VerifyRequest(rq, data); !signedOK {
			log.Printf("Couldn't verify the signature from %s\n", report.ActorID)
			return
		}

		switch report.Object["type"] {
		case "Follow":
			report := activities.FollowReport{
				ActorID:          stricks.StringifyAnything(report.Object["actor"]),
				ObjectID:         stricks.StringifyAnything(report.Object["object"]),
				OriginalActivity: report.Object,
			}
			jobs.ScheduleJSON(jobtype.ReceiveRejectFollow, report)
		}

	default:
		// Not meant to happen
		log.Printf("Invalid report type")
	}
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

	common := emptyCommon()
	common.paginator = types.PaginatorFromURL(rq.URL, currentPage, totalBookmarks)
	common.searchQuery = query
	log.Printf("Searching ‘%s’. Authorized: %v\n", query, authed)
	templateExec(w, rq, templateSearch, dataSearch{
		dataCommon:           common,
		Query:                query,
		BookmarkGroupsInPage: types.GroupLocalBookmarksByDate(bookmarks),
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
	Bookmarks []types.Bookmark
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
	templateExec(w, rq, templateDay, dataDay{
		dataCommon: emptyCommon(),
		DayStamp:   dayStamp,
		Bookmarks:  db.BookmarksForDay(authed, dayStamp),
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
			SiteURL:                   settings.SiteURL(),
			SiteTitle:                 settings.SiteTitle(),
			SiteDescriptionMycomarkup: settings.SiteDescriptionMycomarkup(),
			CustomCSS:                 settings.CustomCSS(),
			FederationEnabled:         settings.FederationEnabled(),
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

	common := emptyCommon()
	common.searchQuery = "#" + tagName
	common.paginator = types.PaginatorFromURL(rq.URL, currentPage, totalBookmarks)
	templateExec(w, rq, templateTag, dataTag{
		Tag: types.Tag{
			Name:        tagName,
			Description: db.DescriptionForTag(tagName),
		},
		BookmarkGroupsInPage: types.GroupLocalBookmarksByDate(bookmarks),
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

	bookmark.URL = rq.FormValue("url")
	bookmark.Title = rq.FormValue("title")
	bookmark.Visibility = types.VisibilityFromString(rq.FormValue("visibility"))
	bookmark.Description = rq.FormValue("description")
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
}

func getSaveBookmark(w http.ResponseWriter, rq *http.Request) {
	var bookmark types.Bookmark

	bookmark.URL = rq.FormValue("url")
	bookmark.Title = rq.FormValue("title")
	bookmark.Visibility = types.VisibilityFromString(rq.FormValue("visibility"))
	bookmark.Description = rq.FormValue("description")
	bookmark.Tags = types.SplitTags(rq.FormValue("tags"))
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

func getBookmarkFedi(w http.ResponseWriter, rq *http.Request) {
	bookmark, ok := extractBookmark(w, rq)
	if !ok {
		return
	}
	if bookmark.RepostOf != nil {
		// TODO: decide
		log.Printf("Trying to get bookmark object of repost no. %d. Not implemented.\n", bookmark.ID)
		handlerNotFound(w, rq)
		return
	}
	log.Printf("Get bookmark object no. %d\n", bookmark.ID)

	obj, err := activities.NoteFromBookmark(*bookmark)
	if err != nil {
		log.Printf("When making Note object for bookmark: %s\n", err)
		handlerNotFound(w, rq)
	}

	w.Header().Set("Content-Type", types.OtherActivityType)
	if err = json.NewEncoder(w).Encode(obj); err != nil {
		log.Printf("When writing JSON: %s\n", err)
		handlerNotFound(w, rq)
	}
}

type dataBookmark struct {
	Bookmark    types.Bookmark
	RepostCount int

	Archives         []types.Archive
	HighlightArchive int64
	*dataCommon
}

func getBookmarkWeb(w http.ResponseWriter, rq *http.Request) {
	bookmark, ok := extractBookmark(w, rq)
	if !ok {
		return
	}
	log.Printf("Get bookmark page no. %d\n", bookmark.ID)

	archivesRepo := db.NewArchivesRepo()
	archives, err := archivesRepo.FetchForBookmark(int64(bookmark.ID))
	if err != nil {
		slog.Error("Failed to fetch archives for bookmark",
			"bookmarkID", bookmark.ID,
			"err", err)
		// TODO: a better error
		handlerNotFound(w, rq)
		return
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
		var n, err = strconv.Atoi(val)
		if err != nil {
			slog.Warn("Invalid value for highlight-archive",
				"val", val, "err", err)
		} else {
			highlightArchive = int64(n)
		}
	}

	bookmark.Tags = db.TagsForBookmarkByID(bookmark.ID)
	templateExec(w, rq, templatePost, dataBookmark{
		Bookmark:         *bookmark,
		RepostCount:      db.CountRepostsOf(bookmark.ID),
		Archives:         archives,
		HighlightArchive: highlightArchive,
		dataCommon:       common,
	})
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
	common.paginator = types.PaginatorFromURL(rq.URL, currentPage, totalBookmarks)

	templateExec(w, rq, templateFeed, dataFeed{
		TotalBookmarks:       totalBookmarks,
		BookmarkGroupsInPage: types.GroupLocalBookmarksByDate(bookmarks),
		SiteDescription:      settings.SiteDescriptionHTML(),
		dataCommon:           common,
	})
}

func getGo(w http.ResponseWriter, rq *http.Request) {
	bookmark, ok := extractBookmark(w, rq)
	if !ok {
		return
	}
	http.Redirect(w, rq, bookmark.URL, http.StatusSeeOther)
}
