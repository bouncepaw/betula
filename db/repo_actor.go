// SPDX-FileCopyrightText: 2024 Danila Gorelko
// SPDX-FileCopyrightText: 2024 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2025 Danila Gorelko
// SPDX-FileCopyrightText: 2025 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package db

import (
	"context"
	"errors"
	"fmt"
	apports "git.sr.ht/~bouncepaw/betula/ports/activitypub"
	"git.sr.ht/~bouncepaw/betula/types"
)

type ActorRepo struct{}

var _ apports.ActorRepository = &ActorRepo{}

func NewActorRepo() *ActorRepo {
	return &ActorRepo{}
}

func (repo *ActorRepo) GetActorByID(
	ctx context.Context,
	id string,
	opts apports.GetActorsOpts,
) (types.Actor, error) {
	if opts.GetPublicKey {
		return repo.getActorByIDWithKey(ctx, id)
	}
	return repo.getActorByIDWithoutKey(ctx, id)
}

func (repo *ActorRepo) getActorByIDWithKey(
	ctx context.Context,
	id string,
) (types.Actor, error) {
	row := db.QueryRowContext(ctx, `
select Actors.ID, PreferredUsername, Inbox, DisplayedName, Summary, Domain, PublicKeyPEM
from Actors
join PublicKeys on Owner = Actors.ID
where Actors.ID = ?
limit 1`, id)

	var actor types.Actor
	err := row.Scan(&actor.ID, &actor.PreferredUsername, &actor.Inbox, &actor.DisplayedName, &actor.Summary, &actor.Domain, &actor.PublicKey.PublicKeyPEM)
	return actor, err
}

func (repo *ActorRepo) getActorByIDWithoutKey(
	ctx context.Context,
	id string,
) (types.Actor, error) {
	row := db.QueryRowContext(ctx, `
select Actors.ID, PreferredUsername, Inbox, DisplayedName, Summary, Domain
from Actors
where Actors.ID = ?
limit 1`, id)

	var actor types.Actor
	err := row.Scan(&actor.ID, &actor.PreferredUsername, &actor.Inbox, &actor.DisplayedName, &actor.Summary, &actor.Domain)
	return actor, err
}

func (repo *ActorRepo) AllActorIDs(ctx context.Context) ([]string, error) {
	rows, err := db.QueryContext(ctx, `select ID from Actors`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var actorIDs []string
	for rows.Next() {
		var actorID string
		if err = rows.Scan(&actorID); err != nil {
			return nil, err
		}
		actorIDs = append(actorIDs, actorID)
	}
	return actorIDs, nil
}

func (repo *ActorRepo) StoreActor(ctx context.Context, a types.Actor) error {
	if !a.Valid() {
		return fmt.Errorf("invalid actor")
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
replace into Actors
    (ID, PreferredUsername, Inbox, DisplayedName, Summary, Domain, LastCheckedAt)
values
	(?, ?, ?, ?, ?, ?, current_timestamp)`,
		a.ID, a.PreferredUsername, a.Inbox, a.DisplayedName, a.Summary, a.Domain)
	if err != nil {
		return errors.Join(err, tx.Rollback())
	}

	_, err = tx.ExecContext(ctx, `
replace into PublicKeys
	(ID, Owner, PublicKeyPEM)
values
	(?, ?, ?)`, a.PublicKey.ID, a.PublicKey.Owner, a.PublicKey.PublicKeyPEM)
	if err != nil {
		return errors.Join(err, tx.Rollback())
	}

	return tx.Commit()
}

func (repo *ActorRepo) GetFollowers(ctx context.Context) ([]types.Actor, error) {
	rows, err := db.QueryContext(ctx, `
		select ActorID, PreferredUsername, Inbox, DisplayedName, Summary, Domain, PublicKeyPEM
		from Followers
		join Actors on ActorID = Actors.ID
		join PublicKeys on Owner = ActorID
`)
	if err != nil {
		return nil, err
	}

	var actors []types.Actor
	for rows.Next() {
		var a types.Actor
		err = rows.Scan(&a.ID, &a.PreferredUsername, &a.Inbox, &a.DisplayedName, &a.Summary, &a.Domain, &a.PublicKey.PublicKeyPEM)
		if err != nil {
			return nil, err
		}
		actors = append(actors, a)
	}
	return actors, nil
}

// Below: old code, to be refactored to modern style.

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

// GetFollowers
//
// Deprecated: Use (*ActorRepo).GetFollowers instead.
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

func GetMutuals() (actors []types.Actor) {
	rows := mustQuery(`
select Following.ActorID, PreferredUsername, Inbox, DisplayedName, Summary, Domain, PublicKeyPEM
from Following
join Actors on Following.ActorID = Actors.ID
join PublicKeys on Owner = Following.ActorID
where Following.ActorID in (
	select Followers.ActorID from Followers
);`)
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

// ActorByID
//
// Deprecated: Use (*ActorRepo).GetActorByID instead.
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

// StoreValidActor
//
// Deprecated: Use (*ActorRepo).StoreActor instead.
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
