// SPDX-FileCopyrightText: 2023 Danila Gorelko
// SPDX-FileCopyrightText: 2023 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2024 Danila Gorelko
// SPDX-FileCopyrightText: 2024 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2025 Danila Gorelko
// SPDX-FileCopyrightText: 2025 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/url"

	"git.sr.ht/~bouncepaw/betula/db"
	"git.sr.ht/~bouncepaw/betula/fediverse"
	"git.sr.ht/~bouncepaw/betula/fediverse/activities"
	"git.sr.ht/~bouncepaw/betula/fediverse/fedisearch"
	"git.sr.ht/~bouncepaw/betula/fediverse/signing"
	"git.sr.ht/~bouncepaw/betula/jobs"
	"git.sr.ht/~bouncepaw/betula/jobs/jobtype"
	"git.sr.ht/~bouncepaw/betula/pkg/stricks"
	likingports "git.sr.ht/~bouncepaw/betula/ports/liking"
	remarkingports "git.sr.ht/~bouncepaw/betula/ports/remarking"
	"git.sr.ht/~bouncepaw/betula/search"
	"git.sr.ht/~bouncepaw/betula/settings"
	"git.sr.ht/~bouncepaw/betula/types"
)

// Query parameters (all required):
//   - id: id of the bookmark to like; either a number or ActivityPub id.
//   - next: url to redirect to.
func postLike(w http.ResponseWriter, rq *http.Request) {
	var (
		bookmarkID = rq.FormValue("id")
		next       = rq.FormValue("next")
	)

	if bookmarkID == "" || next == "" {
		slog.Error("Empty input for like",
			"id", bookmarkID, "next", next)
		handlerBadRequest(w, rq)
		return
	}

	err := svcLiking.Like(rq.Context(), bookmarkID)
	if err != nil {
		slog.Error("Failed to like", "err", err,
			"id", bookmarkID, "next", next)
		handlerBadRequest(w, rq)
		return
	}

	slog.Info("Liked", "id", bookmarkID, "next", next)
	// Not doing url verification for now. Maybe I should?
	http.Redirect(w, rq, next, http.StatusSeeOther)
}

// Query parameters (all required):
//   - id: id of the bookmark to unlike; either a number or ActivityPub id.
//   - next: url to redirect to.
func postUnlike(w http.ResponseWriter, rq *http.Request) {
	var (
		bookmarkID = rq.FormValue("id")
		next       = rq.FormValue("next")
	)

	if bookmarkID == "" || next == "" {
		slog.Error("Empty input for unlike",
			"id", bookmarkID, "next", next)
		handlerBadRequest(w, rq)
		return
	}

	err := svcLiking.Unlike(rq.Context(), bookmarkID)
	if err != nil {
		slog.Error("Failed to unlike", "err", err,
			"id", bookmarkID, "next", next)
		handlerBadRequest(w, rq)
		return
	}

	slog.Info("Unliked", "id", bookmarkID, "next", next)
	// Not doing url verification for now. Maybe I should?
	http.Redirect(w, rq, next, http.StatusSeeOther)
}

type dataTimeline struct {
	*dataCommon

	Following            uint
	TotalBookmarks       uint
	BookmarkGroupsInPage []types.RemoteBookmarkGroup
}

func getTimeline(w http.ResponseWriter, rq *http.Request) {
	slog.Info("You viewed the Timeline")

	common := emptyCommon()

	currentPage := extractPage(rq)
	bookmarks, total := db.GetRemoteBookmarks(currentPage)
	common.paginator = types.PaginatorFromURL(rq.URL, currentPage, total)

	renderedBookmarks := fediverse.RenderRemoteBookmarks(bookmarks)
	if err := svcLiking.FillLikes(rq.Context(), nil, renderedBookmarks); err != nil {
		slog.Error("Failed to fill likes for remote bookmarks", "err", err)
	}

	templateExec(w, rq, templateTimeline, dataTimeline{
		dataCommon:           common,
		TotalBookmarks:       total,
		Following:            db.CountFollowing(),
		BookmarkGroupsInPage: types.GroupRemoteBookmarksByDate(renderedBookmarks),
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
		// Proceed with unfollowing even if sending failed
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

func postInbox(w http.ResponseWriter, rq *http.Request) {
	data, err := io.ReadAll(io.LimitReader(rq.Body, 32*1000*1000)) // Read no more than 32 KB.
	if err != nil {
		log.Fatalln(err)
	}

	report, err := activities.Guess(data)
	if err != nil {
		slog.Error("Failed to parse incoming activity", "err", err)
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

		if report.LikesCollection != nil {
			event := likingports.EventLikeCollectionSeen{
				ID:            report.LikesCollection.ID,
				Type:          report.LikesCollection.Type,
				TotalItems:    report.LikesCollection.TotalItems,
				LikedObjectID: report.Bookmark.ID,
				SourceJSON:    data,
			}
			err = svcLiking.ReceiveLikeCollection(rq.Context(), event)
			if err != nil {
				slog.Error("Failed to receive like collection", "err", err)
			}
		}

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

		if report.LikesCollection != nil {
			event := likingports.EventLikeCollectionSeen{
				ID:            report.LikesCollection.ID,
				Type:          report.LikesCollection.Type,
				TotalItems:    report.LikesCollection.TotalItems,
				LikedObjectID: report.Bookmark.ID,
			}
			slog.Info("The update contained a likes collection; handling",
				"event", event)
			event.SourceJSON = data // not including in logs

			err = svcLiking.ReceiveLikeCollection(rq.Context(), event)
			if err != nil {
				slog.Error("Failed to receive like collection", "err", err)
			}
		}

	case activities.DeleteNoteReport:
		_, err := fediverse.RequestActorByID(report.ActorID)
		if err != nil {
			log.Printf("Couldn't fetch actor: %s\n", err)
			return
		}
		if signedOK := signing.VerifyRequestSignature(rq, data); !signedOK {
			slog.Error("Failed to verify signature", "actorID", report.ActorID)
			return
		}

		log.Printf("%s deleted the bookmark %s.\n", report.ActorID, report.BookmarkID)
		db.DeleteRemoteBookmark(report.BookmarkID)

	case activities.UndoAnnounceReport:
		_, err := fediverse.RequestActorByID(report.ActorID)
		if err != nil {
			log.Printf("Couldn't fetch actor: %s\n", err)
			return
		}
		if signedOK := signing.VerifyRequestSignature(rq, data); !signedOK {
			slog.Error("Failed to verify signature", "actorID", report.ActorID)
			return
		}

		event := remarkingports.EventLegacyUnremark{
			ActorID:    report.ActorID,
			AnnounceID: report.AnnounceID,
			ObjectID:   report.ObjectID,
		}
		if err = svcRemarking.ReceiveLegacyUnremark(rq.Context(), event); err != nil {
			slog.Error("Failed to receive legacy unremark", "err", err, "event", event)
		}

	case activities.AnnounceReport:
		_, err := fediverse.RequestActorByID(report.ActorID)
		if err != nil {
			log.Printf("Couldn't fetch actor: %s\n", err)
			return
		}
		if signedOK := signing.VerifyRequestSignature(rq, data); !signedOK {
			slog.Error("Failed to verify signature", "actorID", report.ActorID)
			return
		}

		event := remarkingports.EventLegacyRemark{
			ActorID:    report.ActorID,
			AnnounceID: report.AnnounceID,
			ObjectID:   report.ObjectID,
		}
		if err = svcRemarking.ReceiveLegacyRemark(rq.Context(), event); err != nil {
			slog.Error("Failed to receive legacy remark", "err", err, "event", event)
		}

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
		if signedOK := signing.VerifyRequestSignature(rq, data); !signedOK {
			slog.Error("Failed to verify signature", "actorID", report.ActorID)
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
		if signedOK := signing.VerifyRequestSignature(rq, data); !signedOK {
			slog.Error("Failed to verify signature", "actorID", report.ActorID)
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
		if signedOK := signing.VerifyRequestSignature(rq, data); !signedOK {
			slog.Error("Failed to verify signature", "actorID", report.ActorID)
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

	case activities.LikeReport:
		_, err := fediverse.RequestActorByID(report.ActorID)
		if err != nil {
			log.Printf("Couldn't fetch actor: %s\n", err)
			return
		}
		if signedOK := signing.VerifyRequestSignature(rq, data); !signedOK {
			slog.Error("Failed to verify signature", "actorID", report.ActorID)
			return
		}

		event := likingports.EventLike{
			LikeID:        report.ID,
			ActorID:       report.ActorID,
			LikedObjectID: report.ObjectID,
			Activity:      report.Activity,
		}
		err = svcLiking.ReceiveLike(rq.Context(), event)
		if err != nil {
			event.Activity = nil
			slog.Error("Failed to receive Like",
				"event", event, "err", err, "activity", report.Activity)
			return
		}

	case activities.UndoLikeReport:
		_, err := fediverse.RequestActorByID(report.Object.ActorID)
		if err != nil {
			log.Printf("Couldn't fetch actor: %s\n", err)
			return
		}
		if signedOK := signing.VerifyRequestSignature(rq, data); !signedOK {
			slog.Error("Failed to verify signature", "actorID", report.Object.ActorID)
			return
		}

		event := likingports.EventUndoLike{
			UndoLikeID: report.ID,
			ActorID:    report.Object.ActorID,
			LikeID:     report.Object.ID,
			Activity:   report.Activity,
		}
		err = svcLiking.ReceiveUndoLike(rq.Context(), event)
		if err != nil {
			event.Activity = nil
			slog.Error("Failed to receive Undo{Like}",
				"event", event, "err", err, "activity", report.Activity)
			return
		}

	default:
		// Not meant to happen
		slog.Error("Invalid report type; this is a bug")
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
	if _, err := fmt.Fprint(w, doc); err != nil {
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
		"icon":      siteURL + "/static/pix/favicon.png",
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
			"version": "1.6.0",
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
	if _, err := fmt.Fprint(w, fmt.Sprintf(doc, settings.SiteURL())); err != nil {
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
				ActorID:    settings.AdminUsername(),
				AnnounceID: fmt.Sprintf("%s/%d", settings.SiteURL(), bookmark.ID),
				ObjectID:   originalPage,
			},
		})
	}
}

type dataRepost struct {
	*dataCommon

	ErrorInvalidURL,
	ErrorEmptyURL,
	ErrorImpossible,
	ErrorTimeout bool
	Err error

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
	} else if err != nil {
		// Check if it's a timeout error
		var urlErr *url.Error
		if errors.As(err, &urlErr) && urlErr.Timeout() {
			formData.ErrorTimeout = true
		} else {
			formData.Err = err
		}
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
	if settings.FederationEnabled() && formData.Visibility == types.Public {
		jobs.ScheduleDatum(jobtype.SendAnnounce, id)
	}
}

type dataFedisearch struct {
	*dataCommon

	Mutuals []types.Actor

	State     *fedisearch.State // nil for empty fedisearch page
	Bookmarks []types.RenderedRemoteBookmark
}

func handlerFediSearch(w http.ResponseWriter, rq *http.Request) {
	var query = rq.FormValue("query")
	if query == "" {
		slog.Info("Access empty fedisearch page")
		templateExec(w, rq, templateFedisearch, dataFedisearch{
			dataCommon: emptyCommon(),
			Mutuals:    db.GetMutuals(),
		})
		return
	}

	slog.Info("Make a federated search", "query", query)
	var prevState, err = fedisearch.StateFromFormParams(rq.Form, fediverse.OurID())
	if err != nil {
		slog.Error("Failed to parse federated search state",
			"query", query, "err", err)
		handlerNotFound(w, rq) // TODO: proper error page
		return
	}

	renderedBookmarks, nextState, err := prevState.FetchPage()
	if err != nil {
		slog.Error("Failed to fetch federated search bookmarks",
			"query", query, "err", err)
		handlerNotFound(w, rq) // TODO: proper error page
		return
	}

	if err := svcLiking.FillLikes(rq.Context(), nil, renderedBookmarks); err != nil {
		slog.Error("Failed to fill likes for remote bookmarks", "err", err)
	}

	slog.Info("Showing federated search bookmarks",
		"nextState", nextState, "prevState", prevState)
	templateExec(w, rq, templateFedisearch, dataFedisearch{
		dataCommon: emptyCommon(),
		Mutuals:    db.GetMutuals(),
		Bookmarks:  renderedBookmarks,
		State:      nextState,
	})
}

func postFediSearchAPI(w http.ResponseWriter, rq *http.Request) {
	var data, err = io.ReadAll(io.LimitReader(rq.Body, 32*1024*1024))
	if err != nil {
		slog.Error("Failed to read body of fedisearch request", "err", err)
		http.Error(w, "Failed to read request", http.StatusBadRequest)
		return
	}

	if ok := signing.VerifyRequestSignature(rq, data); !ok {
		slog.Warn("Failed to verify signature for fedisearch request")
		http.Error(w, "Failed to verify signature", http.StatusUnauthorized)
		return
	}

	req, err := fedisearch.ParseAPIRequest(data)
	if err != nil {
		slog.Error("Failed to parse fedisearch request", "err", err)
		http.Error(w, "Failed to parse request", http.StatusForbidden)
		return
	}

	var bookmarks, totalResults = search.ForFederated(req.Query, uint(req.Offset), uint(req.Limit))
	var moreAvailable = int(totalResults) - len(bookmarks) - req.Offset
	if moreAvailable < 0 {
		moreAvailable = 0 // just in case
	}
	var bookmarkObjects = make([]activities.Dict, 0, len(bookmarks))

	for _, bookmark := range bookmarks {
		var dict, err = activities.NoteFromBookmark(bookmark)
		if err != nil {
			slog.Error("Failed to make a Note from bookmark", "bookmark", bookmark, "err", err)
			continue
		}
		bookmarkObjects = append(bookmarkObjects, dict)
	}

	response, err := json.Marshal(map[string]any{
		"bookmarks":      bookmarkObjects,
		"more_available": moreAvailable,
	})
	if err != nil {
		slog.Error("Failed to marshal response", "err", err)
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(response)
	if err != nil {
		slog.Error("Failed to write response", "err", err)
	}
}
