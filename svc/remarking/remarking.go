// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package remarkingsvc

import (
	"context"
	"log/slog"

	apports "git.sr.ht/~bouncepaw/betula/ports/activitypub"
	remarkingports "git.sr.ht/~bouncepaw/betula/ports/remarking"
	"git.sr.ht/~bouncepaw/betula/types"
)

type Service struct {
	activityPub apports.ActivityPub
	repo        remarkingports.Repository
}

var _ remarkingports.Service = &Service{}

func New(
	activityPub apports.ActivityPub,
	repo remarkingports.Repository,
) *Service {
	return &Service{
		activityPub: activityPub,
		repo:        repo,
	}
}

func (svc *Service) ReceiveLegacyRemark(
	ctx context.Context,
	event remarkingports.EventLegacyRemark,
) error {
	localBookmarkID, err := svc.activityPub.LocalBookmarkIDFromActivityPubID(event.ObjectID)
	if err != nil {
		return err
	}

	slog.Info("Received legacy remark",
		"actorID", event.ActorID, "bookmarkID", localBookmarkID, "remarkURL", event.AnnounceID)
	return svc.repo.SaveRemark(ctx, localBookmarkID, types.RepostInfo{
		URL:  event.AnnounceID,
		Name: event.ActorID,
	})
}

func (svc *Service) ReceiveLegacyUnremark(
	ctx context.Context,
	event remarkingports.EventLegacyUnremark,
) error {
	localBookmarkID, err := svc.activityPub.LocalBookmarkIDFromActivityPubID(event.ObjectID)
	if err != nil {
		return err
	}

	slog.Info("Received legacy unremark",
		"actorID", event.ActorID, "bookmarkID", localBookmarkID, "remarkURL", event.AnnounceID)
	return svc.repo.DeleteRemark(ctx, localBookmarkID, event.AnnounceID)
}
