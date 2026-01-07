// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package likingsvc

import (
	"context"
	apports "git.sr.ht/~bouncepaw/betula/ports/activitypub"
	"strconv"
)

func (svc *Service) ActorsThatLiked(
	ctx context.Context,
	bookmarkID int,
) ([]apports.Actor, bool, error) {
	actorIDs, weLiked, err := svc.likeRepo.ActorsThatLiked(ctx, strconv.Itoa(bookmarkID))
	if err != nil {
		return nil, false, err
	}

	actors := make([]apports.Actor, len(actorIDs))
	for i, actorID := range actorIDs {
		actors[i], err = svc.activityPub.ActorByID(ctx, actorID, apports.GetActorsOpts{
			GetPublicKey: false,
		})
		if err != nil {
			return nil, false, err
		}
	}

	return actors, weLiked, nil
}
