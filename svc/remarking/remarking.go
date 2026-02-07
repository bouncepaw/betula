// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package remarkingsvc

import (
	"context"
	"log/slog"

	"git.sr.ht/~bouncepaw/betula/db"
	apports "git.sr.ht/~bouncepaw/betula/ports/activitypub"
	remarkingports "git.sr.ht/~bouncepaw/betula/ports/remarking"
	"git.sr.ht/~bouncepaw/betula/types"
)

type Service struct {
	activityPub apports.ActivityPub
}

var _ remarkingports.Service = &Service{}

func New(
	activityPub apports.ActivityPub,
) *Service {
	return &Service{
		activityPub: activityPub,
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
	// TODO: make a repo.
	db.SaveRepost(localBookmarkID, types.RepostInfo{
		URL:  event.AnnounceID,
		Name: event.ActorID,
	})
	return nil
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
	// TODO: make a repo.
	db.DeleteRepost(localBookmarkID, event.AnnounceID)
	return nil
}
