package types

import (
	"database/sql"
	"fmt"
	"html/template"
	"time"

	"git.sr.ht/~bouncepaw/betula/stricks"
)

const ActivityType = "application/ld+json; profile=\"https://www.w3.org/ns/activitystreams\""
const OtherActivityType = "application/activity+json"

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
	} `json:"publicKey,omitempty"`

	SubscriptionStatus SubscriptionRelation `json:"-"` // Set manually
	Domain             string               `json:"-"` // Set manually
}

func (a *Actor) Valid() bool {
	urlsOK := stricks.ValidURL(a.ID) && stricks.ValidURL(a.Inbox) && stricks.ValidURL(a.PublicKey.ID) && stricks.ValidURL(a.PublicKey.Owner)
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

type RemoteBookmark struct {
	ID       string
	RepostOf sql.NullString
	ActorID  string

	Title                 string
	URL                   string
	DescriptionHTML       template.HTML
	DescriptionMycomarkup sql.NullString
	PublishedAt           string
	UpdatedAt             sql.NullString
	Activity              []byte

	Tags []Tag
}

type RenderedRemoteBookmark struct {
	ID string

	AuthorAcct          string
	AuthorDisplayedName string
	RepostOf            sql.NullString

	Title       string
	URL         string
	Description template.HTML
	Tags        []Tag
	PublishedAt time.Time
}

type RemoteBookmarkGroup struct {
	Date  string
	Posts []RenderedRemoteBookmark
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
		currentDate      time.Time
		currentBookmarks []RenderedRemoteBookmark
	)

	for _, bookmark := range ungroupedBookmarks {
		if bookmark.PublishedAt != currentDate {
			if currentBookmarks != nil {
				groupedBookmarks = append(groupedBookmarks, RemoteBookmarkGroup{
					Date:  currentDate.Format(TimeLayout),
					Posts: currentBookmarks,
				})
			}
			currentDate = bookmark.PublishedAt
			currentBookmarks = nil
		}

		currentBookmarks = append(currentBookmarks, bookmark)
	}

	return
}
