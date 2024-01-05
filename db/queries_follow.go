package db

import "git.sr.ht/~bouncepaw/betula/types"

func AddFollower(id string) {
	mustExec(`insert into Followers (ActorID) values (?)`, id)
}

func AddFollowing(id string) {
	mustExec(`insert into Following (ActorID) values (?)`, id)
}

func SubscriptionStatus(id string) types.SubscriptionRelation {
	// TODO: make it just 1 request.
	var iFollow, theyFollow bool

	rows := mustQuery(`select 1 from Following where ActorID = ?`, id)
	iFollow = rows.Next()
	_ = rows.Close()

	rows = mustQuery(`select 1 from Followers where ActorID = ?`, id)
	theyFollow = rows.Next()
	_ = rows.Close()

	switch {
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
