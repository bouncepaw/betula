// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package remarkingsvc

import (
	"context"
	"fmt"
	"strconv"
	"time"

	remarkingports "git.sr.ht/~bouncepaw/betula/ports/remarking"
	"git.sr.ht/~bouncepaw/betula/types"
	notiftypes "git.sr.ht/~bouncepaw/betula/types/notif"
)

func (svc *Service) BroadcastCreateRemark(ctx context.Context, bookmark types.Bookmark) error {
	switch {
	case !svc.federationEnabledFn():
		return nil
	case bookmark.Visibility != types.Public:
		return nil
	case bookmark.RemarkedID == nil:
		return fmt.Errorf("bookmark %d. of %q is not a remark", bookmark.ID, bookmark.URL)
	}

	activity, err := svc.assembly.CreateNote(bookmark)
	if err != nil {
		return fmt.Errorf("failed to create remark Note: %w", err)
	}

	err = svc.activityPub.BroadcastToFollowers(ctx, activity)
	if err != nil {
		return fmt.Errorf("failed to broadcast remark Note: %w", err)
	}

	origActor, err := svc.activityPub.AuthorOfRemoteBookmark(*bookmark.RemarkedID)
	if err != nil {
		return fmt.Errorf("failed to determine author of remote bookmark: %w", err)
	}

	status, err := svc.actorRepo.SubscriptionStatus(ctx, origActor.ID())
	if err != nil {
		return fmt.Errorf("failed to get subscription status of original bookmark author: %w", err)
	}

	if status.TheyFollowUs() {
		// Already sent above.
		return nil
	}

	err = origActor.SendSerializedActivity(activity)
	if err != nil {
		return fmt.Errorf("failed to broadcast remark Note to original bookmark author: %w", err)
	}

	return nil
}

func (svc *Service) ReceiveCreateRemark(
	ctx context.Context,
	event remarkingports.EventCreateRemark,
) error {
	if !event.Bookmark.RemarkedID.Valid {
		return fmt.Errorf(
			"failed to receive create remark %s: %w",
			event.Bookmark.ID,
			remarkingports.ErrNotRemark,
		)
	}

	id, err := strconv.Atoi(event.Bookmark.RemarkedID.String)
	if err != nil {
		return remarkingports.ErrRemarkOfRemote
	}

	t, err := time.Parse(types.TimeLayout, event.Bookmark.PublishedAt)
	if err != nil {
		return fmt.Errorf("failed to parse published at %s: %w", event.Bookmark.PublishedAt, err)
	}

	err = svc.remarkRepo.SaveRemark(ctx, id, types.RemarkInfo{
		Timestamp: t,
		URL:       event.Bookmark.ID,
		Name:      event.Bookmark.ActorID,
	})
	if err != nil {
		return fmt.Errorf("failed to save remark: %w", err)
	}

	err = svc.notifRepo.Store(ctx, notiftypes.KindRemark, notiftypes.RemarkPayload{
		ActorID:         event.Bookmark.ActorID,
		BookmarkID:      id,
		RemarkURL:       event.Bookmark.RepresentationURL(),
		Source:          event.Bookmark.Source,
		SourceType:      event.Bookmark.SourceType,
		DescriptionHTML: event.Bookmark.DescriptionHTML, // TODO: add to legacy remarks
	})
	if err != nil {
		return fmt.Errorf("failed to store remark notification: %w", err)
	}

	return nil
}

func (svc *Service) ReceiveUpdateRemark(
	ctx context.Context,
	event remarkingports.EventUpdateRemark,
) error {
	return fmt.Errorf("unimplemented")
}

func (svc *Service) ReceiveDeleteRemark(
	ctx context.Context,
	event remarkingports.EventDeleteRemark,
) error {
	return fmt.Errorf("unimplemented")
}
