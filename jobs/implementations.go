// SPDX-FileCopyrightText: 2024 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2025 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2026 Danila Gorelko
// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"git.sr.ht/~bouncepaw/betula/db"
	"git.sr.ht/~bouncepaw/betula/fediverse"
	"git.sr.ht/~bouncepaw/betula/jobs/jobtype"
	"git.sr.ht/~bouncepaw/betula/pkg/bxstr"
	apports "git.sr.ht/~bouncepaw/betula/ports/activitypub"
	"git.sr.ht/~bouncepaw/betula/settings"
	"git.sr.ht/~bouncepaw/betula/svc/activitypub/assembly"
	"git.sr.ht/~bouncepaw/betula/types"
	notiftypes "git.sr.ht/~bouncepaw/betula/types/notif"
)

// TODO: all shall be in services one day...
var (
	repoNotif          = db.New()
	repoLocalBookmarks = db.NewLocalBookmarksRepo()
	repoActor          = db.NewActorRepo()
	asm                = assembly.New(settings.SiteURL, settings.AdminUsername)
)

func callForJSON[T any](jobcat jobtype.JobCategory, next func(T)) func(jobtype.Job) {
	return func(job jobtype.Job) {
		data, ok := job.Payload.([]byte)
		if !ok {
			slog.Error("Unexpected payload for job", "category", jobcat, "payloadType", fmt.Sprintf("%T", job.Payload), "payload", job.Payload)
			return
		}

		var report T
		err := json.Unmarshal(data, &report)
		if err != nil {
			slog.Error("Failed to unmarshal payload for job", "category", jobcat, "err", err)
			return
		}

		next(report)
	}
}

var catmap = map[jobtype.JobCategory]func(job jobtype.Job){
	jobtype.SendAnnounce:        notifyAboutMyRepost,
	jobtype.SendAcceptFollow:    callForJSON[apports.FollowReport](jobtype.SendAcceptFollow, sendAcceptFollow),
	jobtype.SendRejectFollow:    callForJSON[apports.FollowReport](jobtype.SendRejectFollow, sendRejectFollow),
	jobtype.ReceiveAcceptFollow: callForJSON[apports.FollowReport](jobtype.ReceiveAcceptFollow, receiveAcceptFollow),
	jobtype.ReceiveRejectFollow: callForJSON[apports.FollowReport](jobtype.ReceiveRejectFollow, receiveRejectFollow),
	jobtype.SendCreateNote:      broadcastToFollowers,
	jobtype.SendDeleteNote:      broadcastToFollowers,
	jobtype.SendUpdateNote:      broadcastToFollowers,
}

func byteCast(raw any) ([]byte, error) {
	bytes, ok := raw.([]byte)
	if ok {
		return bytes, nil
	}

	jsonBytes, ok := raw.(json.RawMessage)
	if ok {
		return jsonBytes, nil
	}
	return nil, fmt.Errorf("unexpected type for byte cast: %T", raw)
}

func broadcastToFollowers(job jobtype.Job) {
	// The payload is a []byte we have to send to every follower.
	payload, err := byteCast(job.Payload)
	if err != nil {
		slog.Error("Unexpected payload for broadcast",
			"category", job.Category, "payloadType", fmt.Sprintf("%T", payload), "err", err)
		return
	}

	followers, err := repoActor.GetFollowers(context.Background())
	if err != nil {
		slog.Error("Failed to fetch followers for broadcast", "category", job.Category, "err", err)
		return
	}
	if len(followers) == 0 {
		slog.Info("Nobody to broadcast to :-(")
		return
	}

	slog.Info("Broadcasting to followers", "payload", job.Payload)

	succSends := len(followers)

	// This loop might take some time (n = len(followers)) because we don't parallelize it.
	// I don't we should parallelize it.
	for _, follower := range followers {
		err := SendQuietActivityToInbox(payload, follower.Inbox)
		if err != nil {
			slog.Error("Failed to send to a follower", "inbox", follower.Inbox, "err", err)
			succSends--
		}
	}

	slog.Info("Sent to followers", "category", job.Category, "success", succSends, "total", len(followers))
}

func receiveAcceptFollow(report apports.FollowReport) {
	// We assume that they are actually talking about us, because we filtered out wrong activities in the inbox.

	ctx := context.Background()
	status, err := repoActor.SubscriptionStatus(ctx, report.ObjectID)
	if err != nil {
		slog.Error("Failed to get subscription status", "objectID", report.ObjectID, "err", err)
		return
	}
	if status.IsPending() {
		slog.Info("Received Accept{Follow}", "objectID", report.ObjectID)
		if err := repoActor.MarkAsSurelyFollowing(ctx, report.ObjectID); err != nil {
			slog.Error("Failed to mark as surely following", "objectID", report.ObjectID, "err", err)
		}
	} else {
		slog.Warn("Received invalid Accept{Follow}, ignoring", "status", status, "activity", report.OriginalActivity)
	}
}

func receiveRejectFollow(report apports.FollowReport) {
	// We assume that they are actually talking about us, because we filtered out wrong activities in the inbox.

	ctx := context.Background()
	status, err := repoActor.SubscriptionStatus(ctx, report.ObjectID)
	if err != nil {
		slog.Error("Failed to get subscription status", "objectID", report.ObjectID, "err", err)
		return
	}
	if status.IsPending() {
		slog.Info("Received Reject{Follow}", "objectID", report.ObjectID)
		if err := repoActor.StopFollowing(ctx, report.ObjectID); err != nil {
			slog.Error("Failed to stop following", "objectID", report.ObjectID, "err", err)
		}
	} else {
		slog.Warn("Received invalid Reject{Follow}, ignoring", "status", status, "activity", report.OriginalActivity)
	}
}

func sendRejectFollow(report apports.FollowReport) {
	if !bxstr.IsValidURL(report.ActorID) {
		slog.Error("Invalid actor ID, dropping activity", "actorID", report.ActorID)
	}

	activity, err := asm.NewReject(report.OriginalActivity)
	if err = SendActivityToInbox(activity, fediverse.RequestActorInboxByID(report.ActorID)); err != nil {
		slog.Error("Failed to send Reject activity", "err", err)
	}
}

func sendAcceptFollow(report apports.FollowReport) {
	if !bxstr.IsValidURL(report.ActorID) {
		slog.Error("Dropping activity", "reason", "invalid actor ID", "actorID", report.ActorID)
		return
	}

	activity, err := asm.NewAccept(report.OriginalActivity)
	if err = SendActivityToInbox(activity, fediverse.RequestActorInboxByID(report.ActorID)); err != nil {
		slog.Error("Failed to send activity", "err", err, "recipient", report.ActorID)
	} else {
		if err := repoActor.AddFollower(context.Background(), report.ActorID); err != nil {
			slog.Error("Failed to add follower", "actorID", report.ActorID, "err", err)
		}

		err = repoNotif.Store(context.Background(), notiftypes.KindFollow, notiftypes.FollowPayload{
			ActorID: report.ActorID,
		})
		if err != nil {
			slog.Error("Failed to store follow notification", "err", err)
		}
	}
}

func notifyAboutMyRepost(job jobtype.Job) {
	var postId int
	switch v := job.Payload.(type) {
	case int64:
		postId = int(v)
	default:
		slog.Error("Unexpected payload for notify job", "payloadType", fmt.Sprintf("%T", v), "payload", v)
		return
	}

	post, err := repoLocalBookmarks.GetBookmarkByID(context.Background(), postId)
	if err != nil {
		slog.Error("Failed to notify about bookmark", "bookmarkID", postId, "err", err)
		return
	}

	if post.RepostOf == nil {
		slog.Warn("Bookmark is not a repost, skipping notify", "bookmarkID", postId)
		return
	}

	if post.Visibility != types.Public {
		slog.Info("Bookmark (repost) is not public, not notifying", "bookmarkID", postId)
		return
	}

	activity, err := asm.NewAnnounce(
		*post.RepostOf,
		fmt.Sprintf("%s/%d", settings.SiteURL(), postId),
	)
	if err != nil {
		slog.Error("Failed to create Announce activity", "err", err)
		return
	}

	// TODO: this will have to change. Avoid sending twice if reposting a follower
	err = sendActivity(*post.RepostOf, activity)
	if err != nil {
		slog.Error("Failed to send repost activity", "err", err)
		return
	}

	broadcastToFollowers(jobtype.Job{
		Category: jobtype.SendAnnounce,
		Payload:  activity,
	})
}
