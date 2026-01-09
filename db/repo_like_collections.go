// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package db

import (
	"context"
	likingports "git.sr.ht/~bouncepaw/betula/ports/liking"
)

type RepoLikeCollections struct{}

var _ likingports.LikeCollectionRepository = &RepoLikeCollections{}

func NewLikeCollectionRepo() *RepoLikeCollections {
	return &RepoLikeCollections{}
}

func (repo *RepoLikeCollections) UpsertLikeCollection(
	ctx context.Context,
	likeCollection likingports.LikeCollectionModel,
) error {
	_, err := db.ExecContext(ctx, `
		insert into LikeCollections (ID, LikedObjectID, TotalItems, SourceJSON)
		values (?, ?, ?, ?)
		on conflict do update set
		    TotalItems=excluded.TotalItems,
		    SourceJSON=excluded.SourceJSON
	`, likeCollection.ID,
		likeCollection.LikedObjectID,
		likeCollection.TotalItems,
		likeCollection.SourceJSON)
	return err
}

func (repo *RepoLikeCollections) GetTotalItemsFor(
	ctx context.Context,
	objectID string,
) (int, error) {
	row := db.QueryRowContext(ctx, `
		select TotalItems from LikeCollections
		where LikedObjectID = ?
	`, objectID)

	var totalItems int
	err := row.Scan(&totalItems)
	return totalItems, err
}

func (repo *RepoLikeCollections) IncrementIfPresent(
	ctx context.Context,
	objectID string,
) error {
	_, err := db.ExecContext(ctx, `
		update LikeCollections 
		set TotalItems = TotalItems + 1
		where LikedObjectID = ?
	`, objectID)
	return err
}

func (repo *RepoLikeCollections) DecrementIfPresent(
	ctx context.Context,
	objectID string,
) error {
	_, err := db.ExecContext(ctx, `
		update LikeCollections
		set TotalItems = TotalItems - 1
		where LikedObjectID = ?
	`, objectID)
	return err
}
