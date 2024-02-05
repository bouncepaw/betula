package db

import (
	"git.sr.ht/~bouncepaw/betula/types"
)

func GetRemoteBookmarks(page uint) (bookmarks []types.RemoteBookmark, total uint) {
	total = querySingleValue[uint](`select count(ID) from RemoteBookmarks`)

	rows := mustQuery(`
select ID, RepostOf, ActorID, Title, DescriptionHTML, DescriptionMycomarkup, PublishedAt, UpdatedAt, URL
from RemoteBookmarks
order by PublishedAt desc
limit ?
offset (? * (? - 1))
`, types.PostsPerPage, types.PostsPerPage, page) // same paging for local bookmarks

	for rows.Next() {
		var b types.RemoteBookmark
		mustScan(rows, &b.ID, &b.RepostOf, &b.ActorID, &b.Title, &b.DescriptionHTML, &b.DescriptionMycomarkup, &b.PublishedAt, &b.UpdatedAt, &b.URL)
		bookmarks = append(bookmarks, b)
	}

	// huh up to 64 additional queries??
	for i, _ := range bookmarks {
		rows = mustQuery(`select Name from RemoteTags where BookmarkID = ?`, bookmarks[i].ID)
		for rows.Next() {
			var tag types.Tag
			mustScan(rows, &tag.Name)
			bookmarks[i].Tags = append(bookmarks[i].Tags, tag)
		}
	}

	return
}

func RemoteBookmarkIsStored(bid string) (isStored bool) {
	rows := mustQuery(`select 1 from RemoteBookmarks where ID = ?`, bid)
	isStored = rows.Next()
	_ = rows.Close()
	return
}

func DeleteRemoteBookmark(bid string) {
	mustExec(`delete from RemoteBookmarks where ID = ?`, bid)
}

func InsertRemoteBookmark(b types.RemoteBookmark) {
	mustExec(`
insert into RemoteBookmarks
    (ID,  RepostOf,   ActorID,   Title,   URL, DescriptionHTML,   DescriptionMycomarkup, PublishedAt,  UpdatedAt, Activity)
values
	(?, ?, ?, ?, ?, ?, ?, ?, null, ?)
on conflict do nothing`,
		b.ID, b.RepostOf, b.ActorID, b.Title, b.URL, b.DescriptionHTML, b.DescriptionMycomarkup, b.PublishedAt, b.Activity)

	for _, tag := range b.Tags {
		mustExec(`insert or replace into RemoteTags (Name, BookmarkID) values (?, ?)`, tag.Name, b.ID)
	}
}

func UpdateRemoteBookmark(b types.RemoteBookmark) {
	// Only own bookmarks can be updated. Ownership can't be changed this way. Publishing date too. The id remains.
	mustExec(`
update RemoteBookmarks
set Title = ?, DescriptionHTML = ?, DescriptionMycomarkup = ?, UpdatedAt = ?, Activity = ?, URL = ?
where ID = ?`,
		b.Title, b.DescriptionHTML, b.DescriptionMycomarkup, b.UpdatedAt, b.Activity, b.ID, b.URL)

	mustExec(`delete from RemoteTags where BookmarkID = ?`, b.ID)

	for _, tag := range b.Tags {
		mustExec(`insert or replace into RemoteTags (Name, BookmarkID) values (?, ?)`, tag, b.ID)
	}
}

func KeyPemByID(keyID string) string {
	return querySingleValue[string](`select PublicKeyPEM from PublicKeys where ID = ?`, keyID)
}

func GetFollowing() (actors []types.Actor) {
	rows := mustQuery(`
select ActorID, PreferredUsername, Inbox, DisplayedName, Summary, Domain, PublicKeyPEM
from Following
join Actors on ActorID = Actors.ID
join PublicKeys on Owner = ActorID;`)
	for rows.Next() {
		var a types.Actor
		mustScan(rows, &a.ID, &a.PreferredUsername, &a.Inbox, &a.DisplayedName, &a.Summary, &a.Domain, &a.PublicKey.PublicKeyPEM)
		actors = append(actors, a)
	}
	return
}

func GetFollowers() (actors []types.Actor) {
	rows := mustQuery(`
select ActorID, PreferredUsername, Inbox, DisplayedName, Summary, Domain, PublicKeyPEM
from Followers
join Actors on ActorID = Actors.ID
join PublicKeys on Owner = ActorID;
`)
	for rows.Next() {
		var a types.Actor
		mustScan(rows, &a.ID, &a.PreferredUsername, &a.Inbox, &a.DisplayedName, &a.Summary, &a.Domain, &a.PublicKey.PublicKeyPEM)
		actors = append(actors, a)
	}
	return
}

func ActorByAcct(user, host string) (a *types.Actor, found bool) {
	rows := mustQuery(`
select Actors.ID, PreferredUsername, Inbox, DisplayedName, Summary, Domain, PublicKeyPEM
from Actors
join PublicKeys on Owner = Actors.ID
where PreferredUsername = ? and Domain = ?
limit 1`, user, host)
	for rows.Next() {
		var actor types.Actor
		mustScan(rows, &actor.ID, &actor.PreferredUsername, &actor.Inbox, &actor.DisplayedName, &actor.Summary, &actor.Domain, &actor.PublicKey.PublicKeyPEM)
		found = true
		a = &actor
	}
	return
}

func ActorByID(actorID string) (a *types.Actor, found bool) {
	rows := mustQuery(`
select Actors.ID, PreferredUsername, Inbox, DisplayedName, Summary, Domain, PublicKeyPEM
from Actors
join PublicKeys on Owner = Actors.ID
where Actors.ID = ?
limit 1`, actorID)
	for rows.Next() {
		var actor types.Actor
		mustScan(rows, &actor.ID, &actor.PreferredUsername, &actor.Inbox, &actor.DisplayedName, &actor.Summary, &actor.Domain, &actor.PublicKey.PublicKeyPEM)
		found = true
		a = &actor
	}
	return
}

func StoreValidActor(a types.Actor) {
	// assume actor.Valid()
	mustExec(`
replace into Actors
    (ID, PreferredUsername, Inbox, DisplayedName, Summary, Domain, LastCheckedAt)
values
	(?, ?, ?, ?, ?, ?, current_timestamp)`,
		a.ID, a.PreferredUsername, a.Inbox, a.DisplayedName, a.Summary, a.Domain)
	mustExec(`
replace into PublicKeys
	(ID, Owner, PublicKeyPEM)
values
	(?, ?, ?)`, a.PublicKey.ID, a.PublicKey.Owner, a.PublicKey.PublicKeyPEM)
}

func AddFollower(id string) {
	mustExec(`replace into Followers (ActorID) values (?)`, id)
}

func RemoveFollower(id string) {
	mustExec(`delete from Followers where ActorID = ?`, id)
}

func AddPendingFollowing(id string) {
	mustExec(`replace into Following (ActorID) values (?)`, id)
}

func MarkAsSurelyFollowing(id string) {
	mustExec(`update Following set AcceptedStatus = 1 where ActorID = ?`, id)
}

func StopFollowing(id string) {
	mustExec(`delete from Following where ActorID = ?`, id)
}

func CountFollowing() uint {
	return querySingleValue[uint](`select count(*) from Following;`)
}
func CountFollowers() uint {
	return querySingleValue[uint](`select count(*) from Followers;`)
}

func SubscriptionStatus(id string) types.SubscriptionRelation {
	// TODO: make it just 1 request.
	var iFollow, theyFollow bool
	var status int

	rows := mustQuery(`select AcceptedStatus from Following where ActorID = ?`, id)
	for rows.Next() {
		iFollow = true
		mustScan(rows, &status)
	}

	rows = mustQuery(`select 1 from Followers where ActorID = ?`, id)
	theyFollow = rows.Next()
	_ = rows.Close()

	pending := status == 0

	switch {
	case pending && iFollow && theyFollow:
		return types.SubscriptionPendingMutual
	case pending && iFollow:
		return types.SubscriptionPending
	case iFollow && theyFollow:
		return types.SubscriptionMutual
	case iFollow:
		return types.SubscriptionIFollow
	case theyFollow:
		return types.SubscriptionTheyFollow
	default:
		return types.SubscriptionNone
	}
}
