// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package likingsvc

import (
	"context"
	"git.sr.ht/~bouncepaw/betula/fediverse/activities"
)

func (svc *Service) Unlike(ctx context.Context, bookmarkID string) error {
	isRemote, err := svc.validBookmarkID(ctx, bookmarkID)
	if err != nil {
		return err
	}

	if err := svc.likeRepo.DeleteOurLikeOf(ctx, bookmarkID); err != nil {
		return err
	}

	if isRemote {
		go func() {
			_ = svc.sendUndoLikeToRemoteActor(bookmarkID)
		}()

		// Optimistic UI.
		err = svc.likeCollectionRepo.DecrementIfPresent(ctx, bookmarkID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (svc *Service) sendUndoLikeToRemoteActor(bookmarkID string) error {
	actor, err := svc.activityPub.AuthorOfRemoteBookmark(bookmarkID)
	if err != nil {
		svc.logger.Error("Failed to get author to make Undo{Like} activity",
			"err", err, "bookmarkID", bookmarkID)
		return err
	}

	activity, err := activities.NewUndoLike(bookmarkID, actor.ID())
	if err != nil {
		svc.logger.Error("Failed to make Undo{Like} activity",
			"err", err, "bookmarkID", bookmarkID)
		return err
	}

	err = actor.SendSerializedActivity(activity)
	if err != nil {
		svc.logger.Error("Failed to send activity Undo{Like} to inbox",
			"err", err, "activity", string(activity))
		return err
	}

	return nil
}
