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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"git.sr.ht/~bouncepaw/betula/fediverse"
	"git.sr.ht/~bouncepaw/betula/fediverse/fedisearch"
	"git.sr.ht/~bouncepaw/betula/fediverse/signing"
	"git.sr.ht/~bouncepaw/betula/pkg/bxstr"
	apports "git.sr.ht/~bouncepaw/betula/ports/activitypub"
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

	err := ctrl.SvcLiking.Like(rq.Context(), bookmarkID)
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

	err := ctrl.SvcLiking.Unlike(rq.Context(), bookmarkID)
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
	bookmarks, total := ctrl.RepoRemoteBookmark.GetRemoteBookmarks(currentPage)
	common.paginator = types.PaginatorFromURL(rq.URL, currentPage, total)

	renderedBookmarks, _ := ctrl.SvcRemoteBookmarks.Render(rq.Context(), bookmarks)
	if err := ctrl.SvcLiking.FillLikes(rq.Context(), nil, renderedBookmarks); err != nil {
		slog.Error("Failed to fill likes for remote bookmarks", "err", err)
	}

	followingCount, err := ctrl.RepoActor.CountFollowing(rq.Context())
	if err != nil {
		slog.Error("Failed to count following", "err", err)
		handlerBadRequest(w, rq)
		return
	}

	templateExec(w, rq, templateTimeline, dataTimeline{
		dataCommon:           common,
		TotalBookmarks:       total,
		Following:            followingCount,
		BookmarkGroupsInPage: types.GroupRemoteBookmarksByDate(renderedBookmarks),
	})
}

func getBookmarkFedi(w http.ResponseWriter, rq *http.Request) {
	bookmark, ok := extractBookmark(w, rq)
	if !ok {
		return
	}
	if bookmark.RemarkOf != nil {
		// TODO: decide
		slog.Warn("Get bookmark object of remark not implemented", "bookmarkID", bookmark.ID)
		handlerNotFound(w, rq)
		return
	}
	slog.Info("Get bookmark object", "bookmarkID", bookmark.ID)

	obj, err := ctrl.Assembly.NoteFromBookmark(*bookmark)
	if err != nil {
		slog.Error("Failed to make Note object for bookmark", "err", err)
		handlerNotFound(w, rq)
	}

	w.Header().Set("Content-Type", types.OtherActivityType)
	if err = json.NewEncoder(w).Encode(obj); err != nil {
		slog.Error("Failed to write JSON", "err", err)
		handlerNotFound(w, rq)
	}
}

func getWebFinger(w http.ResponseWriter, rq *http.Request) {
	adminUsername := settings.AdminUsername()

	resource := rq.FormValue("resource")
	expected := fmt.Sprintf("acct:%s@%s", adminUsername, types.CleanerLink(settings.SiteURL()))
	if resource != expected {
		slog.Info("WebFinger: unexpected resource", "resource", resource)
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
		slog.Error("Error when serving WebFinger", "err", err)
	}
}

func getLocalActorObject(w http.ResponseWriter, rq *http.Request) {
	var (
		siteURL       = settings.SiteURL()
		adminUsername = settings.AdminUsername()

		b   bytes.Buffer
		enc = json.NewEncoder(&b)
	)

	enc.SetIndent("", "\t")
	err := enc.Encode(map[string]any{
		"@context": []string{
			"https://www.w3.org/ns/activitystreams",
			"https://w3id.org/security/v1",
			"https://purl.archive.org/socialweb/webfinger", // FEP-2c59
		},
		"type":              "Person",
		"id":                fediverse.OurID(),
		"preferredUsername": adminUsername,
		"name":              settings.SiteName(),
		"inbox":             siteURL + "/inbox",
		"summary":           settings.SiteDescriptionHTML(),
		"publicKey": map[string]string{
			"id":           fediverse.OurID() + "#main-key",
			"owner":        fediverse.OurID(),
			"publicKeyPem": signing.PublicKey(),
		},
		"followers": siteURL + "/followers",
		"following": siteURL + "/following",
		"outbox":    siteURL + "/outbox",
		"url":       fediverse.OurID(),
		"icon": map[string]string{
			"type":      "Image",
			"mediaType": "image/png",
			"url":       siteURL + "/static/pix/favicon.png",
		},
		"webfinger": fmt.Sprintf("acct:%s@%s", adminUsername, settings.SiteDomain()),
	})
	if err != nil {
		slog.Error("Failed to marshal actor activity", "err", err)
		handlerNotFound(w, rq)
		return
	}

	w.Header().Set("Content-Type", types.OtherActivityType)
	if _, err := w.Write(b.Bytes()); err != nil {
		slog.Error("Failed to serve Actor", "err", err)
	}
}

func getNodeInfo(w http.ResponseWriter, rq *http.Request) {
	// See:
	// => https://github.com/jhass/nodeinfo/blob/main/schemas/2.0/example.json
	// => https://mastodon.social/nodeinfo/2.0
	bookmarkCount, err := localBookmarks.BookmarkCount(rq.Context(), false)
	if err != nil {
		slog.Error("Failed to count bookmarks for /nodeinfo/2.0", "err", err)
		http.Error(w, "Failed to gather node info", http.StatusInternalServerError)
		return
	}
	doc, err := json.Marshal(map[string]any{
		"version": "2.0",
		"software": map[string]string{
			"name":    "betula",
			"version": "1.8.1",
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
			"localBookmarks":    bookmarkCount,
			"localComments": 0,
		},
		"metadata": map[string]string{
			"nodeName":        settings.SiteName(),
			"nodeDescription": settings.SiteDescriptionMycomarkup(),
		},
	})
	if err != nil {
		slog.Error("Failed to marshal /nodeinfo/2.0", "err", err)
		return
	}
	w.Header().Set("Content-Type", "application/json; profile=\"http://nodeinfo.diaspora.software/ns/schema/2.0#\"")

	if _, err = w.Write(doc); err != nil {
		slog.Error("Error when serving /nodeinfo/2.0", "err", err)
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
		slog.Error("Error when serving /.well-known/nodeinfo", "err", err)
	}
}

type dataRemark struct {
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

func remarkFormData(rq *http.Request) dataRemark {
	return dataRemark{
		dataCommon: emptyCommon(),
		URL:        rq.FormValue("url"),
		Visibility: types.VisibilityFromString(rq.FormValue("visibility")),
		CopyTags:   rq.FormValue("copy-tags") == "true",
	}
}

func getRemark(w http.ResponseWriter, rq *http.Request) {
	templateExec(w, rq, templateRemark, remarkFormData(rq))
}

func postRemark(w http.ResponseWriter, rq *http.Request) {
	formData := remarkFormData(rq)
	// Input validation
	if formData.URL == "" {
		formData.ErrorEmptyURL = true
	} else if !bxstr.IsValidURL(formData.URL) {
		formData.ErrorInvalidURL = true
	} else {
		goto fetchRemoteBookmark
	}
	w.WriteHeader(http.StatusBadRequest)
	templateExec(w, rq, templateRemark, formData)
	return

fetchRemoteBookmark:
	bookmark, err := fediverse.FetchBookmarkAsRemark(formData.URL)
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
		goto remarking
	}
	w.WriteHeader(http.StatusBadRequest)
	templateExec(w, rq, templateRemark, formData)
	return

remarking:
	if !formData.CopyTags {
		bookmark.Tags = nil // 🐸
	}

	bookmark.CreationTime = time.Now().UTC().Format(types.TimeLayout)
	id, err := localBookmarks.InsertBookmark(rq.Context(), *bookmark)
	if err != nil {
		slog.Error("Failed to insert remark bookmark", "err", err)
		http.Error(w, "Failed to save remark", http.StatusInternalServerError)
		return
	}
	bookmark.ID = int(id)

	if settings.FederationEnabled() && formData.Visibility == types.Public {
		err = ctrl.SvcRemarking.BroadcastCreateRemark(rq.Context(), *bookmark)
		if err != nil {
			slog.Error("Failed to broadcast remark", "err", err, "url", formData.URL, "id", id)
			http.Error(w, "Failed to broadcast remark", http.StatusInternalServerError)
			return
		}
	}

	http.Redirect(w, rq, fmt.Sprintf("/%d", id), http.StatusSeeOther)
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
		mutuals, err := ctrl.RepoActor.GetMutuals(rq.Context())
		if err != nil {
			slog.Error("Failed to get mutuals", "err", err)
			handlerNotFound(w, rq)
			return
		}
		templateExec(w, rq, templateFedisearch, dataFedisearch{
			dataCommon: emptyCommon(),
			Mutuals:    mutuals,
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

	renderedBookmarks, nextState, err := prevState.FetchPage(rq.Context(), ctrl.SvcRemoteBookmarks)
	if err != nil {
		slog.Error("Failed to fetch federated search bookmarks",
			"query", query, "err", err)
		handlerNotFound(w, rq) // TODO: proper error page
		return
	}

	if err := ctrl.SvcLiking.FillLikes(rq.Context(), nil, renderedBookmarks); err != nil {
		slog.Error("Failed to fill likes for remote bookmarks", "err", err)
	}

	mutuals, err := ctrl.RepoActor.GetMutuals(rq.Context())
	if err != nil {
		slog.Error("Failed to get mutuals", "err", err)
		handlerNotFound(w, rq)
		return
	}

	slog.Info("Showing federated search bookmarks",
		"nextState", nextState, "prevState", prevState)
	templateExec(w, rq, templateFedisearch, dataFedisearch{
		dataCommon: emptyCommon(),
		Mutuals:    mutuals,
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

	var bookmarks, totalResults = ctrl.SvcSearching.ForFederated(req.Query, uint(req.Offset), uint(req.Limit))
	var moreAvailable = int(totalResults) - len(bookmarks) - req.Offset
	if moreAvailable < 0 {
		moreAvailable = 0 // just in case
	}
	var bookmarkObjects = make([]apports.Dict, 0, len(bookmarks))

	for _, bookmark := range bookmarks {
		var dict, err = ctrl.Assembly.NoteFromBookmark(bookmark)
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

func postRefetchActors(w http.ResponseWriter, rq *http.Request) {
	err := ctrl.ActivityPub.RefetchAllActors(rq.Context())
	if err != nil {
		slog.Error("Failed to refetch all actors", "err", err)
		http.Error(
			w,
			fmt.Sprintf("Failed to refetch all actors: %s", err.Error()),
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	_, err = w.Write([]byte("OK"))
}
