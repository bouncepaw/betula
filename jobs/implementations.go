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
	"git.sr.ht/~bouncepaw/betula/fediverse/activities"
	"git.sr.ht/~bouncepaw/betula/jobs/jobtype"
	"git.sr.ht/~bouncepaw/betula/pkg/stricks"
	"git.sr.ht/~bouncepaw/betula/settings"
	"git.sr.ht/~bouncepaw/betula/types"
	notiftypes "git.sr.ht/~bouncepaw/betula/types/notif"
)

var repoNotif = db.New()

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
	jobtype.SendUndoAnnounce:    notifyAboutMyUnrepost,
	jobtype.SendAcceptFollow:    callForJSON[activities.FollowReport](jobtype.SendAcceptFollow, sendAcceptFollow),
	jobtype.SendRejectFollow:    callForJSON[activities.FollowReport](jobtype.SendRejectFollow, sendRejectFollow),
	jobtype.ReceiveAcceptFollow: callForJSON[activities.FollowReport](jobtype.ReceiveAcceptFollow, receiveAcceptFollow),
	jobtype.ReceiveRejectFollow: callForJSON[activities.FollowReport](jobtype.ReceiveRejectFollow, receiveRejectFollow),
	jobtype.SendCreateNote:      broadcastToFollowers,
	jobtype.SendDeleteNote:      broadcastToFollowers,
	jobtype.SendUpdateNote:      broadcastToFollowers,
}

func broadcastToFollowers(job jobtype.Job) {
	// The payload is a []byte we have to send to every follower.
	payload, ok := job.Payload.([]byte)
	if !ok {
		slog.Error("Unexpected payload for broadcast", "category", job.Category, "payloadType", fmt.Sprintf("%T", payload))
		return
	}

	followers := db.GetFollowers()
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

func receiveAcceptFollow(report activities.FollowReport) {
	// We assume that they are actually talking about us, because we filtered out wrong activities in the inbox.

	if status := db.SubscriptionStatus(report.ObjectID); status.IsPending() {
		slog.Info("Received Accept{Follow}", "objectID", report.ObjectID)
		db.MarkAsSurelyFollowing(report.ObjectID)
	} else {
		slog.Warn("Received invalid Accept{Follow}, ignoring", "status", status, "activity", report.OriginalActivity)
	}
}

func receiveRejectFollow(report activities.FollowReport) {
	// We assume that they are actually talking about us, because we filtered out wrong activities in the inbox.

	if status := db.SubscriptionStatus(report.ObjectID); status.IsPending() {
		slog.Info("Received Reject{Follow}", "objectID", report.ObjectID)
		db.StopFollowing(report.ObjectID)
	} else {
		slog.Warn("Received invalid Reject{Follow}, ignoring", "status", status, "activity", report.OriginalActivity)
	}
}

func sendRejectFollow(report activities.FollowReport) {
	if !stricks.ValidURL(report.ActorID) {
		slog.Error("Invalid actor ID, dropping activity", "actorID", report.ActorID)
	}

	activity, err := activities.NewReject(report.OriginalActivity)
	if err = SendActivityToInbox(activity, fediverse.RequestActorInboxByID(report.ActorID)); err != nil {
		slog.Error("Failed to send Reject activity", "err", err)
	}
}

func sendAcceptFollow(report activities.FollowReport) {
	if !stricks.ValidURL(report.ActorID) {
		slog.Error("Dropping activity", "reason", "invalid actor ID", "actorID", report.ActorID)
		return
	}

	activity, err := activities.NewAccept(report.OriginalActivity)
	if err = SendActivityToInbox(activity, fediverse.RequestActorInboxByID(report.ActorID)); err != nil {
		slog.Error("Failed to send activity", "err", err, "recipient", report.ActorID)
	} else {
		db.AddFollower(report.ActorID)

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

	post, found := db.GetBookmarkByID(postId)
	if !found {
		slog.Error("Failed to notify about non-existent bookmark", "bookmarkID", postId)
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

	activity, err := activities.NewAnnounce(
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
		slog.Error("Failed to send unrepost activity", "err", err)
		return
	}

	broadcastToFollowers(jobtype.Job{
		Category: jobtype.SendAnnounce,
		Payload:  activity,
	})
}

func notifyAboutMyUnrepost(job jobtype.Job) {
	var report activities.UndoAnnounceReport

	switch v := job.Payload.(type) {
	case []byte:
		if err := json.Unmarshal(v, &report); err != nil {
			slog.Error("Failed to unmarshal UndoAnnounceReport", "err", err, "payload", v)
			return
		}

		// TODO: avoid sending twice if unreposted from follower
		err := sendActivity(report.ObjectID, v)
		if err != nil {
			slog.Error("Failed to send unrepost notification", "err", err)
		}
		broadcastToFollowers(jobtype.Job{
			Category: jobtype.SendUndoAnnounce,
			Payload:  v,
		})
	default:
		slog.Error("Bad payload for ReceiveUnrepost job", "payload", v)
		return
	}
}
