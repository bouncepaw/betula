// SPDX-FileCopyrightText: 2023 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2024 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2026 Danila Gorelko
// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package types

import (
	"database/sql"
	"fmt"
	"html/template"
	"strings"
	"time"

	"git.sr.ht/~bouncepaw/betula/pkg/bxstr"
)

func ContainsActivityType(s string) bool {
	return strings.Contains(s, ActivityType) || strings.Contains(s, OtherActivityType)
}

const (
	ActivityType      = "application/ld+json; profile=\"https://www.w3.org/ns/activitystreams\""
	OtherActivityType = "application/activity+json"
)

type Actor struct {
	ID                string `json:"id"`
	Inbox             string `json:"inbox"`
	PreferredUsername string `json:"preferredUsername"`
	DisplayedName     string `json:"name"`
	Summary           string `json:"summary,omitempty"`
	PublicKey         struct {
		ID           string `json:"id"`
		Owner        string `json:"owner"`
		PublicKeyPEM string `json:"publicKeyPem"`
	} `json:"publicKey"`

	SubscriptionStatus SubscriptionRelation `json:"-"` // Set manually
	Domain             string               `json:"-"` // Set manually
}

func (a Actor) Valid() bool {
	urlsOK := bxstr.IsValidURL(a.ID) && bxstr.IsValidURL(a.Inbox) && bxstr.IsValidURL(a.PublicKey.Owner)
	nonEmpty := a.PreferredUsername != "" && a.PublicKey.PublicKeyPEM != "" && a.Domain != ""
	return urlsOK && nonEmpty
}

func (a Actor) Acct() string {
	return fmt.Sprintf("@%s@%s", a.PreferredUsername, a.Domain)
}

type SubscriptionRelation string

const (
	SubscriptionNone          SubscriptionRelation = ""
	SubscriptionTheyFollow    SubscriptionRelation = "follower"
	SubscriptionIFollow       SubscriptionRelation = "following"
	SubscriptionMutual        SubscriptionRelation = "mutual"
	SubscriptionPending       SubscriptionRelation = "pending"
	SubscriptionPendingMutual SubscriptionRelation = "pending mutual" // yours pending, theirs accepted
)

func (sr SubscriptionRelation) IsPending() bool {
	return sr == SubscriptionPending || sr == SubscriptionPendingMutual
}

func (sr SubscriptionRelation) TheyFollowUs() bool {
	return sr == SubscriptionTheyFollow || sr == SubscriptionMutual || sr == SubscriptionPendingMutual
}

func (sr SubscriptionRelation) WeFollowThem() bool {
	// TODO: if our request is pending, but we receive a post from them, does it mean they accepted?
	return sr == SubscriptionIFollow || sr == SubscriptionMutual || sr == SubscriptionPendingMutual || sr == SubscriptionPending
}

type SourceType string

const (
	SourceMycomarkup SourceType = "text/mycomarkup"
	SourcePlainText  SourceType = "text/plain"
)

func SourceTypeFromDB(v sql.NullString) SourceType {
	if v.Valid && v.String == "P" {
		return SourcePlainText
	}
	return SourceMycomarkup
}

func (st SourceType) ToDB() sql.NullString {
	if st == SourcePlainText {
		return sql.NullString{String: "P", Valid: true}
	}
	return sql.NullString{}
}

type RemoteBookmark struct {
	ID         string
	RemarkedID sql.NullString
	ActorID    string

	Title           string
	URL             string
	WebURL          sql.NullString
	DescriptionHTML template.HTML
	Source          sql.NullString
	SourceType      SourceType
	PublishedAt     string
	UpdatedAt       sql.NullString
	Activity        []byte

	Tags []Tag
}

func (rb RemoteBookmark) RepresentationURL() string {
	if rb.WebURL.Valid {
		return rb.WebURL.String
	}
	return rb.ID
}

func (rb RemoteBookmark) IsValidObject() bool {
	return rb.ID != "" && rb.ActorID != "" && rb.PublishedAt != ""
}

func (rb RemoteBookmark) IsRegularBookmark() bool {
	return rb.IsValidObject() && rb.Title != "" && rb.URL != "" && !rb.RemarkedID.Valid
}

func (rb RemoteBookmark) IsRemark() bool {
	return rb.IsValidObject() && rb.Title == "" && rb.URL == "" && rb.RemarkedID.Valid
}

type RenderedRemoteBookmark struct {
	ID string

	AuthorAcct          string
	AuthorDisplayedName string
	RemarkedID          sql.NullString

	// For a remark, the author of the original, remarked bookmark. Empty for
	// plain bookmarks and when the original author's actor is unknown.
	OriginalAuthorAcct          string
	OriginalAuthorDisplayedName string
	OriginalWebURL              string

	Title       string
	URL         string
	WebURL      string
	Description template.HTML
	Tags        []Tag
	PublishedAt time.Time

	LikedByUs   bool
	LikeCounter int
}

type RemoteBookmarkGroup struct {
	Date      string
	Bookmarks []RenderedRemoteBookmark
}

var remoteCutoff RenderedRemoteBookmark = (func() RenderedRemoteBookmark {
	bigtime := "9999-01-02T15:04:05+07:00"
	t, err := time.Parse(time.RFC3339, bigtime)
	if err != nil {
		panic(err)
	}
	return RenderedRemoteBookmark{PublishedAt: t}
})()

// GroupRemoteBookmarksByDate groups the bookmarks by date. The dates are strings like 2006-01-02T15:04:05Z07:00 (ActivityPub-style). This function expects the input bookmarks to be sorted by date.
func GroupRemoteBookmarksByDate(ungroupedBookmarks []RenderedRemoteBookmark) (groupedBookmarks []RemoteBookmarkGroup) {
	if len(ungroupedBookmarks) == 0 {
		return nil
	}

	ungroupedBookmarks = append(ungroupedBookmarks, remoteCutoff)

	var (
		currentDate      string
		currentBookmarks []RenderedRemoteBookmark
	)

	for _, bookmark := range ungroupedBookmarks {
		if bookmark.PublishedAt.Format(time.DateOnly) != currentDate {
			if currentBookmarks != nil {
				groupedBookmarks = append(groupedBookmarks, RemoteBookmarkGroup{
					Date:      currentDate,
					Bookmarks: currentBookmarks,
				})
			}
			currentDate = bookmark.PublishedAt.Format(time.DateOnly)
			currentBookmarks = nil
		}

		currentBookmarks = append(currentBookmarks, bookmark)
	}

	return
}
