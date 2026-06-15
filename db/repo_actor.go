// SPDX-FileCopyrightText: 2024 Danila Gorelko
// SPDX-FileCopyrightText: 2024 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2025 Danila Gorelko
// SPDX-FileCopyrightText: 2025 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2026 Danila Gorelko
// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package db

import (
	"context"
	"database/sql"
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
	return scanActorsWithKey(rows)
}

func (repo *ActorRepo) GetFollowing(ctx context.Context) ([]types.Actor, error) {
	rows, err := db.QueryContext(ctx, `
select ActorID, PreferredUsername, Inbox, DisplayedName, Summary, Domain, PublicKeyPEM
from Following
join Actors on ActorID = Actors.ID
join PublicKeys on Owner = ActorID;`)
	if err != nil {
		return nil, err
	}
	return scanActorsWithKey(rows)
}

func (repo *ActorRepo) GetMutuals(ctx context.Context) ([]types.Actor, error) {
	rows, err := db.QueryContext(ctx, `
select Following.ActorID, PreferredUsername, Inbox, DisplayedName, Summary, Domain, PublicKeyPEM
from Following
join Actors on Following.ActorID = Actors.ID
join PublicKeys on Owner = Following.ActorID
where Following.ActorID in (
	select Followers.ActorID from Followers
);`)
	if err != nil {
		return nil, err
	}
	return scanActorsWithKey(rows)
}

// ActorByAcct returns the cached actor with the given handle. It returns
// sql.ErrNoRows when no such actor is known.
func (repo *ActorRepo) ActorByAcct(ctx context.Context, user, host string) (types.Actor, error) {
	var actor types.Actor
	err := db.QueryRowContext(ctx, `
select Actors.ID, PreferredUsername, Inbox, DisplayedName, Summary, Domain, PublicKeyPEM
from Actors
join PublicKeys on Owner = Actors.ID
where PreferredUsername = ? and Domain = ?
limit 1`, user, host).Scan(
		&actor.ID, &actor.PreferredUsername, &actor.Inbox, &actor.DisplayedName, &actor.Summary, &actor.Domain, &actor.PublicKey.PublicKeyPEM)
	return actor, err
}

func (repo *ActorRepo) KeyPemByID(ctx context.Context, keyID string) (string, error) {
	var pem string
	err := db.QueryRowContext(ctx, `select PublicKeyPEM from PublicKeys where ID = ?`, keyID).Scan(&pem)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return pem, err
}

func (repo *ActorRepo) AddFollower(ctx context.Context, id string) error {
	_, err := db.ExecContext(ctx, `replace into Followers (ActorID) values (?)`, id)
	return err
}

func (repo *ActorRepo) RemoveFollower(ctx context.Context, id string) error {
	_, err := db.ExecContext(ctx, `delete from Followers where ActorID = ?`, id)
	return err
}

func (repo *ActorRepo) AddPendingFollowing(ctx context.Context, id string) error {
	_, err := db.ExecContext(ctx, `replace into Following (ActorID) values (?)`, id)
	return err
}

func (repo *ActorRepo) MarkAsSurelyFollowing(ctx context.Context, id string) error {
	_, err := db.ExecContext(ctx, `update Following set AcceptedStatus = 1 where ActorID = ?`, id)
	return err
}

func (repo *ActorRepo) StopFollowing(ctx context.Context, id string) error {
	_, err := db.ExecContext(ctx, `delete from Following where ActorID = ?`, id)
	return err
}

func (repo *ActorRepo) CountFollowing(ctx context.Context) (uint, error) {
	var count uint
	err := db.QueryRowContext(ctx, `select count(*) from Following;`).Scan(&count)
	return count, err
}

func (repo *ActorRepo) CountFollowers(ctx context.Context) (uint, error) {
	var count uint
	err := db.QueryRowContext(ctx, `select count(*) from Followers;`).Scan(&count)
	return count, err
}

func (repo *ActorRepo) SubscriptionStatus(ctx context.Context, id string) (types.SubscriptionRelation, error) {
	var (
		theyFollow      bool
		followingStatus sql.NullInt64
	)
	err := db.QueryRowContext(ctx, `
select f.AcceptedStatus, fr.ActorID is not null
from (select 1)
left join Following f on f.ActorID = ?
left join Followers fr on fr.ActorID = ?`, id, id).Scan(&followingStatus, &theyFollow)
	if err != nil {
		return types.SubscriptionNone, err
	}

	iFollow := followingStatus.Valid
	pending := iFollow && followingStatus.Int64 == 0

	switch {
	case pending && theyFollow:
		return types.SubscriptionPendingMutual, nil
	case pending:
		return types.SubscriptionPending, nil
	case iFollow && theyFollow:
		return types.SubscriptionMutual, nil
	case iFollow:
		return types.SubscriptionIFollow, nil
	case theyFollow:
		return types.SubscriptionTheyFollow, nil
	default:
		return types.SubscriptionNone, nil
	}
}

// scanActorsWithKey reads actors (each joined with its public key) from rows
// and closes them.
func scanActorsWithKey(rows *sql.Rows) ([]types.Actor, error) {
	defer rows.Close()

	var actors []types.Actor
	for rows.Next() {
		var a types.Actor
		if err := rows.Scan(&a.ID, &a.PreferredUsername, &a.Inbox, &a.DisplayedName, &a.Summary, &a.Domain, &a.PublicKey.PublicKeyPEM); err != nil {
			return nil, err
		}
		actors = append(actors, a)
	}
	return actors, rows.Err()
}
