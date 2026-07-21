// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package remarkingsvc

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"git.sr.ht/~bouncepaw/betula/pkg/bxerr"
	"git.sr.ht/~bouncepaw/betula/pkg/bxstr"
	apports "git.sr.ht/~bouncepaw/betula/ports/activitypub"
	remarkingports "git.sr.ht/~bouncepaw/betula/ports/remarking"
	"git.sr.ht/~bouncepaw/betula/types"
)

func (svc *Service) ReceiveLegacyRemark(
	ctx context.Context,
	event remarkingports.EventLegacyRemark,
) error {
	return errors.Join(
		svc.receiveLegacyRemarkRemote(ctx, event),
		bxerr.IgnoreAkin(
			svc.receiveLegacyRemarkLocal(ctx, event),
			apports.ErrNotLocal,
		),
	)
}

func (svc *Service) receiveLegacyRemarkLocal(
	ctx context.Context,
	event remarkingports.EventLegacyRemark,
) error {
	localBookmarkID, err := svc.activityPub.LocalBookmarkIDFromActivityPubID(event.ObjectID)
	if err != nil {
		return err
	}

	slog.Info("Received legacy remark",
		"actorID", event.ActorID, "bookmarkID", localBookmarkID, "remarkURL", event.AnnounceID)
	err = svc.remarkRepo.SaveRemark(ctx, localBookmarkID, types.RemarkInfo{
		URL:  event.AnnounceID,
		Name: event.ActorID,
	})
	return err
}

func (svc *Service) receiveLegacyRemarkRemote(
	ctx context.Context,
	event remarkingports.EventLegacyRemark,
) error {
	status, err := svc.actorRepo.SubscriptionStatus(ctx, event.ActorID)
	if err != nil {
		return err
	}
	if !status.WeFollowThem() {
		slog.Info("Received legacy remark, not following them; ignoring",
			"actorID", event.ActorID, "announceID", event.AnnounceID, "objectID", event.ObjectID)
		return nil
	}

	svc.remoteBookmarksRepo.InsertRemoteBookmark(types.RemoteBookmark{
		ID:          event.AnnounceID,
		RemarkedID:  bxstr.NullStringFromString(event.ObjectID),
		ActorID:     event.ActorID,
		PublishedAt: time.Now().UTC().Format(types.TimeLayout),
		Activity:    event.SourceActivity,
	})

	knows, err := svc.activityPub.KnowsRemoteBookmark(event.ObjectID)
	if err != nil {
		return err
	}

	if !knows {
		bm, err := svc.activityPub.DerefRemoteBookmark(ctx, event.ObjectID)
		if err != nil {
			return err
		}
		svc.remoteBookmarksRepo.InsertRemoteBookmark(bm)
	}
	return nil
}

func (svc *Service) ReceiveLegacyUnremark(
	ctx context.Context,
	event remarkingports.EventLegacyUnremark,
) error {
	return errors.Join(
		svc.receiveLegacyUnremarkRemote(ctx, event),
		bxerr.IgnoreAkin(
			svc.receiveLegacyUnremarkLocal(ctx, event),
			apports.ErrNotLocal,
		),
	)
}

func (svc *Service) receiveLegacyUnremarkLocal(
	ctx context.Context,
	event remarkingports.EventLegacyUnremark,
) error {
	localBookmarkID, err := svc.activityPub.LocalBookmarkIDFromActivityPubID(event.ObjectID)
	if err != nil {
		return err
	}

	slog.Info("Received legacy unremark",
		"actorID", event.ActorID, "bookmarkID", localBookmarkID, "remarkURL", event.AnnounceID)
	return svc.remarkRepo.DeleteRemark(ctx, localBookmarkID, event.AnnounceID)
}

func (svc *Service) receiveLegacyUnremarkRemote(
	ctx context.Context,
	event remarkingports.EventLegacyUnremark,
) error {
	return svc.remoteBookmarksRepo.Delete(ctx, event.AnnounceID)
}
