// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package likingsvc

import (
	"context"
	"fmt"
	likingports "git.sr.ht/~bouncepaw/betula/ports/liking"
	"git.sr.ht/~bouncepaw/betula/stricks"
	notiftypes "git.sr.ht/~bouncepaw/betula/types/notif"
	"strconv"
)

func (svc *Service) ReceiveLike(ctx context.Context, event likingports.EventLike) error {
	localBookmarkID, err := svc.activityPub.LocalBookmarkIDFromActivityPubID(event.LikedObjectID)
	if err != nil {
		return err
	}

	exists, err := svc.localBookmarkRepo.Exists(ctx, localBookmarkID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("local bookmark %d does not exist", localBookmarkID)
	}

	likeModel := likingports.LikeModel{
		ID:         stricks.NullStringFromString(event.LikeID),
		ActorID:    stricks.NullStringFromString(event.ActorID),
		ObjectID:   strconv.Itoa(localBookmarkID),
		SourceJSON: event.Activity,
	}
	if err := svc.likeRepo.InsertLike(ctx, likeModel); err != nil {
		return err
	}

	err = svc.notifRepo.Store(ctx, notiftypes.KindLike, notiftypes.LikePayload{
		ActorID:    event.ActorID,
		BookmarkID: localBookmarkID,
	})
	if err != nil {
		return err
	}

	svc.logger.Info("Received Like",
		"id", event.LikeID, "actorID", event.ActorID, "objectID", event.LikedObjectID)
	return nil
}

func (svc *Service) ReceiveUndoLike(ctx context.Context, event likingports.EventUndoLike) error {
	err := svc.likeRepo.DeleteLikeBy(ctx, event.LikeID, event.ActorID)
	if err != nil {
		return err
	}

	svc.logger.Info("Received Undo{Like}",
		"id", event.UndoLikeID, "actorID", event.ActorID, "likeID", event.LikeID)
	return nil
}
