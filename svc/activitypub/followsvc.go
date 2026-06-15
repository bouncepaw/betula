// SPDX-FileCopyrightText: 2025 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2025 Danila Gorelko
// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package apsvc

import (
	"context"
	"fmt"
	"log/slog"

	"git.sr.ht/~bouncepaw/betula/fediverse"
	"git.sr.ht/~bouncepaw/betula/fediverse/activities"
	"git.sr.ht/~bouncepaw/betula/jobs"
	apports "git.sr.ht/~bouncepaw/betula/ports/activitypub"
)

type FollowService struct {
	repo apports.ActorRepository
}

var _ apports.FollowService = &FollowService{}

func NewFollowService(repo apports.ActorRepository) *FollowService {
	return &FollowService{repo: repo}
}

func (svc *FollowService) Follow(ctx context.Context, nickname string) error {
	actor, err := fediverse.RequestActorByNickname(nickname)
	if err != nil {
		slog.Error("Failed to get actor to follow", "err", err, "nickname", nickname)
		return fmt.Errorf("failed to get actor %s: %w", nickname, err)
	}

	activity, err := activities.NewFollowFromUs(actor.ID)
	if err != nil {
		slog.Error("Failed to create Follow activity", "err", err)
		return fmt.Errorf("failed to create Follow activity: %w", err)
	}

	err = jobs.SendActivityToInbox(activity, actor.Inbox)
	if err != nil {
		slog.Error("Failed to send Follow activity", "err", err)
		return fmt.Errorf("failed to send activity to inbox: %w", err)
	}

	err = svc.repo.AddPendingFollowing(ctx, actor.ID)
	if err != nil {
		slog.Error("Failed to add pending following", "actorID", actor.ID, "err", err)
		return fmt.Errorf("failed to add pending following %s: %w", actor.ID, err)
	}
	return nil
}

func (svc *FollowService) Unfollow(ctx context.Context, nickname string) error {
	actor, err := fediverse.RequestActorByNickname(nickname)
	if err != nil {
		slog.Error("Failed to get actor to unfollow", "err", err, "nickname", nickname)
		return fmt.Errorf("failed to get actor %s: %w", nickname, err)
	}

	activity, err := activities.NewUndoFollowFromUs(actor.ID)
	if err != nil {
		slog.Error("Failed to create Undo{Follow} activity", "err", err)
		return fmt.Errorf("failed to create Undo{Follow} activity: %w", err)
	}

	err = jobs.SendActivityToInbox(activity, actor.Inbox)
	if err != nil {
		slog.Error("Failed to send Undo{Follow} activity; ignoring", "err", err)
		// Proceed with unfollowing even if sending failed
	}

	err = svc.repo.StopFollowing(ctx, actor.ID)
	if err != nil {
		slog.Error("Failed to stop following", "actorID", actor.ID, "err", err)
		return fmt.Errorf("failed to stop following: %w", err)
	}
	return nil
}
