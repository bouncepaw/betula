// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package db

import (
	"context"
	"errors"
	likingports "git.sr.ht/~bouncepaw/betula/ports/liking"
	"time"
)

type RepoLikes struct{}

var _ likingports.LikeRepository = &RepoLikes{}

func NewLikeRepo() *RepoLikes {
	return &RepoLikes{}
}

func (repo *RepoLikes) InsertLike(like likingports.LikeModel) error {
	_, err := db.Exec(`
		insert into Likes (ID, ActorID, ObjectID, SourceJSON)
		values (?, ?, ?, ?)
	`, like.ID, like.ActorID, like.ObjectID, like.SourceJSON)
	return err
}

func (repo *RepoLikes) DeleteOurLikeOf(objectID string) error {
	_, err := db.Exec(`
		delete from Likes where ObjectID = ? and ActorID is null
	`, objectID)
	return err
}

func (repo *RepoLikes) StatiFor(objectIDs []string) (map[string]likingports.LikeStatus, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	res := map[string]likingports.LikeStatus{}
	for _, objectID := range objectIDs {
		// From docs: “The count(X) function returns a count of the number of times that
		// X is not NULL in a group. The count(*) function (with no arguments) returns
		// the total number of rows in the group.” Thus, count(*) - count(X) is times
		// X is NULL in a group. If there's a NULL like actor, that means it's us.
		row := tx.QueryRow(`
			select
				count (*),
				(count (*) - count (ActorID)) > 0
			from Likes
			where ObjectID = ?
		`, objectID)

		var status likingports.LikeStatus
		err = row.Scan(&status.Count, &status.LikedByUs)
		if err != nil {
			return nil, errors.Join(err, tx.Rollback())
		}
		res[objectID] = status
	}

	return res, tx.Commit()
}
