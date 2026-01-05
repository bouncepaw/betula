// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package likingsvc

import (
	"context"
	"git.sr.ht/~bouncepaw/betula/fediverse/activities"
	likingports "git.sr.ht/~bouncepaw/betula/ports/liking"
)

func (svc *Service) Like(ctx context.Context, bookmarkID string) error {
	isRemote, err := svc.validBookmarkID(ctx, bookmarkID)
	if err != nil {
		return err
	}

	err = svc.likeRepo.InsertLike(ctx, likingports.LikeModel{
		ObjectID: bookmarkID,
	})
	if err != nil {
		return err
	}

	if isRemote {
		go func() {
			_ = svc.sendLikeToRemoteActor(bookmarkID)
		}()
	}
	return nil
}

func (svc *Service) sendLikeToRemoteActor(bookmarkID string) error {
	actor, err := svc.activityPub.AuthorOfRemoteBookmark(bookmarkID)
	if err != nil {
		svc.logger.Error("Failed to get author to make Like activity",
			"err", err, "bookmarkID", bookmarkID)
		return err
	}

	activity, err := activities.NewLike(bookmarkID, actor.ID())
	if err != nil {
		svc.logger.Error("Failed to make Like activity",
			"err", err, "bookmarkID", bookmarkID)
		return err
	}

	err = actor.SendSerializedActivity(activity)
	if err != nil {
		svc.logger.Error("Failed to send Like activity to inbox",
			"err", err, "activity", string(activity))
		return err
	}

	return nil
}
