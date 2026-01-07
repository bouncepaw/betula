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
