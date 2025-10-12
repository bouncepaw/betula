package web

import (
	"net/http/httptest"
	"testing"

	"git.sr.ht/~bouncepaw/betula/db"
	"git.sr.ht/~bouncepaw/betula/fediverse/activities"
	"git.sr.ht/~bouncepaw/betula/fediverse/signing"
	"git.sr.ht/~bouncepaw/betula/settings"
	"git.sr.ht/~bouncepaw/betula/types"
)

func TestUnfollowRemovesFollowingOnSendError(t *testing.T) {
	db.InitInMemoryDB()
	settings.Index()
	signing.EnsureKeysFromDatabase()
	activities.GenerateBetulaActor()

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
	db.StoreValidActor(actor)
	db.AddPendingFollowing(actor.ID)

	rq := httptest.NewRequest("POST", "/unfollow?account=@dan@betula.klava.wiki&next=/", nil)
	rw := httptest.NewRecorder()
	postUnfollow(rw, rq)

	if db.SubscriptionStatus(actor.ID) != types.SubscriptionNone {
		t.Errorf("Unfollow did not remove following when sending Undo failed")
	}
}
