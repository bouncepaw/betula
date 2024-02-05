package web

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"git.sr.ht/~bouncepaw/betula/fediverse"
	"git.sr.ht/~bouncepaw/betula/fediverse/activities"
	"git.sr.ht/~bouncepaw/betula/fediverse/signing"
	"git.sr.ht/~bouncepaw/betula/jobs"
	"git.sr.ht/~bouncepaw/betula/jobs/jobtype"
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

func init() {
	mux.HandleFunc("/", getIndex)
	mux.HandleFunc("/reposts-of/", getRepostsOf)
	mux.HandleFunc("/help/en/", getEnglishHelp)
	mux.HandleFunc("/help", getHelp)
	mux.HandleFunc("/text/", getText)
	mux.HandleFunc("/digest-rss", getDigestRss)
	mux.HandleFunc("/posts-rss", getPostsRss)
	mux.HandleFunc("/go/", getGo)
	mux.HandleFunc("/about", getAbout)
	mux.HandleFunc("/tag/", getTag)
	mux.HandleFunc("/day/", getDay)
	mux.HandleFunc("/search", getSearch)
	mux.HandleFunc("/static/style.css", getStyle)

	mux.HandleFunc("/register", postOnly(postRegister))
	mux.HandleFunc("/login", handlerLogin)
	mux.HandleFunc("/logout", handlerLogout)
	mux.HandleFunc("/settings", adminOnly(handlerSettings))
	mux.HandleFunc("/bookmarklet", adminOnly(getBookmarklet))

	// Create & Modify
	mux.HandleFunc("/repost", adminOnly(handlerRepost))
	mux.HandleFunc("/unrepost/", adminOnly(postOnly(postUnrepost)))
	mux.HandleFunc("/save-link", adminOnly(handlerSaveBookmark))
	mux.HandleFunc("/edit-link/", adminOnly(handlerEditBookmark))
	mux.HandleFunc("/edit-link-tags/", adminOnly(postOnly(postEditBookmarkTags)))
	mux.HandleFunc("/delete-link/", adminOnly(postOnly(postDeleteBookmark)))
	mux.HandleFunc("/edit-tag/", adminOnly(handlerEditTag))
	mux.HandleFunc("/delete-tag/", adminOnly(postOnly(postDeleteTag)))

	// Federation interface
	mux.HandleFunc("/follow", postOnly(adminOnly(federatedOnly(postFollow))))
	mux.HandleFunc("/unfollow", postOnly(adminOnly(federatedOnly(postUnfollow))))
	mux.HandleFunc("/following", fediverseWebFork(nil, getFollowingWeb))
	mux.HandleFunc("/followers", fediverseWebFork(nil, getFollowersWeb))
	mux.HandleFunc("/timeline", adminOnly(federatedOnly(getTimeline)))

	// ActivityPub
	mux.HandleFunc("/inbox", federatedOnly(postOnly(postInbox)))

	// NodeInfo
	mux.HandleFunc("/.well-known/nodeinfo", getWellKnownNodeInfo)
	mux.HandleFunc("/nodeinfo/2.0", getNodeInfo)

	// WebFinger
	mux.HandleFunc("/.well-known/webfinger", federatedOnly(getWebFinger))

	// Static files
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(fs))))
}

type dataTimeline struct {
	*dataCommon

	Following            uint
	TotalBookmarks       uint
	BookmarkGroupsInPage []types.RemoteBookmarkGroup
}

func getTimeline(w http.ResponseWriter, rq *http.Request) {
	var currentPage uint
	if page, err := strconv.Atoi(rq.FormValue("page")); err != nil || page == 0 {
		currentPage = 1
	} else {
		currentPage = uint(page)
	}

	bookmarks, total := db.GetRemoteBookmarks(currentPage)

	common := emptyCommon()
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

type dataRemoteProfile struct {
	*dataCommon

	Account types.Actor
}

func handlerAt(w http.ResponseWriter, rq *http.Request) {
	/*
		Show profile. Imagine this Betula's author username is goremyka. Then:

			* /@goremyka resolves to their profile. It is public.
			* /@anything is 404 for everyone.
			* /@boris@godun.ov is not available for strangers, but the Boris's profile for the admin.

		This endpoint is available in both HTML and Activity form.

			* The HTML form shows what you expect. Some posts in the future, maybe. Available for both local profile and remote profiles.
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

		common := emptyCommon()
		common.searchQuery = actor.Acct()
		templateExec(w, rq, templateRemoteProfile, dataRemoteProfile{
			dataCommon: common,
			Account:    *actor,
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

type dataSubscribe struct {
	*dataCommon

	// GET
	SiteURL string

	// POST results
	ErrCannotSubscribe bool
	ErrMessage         string
	RequestWasSent     bool
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
			"version": "1.2.0",
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

	post, found := db.PostForID(id)
	if !found {
		log.Printf("Trying to unrepost non-existent post no. %d\n", id)
		handlerNotFound(w, rq)
		return
	}
	if post.RepostOf == nil {
		log.Printf("Trying to unrepost a non-repost post no. %d\n", id)
		handlerNotFound(w, rq)
		return
	}

	originalPage := *post.RepostOf
	post.RepostOf = nil
	db.EditPost(post)
	http.Redirect(w, rq, fmt.Sprintf("/%d", id), http.StatusSeeOther)
	report := activities.UndoAnnounceReport{
		AnnounceReport: activities.AnnounceReport{
			ReposterUsername: settings.AdminUsername(),
			RepostPage:       fmt.Sprintf("%s/%d", settings.SiteURL(), post.ID),
			OriginalPage:     originalPage,
		},
	}
	go jobs.ScheduleJSON(jobtype.SendUndoAnnounce, report)
}

type dataRepostsOf struct {
	*dataCommon

	types.Bookmark
	Reposts []types.RepostInfo
}

func getRepostsOf(w http.ResponseWriter, rq *http.Request) {
	id, ok := extractBookmarkID(w, rq)
	if !ok {
		return
	}

	post, found := db.PostForID(id)
	if !found {
		log.Printf("Did not find post no. %d\n", id)
		handlerNotFound(w, rq)
		return
	}

	authed := auth.AuthorizedFromRequest(rq)
	if post.Visibility == types.Private && !authed {
		log.Printf("Unauthorized attempt to access %s. %d.\n", rq.URL.Path, http.StatusUnauthorized)
		handlerUnauthorized(w, rq)
		return
	}

	reposts, err := db.RepostsOf(post.ID)
	if err != nil {
		// time parsing issues! whatever
		handlerNotFound(w, rq)
		return
	}
	templateExec(w, rq, templateRepostsFor, dataRepostsOf{
		dataCommon: emptyCommon(),
		Bookmark:   post,
		Reposts:    reposts,
	})

	log.Printf("Show %d reposts for post no. %d\n", len(reposts), id)
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

	if rq.Method == http.MethodGet {
		templateExec(w, rq, templateRepost, repost)
		return
	}

	goto good

catchTheFire:
	// All errors end up here.
	w.WriteHeader(http.StatusBadRequest)
	templateExec(w, rq, templateRepost, repost)
	return

good:
	if repost.URL == "" {
		repost.ErrorEmptyURL = true
		goto catchTheFire
	} else if !stricks.ValidURL(repost.URL) {
		repost.ErrorInvalidURL = true
		goto catchTheFire
	}

	foundData, err := readpage.FindDataForMyRepost(repost.URL)

	if errors.Is(err, readpage.ErrTimeout) {
		repost.ErrorTimeout = true
		goto catchTheFire
	} else if err != nil {
		repost.Err = err
		goto catchTheFire
	} else if foundData.IsHFeed || foundData.BookmarkOf == "" || foundData.PostName == "" {
		repost.ErrorImpossible = true
		goto catchTheFire
	}

	post := types.Bookmark{
		URL:         foundData.BookmarkOf,
		Title:       foundData.PostName,
		Description: foundData.Mycomarkup,
		Visibility:  repost.Visibility,
		RepostOf:    &repost.URL,
	}

	if repost.CopyTags {
		post.Tags = types.TagsFromStringSlice(foundData.Tags)
	}

	id := db.AddPost(post)

	go jobs.ScheduleDatum(jobtype.SendAnnounce, id)
	http.Redirect(w, rq, fmt.Sprintf("/%d", id), http.StatusSeeOther)
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
	Query            string
	TotalPosts       uint
	PostGroupsInPage []types.LocalBookmarkGroup
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
	currentPage, err := strconv.Atoi(rq.FormValue("page"))
	if err != nil || currentPage <= 0 {
		currentPage = 1
	}
	posts, totalPosts := search.For(query, authed, uint(currentPage))

	common := emptyCommon()
	common.paginator = types.PaginatorFromURL(rq.URL, uint(currentPage), totalPosts)
	common.searchQuery = query
	log.Printf("Searching ‘%s’. Authorized: %v\n", query, authed)
	templateExec(w, rq, templateSearch, dataSearch{
		dataCommon:       common,
		Query:            query,
		PostGroupsInPage: types.GroupLocalBookmarksByDate(posts),
		TotalPosts:       totalPosts,
	})
}

func getText(w http.ResponseWriter, rq *http.Request) {
	id, ok := extractBookmarkID(w, rq)
	if !ok {
		return
	}

	post, found := db.PostForID(id)
	if !found {
		log.Printf("Did not find post no. %d\n", id)
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

func getPostsRss(w http.ResponseWriter, _ *http.Request) {
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

func getDigestRss(w http.ResponseWriter, _ *http.Request) {
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
	Posts    []types.Bookmark
}

func getDay(w http.ResponseWriter, rq *http.Request) {
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
	templateExec(w, rq, templateDay, dataDay{
		dataCommon: emptyCommon(),
		DayStamp:   dayStamp,
		Posts:      db.PostsForDay(authed, dayStamp),
	})
}

type dataSettings struct {
	types.Settings
	*dataCommon
	ErrBadPort  bool
	FirstRun    bool
	RequestHost string
}

func handlerSettings(w http.ResponseWriter, rq *http.Request) {
	isFirstRun := rq.FormValue("first-run") == "true"
	if rq.Method == http.MethodGet {
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

	post, found := db.PostForID(id)

	if !found {
		log.Println("Trying to delete a non-existent post.")
		handlerNotFound(w, rq)
		return
	}

	db.DeletePost(id)
	http.Redirect(w, rq, "/", http.StatusSeeOther)

	if settings.FederationEnabled() {
		go func(post types.Bookmark) {
			if post.Visibility != types.Public {
				return
			}
			data, err := activities.DeleteNote(post.ID)
			if err != nil {
				log.Printf("When creating Delete{Note} activity for post no. %d: %s\n", id, err)
				return
			}
			jobs.ScheduleDatum(jobtype.SendDeleteNote, data)
		}(post)
	}
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

func handlerLogout(w http.ResponseWriter, rq *http.Request) {
	if rq.Method == http.MethodGet {
		templateExec(w, rq, templateLogoutForm, dataAuthorized{
			dataCommon: emptyCommon(),
		})
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
		templateExec(w, rq, templateLoginForm, dataLogin{
			dataCommon: emptyCommon(),
		})
		return
	}

	var (
		name = rq.FormValue("name")
		pass = rq.FormValue("pass")
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

	auth.LogInResponse(w)
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
		name = rq.FormValue("name")
		pass = rq.FormValue("pass")
	)
	auth.SetCredentials(name, pass)
	auth.LogInResponse(w)
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
	TotalPosts       uint
	PostGroupsInPage []types.LocalBookmarkGroup
}

func getTag(w http.ResponseWriter, rq *http.Request) {
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
	templateExec(w, rq, templateTag, dataTag{
		Tag: types.Tag{
			Name:        tagName,
			Description: db.DescriptionForTag(tagName),
		},
		PostGroupsInPage: types.GroupLocalBookmarksByDate(posts),
		TotalPosts:       totalPosts,
		dataCommon:       common,
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
			// the handler never modifies the post visibility, so we don't care about the past, so we only look at the current value
			post, found := db.PostForID(id)
			if !found {
				log.Printf("When federating bookmark no. %d: bookmark not found\n", post.ID)
				return
			}
			if post.Visibility != types.Public {
				return
			}

			// The post remains public
			data, err := activities.UpdateNote(post)
			if err != nil {
				log.Printf("When creating Update{Note} activity for post no. %d: %s\n", post.ID, err)
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

func handlerEditBookmark(w http.ResponseWriter, rq *http.Request) {
	common := emptyCommon()
	common.head = `<script defer src="/static/autocompletion.js"></script>`

	s := strings.TrimPrefix(rq.URL.Path, "/edit-link/")
	if s == "" {
		http.Redirect(w, rq, "/save-link", http.StatusSeeOther)
		return
	}

	id, ok := extractBookmarkID(w, rq)
	if !ok {
		return
	}

	post, found := db.PostForID(id)
	if !found {
		log.Printf("Trying to edit post no. %d that does not exist. %d.\n", id, http.StatusNotFound)
		handlerNotFound(w, rq)
		return
	}
	oldVisibility := post.Visibility

	if rq.Method == http.MethodGet {
		post.Tags = db.TagsForPost(id)
		templateExec(w, rq, templateEditLink, dataEditLink{
			Bookmark:   post,
			dataCommon: common,
		})
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

			post.CreationTime = time.Now().Format(types.TimeLayout) // It shall match the one generated in DB

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
		}(post, oldVisibility)
	}
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
		templateExec(w, rq, templateEditTag, dataEditTag{
			Tag:        oldTag,
			dataCommon: emptyCommon(),
		})
		return
	}

	var newTag types.Tag
	newName := types.CanonicalTagName(rq.FormValue("new-name"))
	newTag.Name = newName
	newTag.Description = strings.TrimSpace(rq.FormValue("description"))

	merge := rq.FormValue("merge")

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
	types.Bookmark
	Another bool

	// The following three fields can be non-empty, when set through URL parameters or when an erroneous request was made.
	ErrorEmptyURL      bool
	ErrorInvalidURL    bool
	ErrorTitleNotFound bool
}

func handlerSaveBookmark(w http.ResponseWriter, rq *http.Request) {
	var viewData dataSaveLink
	var post types.Bookmark

	common := emptyCommon()
	common.head = `<script defer src="/static/autocompletion.js"></script>`

	if rq.Method == http.MethodGet {
		post.URL = rq.FormValue("url")
		post.Title = rq.FormValue("title")
		post.Visibility = types.VisibilityFromString(rq.FormValue("visibility"))
		post.Description = rq.FormValue("description")
		post.Tags = types.SplitTags(rq.FormValue("tags"))
		// TODO: Document the param behaviour
		templateExec(w, rq, templateSaveLink, dataSaveLink{
			Bookmark:   post,
			dataCommon: common,
		})
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
	post.ID = int(id)

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
		go func(post types.Bookmark) {
			if post.Visibility != types.Public {
				return
			}
			post.CreationTime = time.Now().Format(types.TimeLayout) // It shall match the one generated in DB
			data, err := activities.CreateNote(post)
			if err != nil {
				log.Printf("When creating Create{Note} activity for post no. %d: %s\n", id, err)
				return
			}
			jobs.ScheduleDatum(jobtype.SendCreateNote, data)
		}(post)
	}
}

type dataPost struct {
	Post        types.Bookmark
	RepostCount int
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

	common := emptyCommon()
	common.head = template.HTML(fmt.Sprintf(`<link rel="alternate" type="text/mycomarkup" href="/text/%d">`, id))

	post.Tags = db.TagsForPost(id)
	templateExec(w, rq, templatePost, dataPost{
		Post:        post,
		RepostCount: db.CountRepostsOf(id),
		dataCommon:  common,
	})
}

type dataFeed struct {
	TotalPosts       uint
	PostGroupsInPage []types.LocalBookmarkGroup
	SiteDescription  template.HTML
	*dataCommon
}

var regexpPost = regexp.MustCompile("^/[0-9]+")

func getIndex(w http.ResponseWriter, rq *http.Request) {
	if regexpPost.MatchString(rq.URL.Path) {
		handlerPost(w, rq)
		return
	}
	if strings.HasPrefix(rq.URL.Path, "/@") {
		handlerAt(w, rq)
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

	templateExec(w, rq, templateFeed, dataFeed{
		TotalPosts:       totalPosts,
		PostGroupsInPage: types.GroupLocalBookmarksByDate(posts),
		SiteDescription:  settings.SiteDescriptionHTML(),
		dataCommon:       common,
	})
}

func getGo(w http.ResponseWriter, rq *http.Request) {
	id, ok := extractBookmarkID(w, rq)
	if !ok {
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
