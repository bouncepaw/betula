// SPDX-FileCopyrightText: 2025 Danila Gorelko
// SPDX-FileCopyrightText: 2026 Danila Gorelko
// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package apsvc

import (
	"testing"

	"github.com/nalgeon/be"

	"git.sr.ht/~bouncepaw/betula/db"
	"git.sr.ht/~bouncepaw/betula/fediverse/signing"
	apgw "git.sr.ht/~bouncepaw/betula/gateways/activitypub"
	webfingerports "git.sr.ht/~bouncepaw/betula/ports/webfinger"
	"git.sr.ht/~bouncepaw/betula/settings"
	"git.sr.ht/~bouncepaw/betula/svc/activitypub/assembly"
	"git.sr.ht/~bouncepaw/betula/types"
)

type fakeWebFinger struct {
	ids map[string]string
}

func (f fakeWebFinger) DereferenceAcct(acct webfingerports.Acct) (string, error) {
	return f.ids[acct.String()], nil
}

func TestUnfollowRemovesFollowingOnSendError(t *testing.T) {
	db.InitInMemoryDB()
	settings.Index()
	signing.EnsureKeysFromDatabase()
	asm := assembly.New(settings.SiteURL, settings.AdminUsername)

	actor := types.Actor{
		ID:                "https://betula.klava.wiki/@dan",
		Inbox:             "http://127.0.0.1:0/inbox", // invalid port to force client error
		PreferredUsername: "dan",
		DisplayedName:     "dan",
		Summary:           "",
		Domain:            "betula.klava.wiki",
	}
	actor.PublicKey.ID = actor.ID + "#main-key"
	actor.PublicKey.Owner = actor.ID
	actor.PublicKey.PublicKeyPEM = signing.PublicKey()

	ctx := t.Context()
	repo := db.NewActorRepo()
	be.Err(t, repo.StoreActor(ctx, actor), nil)
	be.Err(t, repo.AddPendingFollowing(ctx, actor.ID), nil)

	webfinger := fakeWebFinger{ids: map[string]string{
		"acct:dan@betula.klava.wiki": actor.ID,
	}}
	activityPub := apgw.NewActivityPub(repo, db.NewRemoteBookmarkRepo())
	svc := NewFollowService(repo, nil, activityPub, webfinger, asm)
	be.Err(t, svc.Unfollow(ctx, "@dan@betula.klava.wiki"), nil)

	status, err := repo.SubscriptionStatus(ctx, actor.ID)
	be.Err(t, err, nil)
	be.Equal(t, status, types.SubscriptionNone)
}
