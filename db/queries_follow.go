package db

import "git.sr.ht/~bouncepaw/betula/types"

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

// Only ID and Acct are set in the actors
func GetFollowing() (actors []types.Actor) {
	rows := mustQuery(`select ActorID, Acct from Following, WebFingerAccts where ActorID = ActorURL`)
	for rows.Next() {
		var actor types.Actor
		mustScan(rows, &actor.ID, &actor.Acct)
		actors = append(actors, actor)
	}
	return
}

func GetFollowers() (actors []types.Actor) {
	rows := mustQuery(`select ActorID, Acct from Followers, WebFingerAccts where ActorID = ActorURL`)
	for rows.Next() {
		var actor types.Actor
		mustScan(rows, &actor.ID, &actor.Acct)
		actors = append(actors, actor)
	}
	return
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
