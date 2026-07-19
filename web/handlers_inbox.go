// SPDX-FileCopyrightText: 2025 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package web

import (
	"io"
	"log/slog"
	"net/http"

	"git.sr.ht/~bouncepaw/betula/fediverse"
	"git.sr.ht/~bouncepaw/betula/fediverse/signing"
	"git.sr.ht/~bouncepaw/betula/jobs"
	"git.sr.ht/~bouncepaw/betula/jobs/jobtype"
	"git.sr.ht/~bouncepaw/betula/pkg/bxstr"
	apports "git.sr.ht/~bouncepaw/betula/ports/activitypub"
	"git.sr.ht/~bouncepaw/betula/ports/liking"
	"git.sr.ht/~bouncepaw/betula/ports/remarking"
)

func postInbox(w http.ResponseWriter, rq *http.Request) {
	data, err := io.ReadAll(io.LimitReader(rq.Body, 32*1024*1024)) // Read no more than 32 KiB.
	if err != nil {
		slog.Error("Failed to read inbox body", "err", err, "prefix", string(data))
		http.Error(w, "Failed to read inbox body, input must be less than 32 KiB", http.StatusRequestEntityTooLarge)
		return
	}

	report, err := ctrl.Guesser.Guess(data)
	if err != nil {
		slog.Error("Failed to parse incoming activity", "err", err)
		return
	}
	if report == nil {
		// Ignored
		return
	}

	switch report := report.(type) {
	case apports.CreateNoteReport:
		status, err := ctrl.RepoActor.SubscriptionStatus(rq.Context(), report.Bookmark.ActorID)
		if err != nil {
			slog.Error("Failed to get subscription status", "actorID", report.Bookmark.ActorID, "err", err)
			return
		}
		if !status.WeFollowThem() {
			slog.Info("Received bookmark from non-follower, ignoring", "actorID", report.Bookmark.ActorID, "bookmarkID", report.Bookmark.ID)
			return
		}

		slog.Info("Received bookmark from follower", "actorID", report.Bookmark.ActorID, "bookmarkID", report.Bookmark.ID)
		ctrl.RepoRemoteBookmark.InsertRemoteBookmark(report.Bookmark)

		if report.LikesCollection != nil {
			event := likingports.EventLikeCollectionSeen{
				ID:            report.LikesCollection.ID,
				Type:          report.LikesCollection.Type,
				TotalItems:    report.LikesCollection.TotalItems,
				LikedObjectID: report.Bookmark.ID,
				SourceJSON:    data,
			}
			err = ctrl.SvcLiking.ReceiveLikeCollection(rq.Context(), event)
			if err != nil {
				slog.Error("Failed to receive like collection", "err", err)
			}
		}

	case apports.UpdateNoteReport:
		exists, err := ctrl.RepoRemoteBookmark.Exists(report.Bookmark.ID)
		if err != nil {
			slog.Error("Failed to check if bookmark exists", "err", err, "bookmarkID", report.Bookmark.ID)
			return
		}

		if !exists {
			// TODO: maybe store them?
			slog.Info("Received update for unknown bookmark, ignoring", "actorID", report.Bookmark.ActorID, "bookmarkID", report.Bookmark.ID)
			return
		}

		slog.Info("Updated remote bookmark", "actorID", report.Bookmark.ActorID, "bookmarkID", report.Bookmark.ID)
		ctrl.RepoRemoteBookmark.UpdateRemoteBookmark(report.Bookmark)

		if report.LikesCollection != nil {
			event := likingports.EventLikeCollectionSeen{
				ID:            report.LikesCollection.ID,
				Type:          report.LikesCollection.Type,
				TotalItems:    report.LikesCollection.TotalItems,
				LikedObjectID: report.Bookmark.ID,
			}
			slog.Info("The update contained a likes collection; handling",
				"event", event)
			event.SourceJSON = data // not including in logs

			err = ctrl.SvcLiking.ReceiveLikeCollection(rq.Context(), event)
			if err != nil {
				slog.Error("Failed to receive like collection", "err", err)
			}
		}

	case apports.DeleteNoteReport:
		_, err := fediverse.RequestActorByID(report.ActorID)
		if err != nil {
			slog.Error("Failed to fetch actor", "err", err)
			return
		}
		if signedOK := signing.VerifyRequestSignature(rq, data); !signedOK {
			slog.Error("Failed to verify signature", "actorID", report.ActorID)
			return
		}

		slog.Info("Deleted remote bookmark", "actorID", report.ActorID, "bookmarkID", report.BookmarkID)
		err = ctrl.RepoRemoteBookmark.Delete(rq.Context(), report.BookmarkID)
		if err != nil {
			slog.Error("Failed to delete remote bookmark", "err", err)
		}

	case apports.UndoAnnounceReport:
		_, err := fediverse.RequestActorByID(report.ActorID)
		if err != nil {
			slog.Error("Failed to fetch actor", "err", err)
			return
		}
		if signedOK := signing.VerifyRequestSignature(rq, data); !signedOK {
			slog.Error("Failed to verify signature", "actorID", report.ActorID)
			return
		}

		event := remarkingports.EventLegacyUnremark{
			ActorID:    report.ActorID,
			AnnounceID: report.AnnounceID,
			ObjectID:   report.ObjectID,
		}
		if err = ctrl.SvcRemarking.ReceiveLegacyUnremark(rq.Context(), event); err != nil {
			slog.Error("Failed to receive legacy unremark", "err", err, "event", event)
		}

	case apports.AnnounceReport:
		_, err := fediverse.RequestActorByID(report.ActorID)
		if err != nil {
			slog.Error("Failed to fetch actor", "err", err)
			return
		}
		if signedOK := signing.VerifyRequestSignature(rq, data); !signedOK {
			slog.Error("Failed to verify signature", "actorID", report.ActorID)
			return
		}

		event := remarkingports.EventLegacyRemark{
			ActorID:        report.ActorID,
			AnnounceID:     report.AnnounceID,
			ObjectID:       report.ObjectID,
			SourceActivity: data,
		}
		if err = ctrl.SvcRemarking.ReceiveLegacyRemark(rq.Context(), event); err != nil {
			slog.Error("Failed to receive legacy remark", "err", err, "event", event)
		}

	case apports.UndoFollowReport:
		// We'll schedule no job because we are making no network request to handle this.
		if report.ObjectID != fediverse.OurID() {
			slog.Info("Unfollow request for someone else, ignoring", "actorID", report.ActorID, "objectID", report.ObjectID)
			return
		}
		status, err := ctrl.RepoActor.SubscriptionStatus(rq.Context(), report.ActorID)
		if err != nil {
			slog.Error("Failed to get subscription status", "actorID", report.ActorID, "err", err)
			return
		}
		if !status.TheyFollowUs() {
			slog.Info("Unfollow from non-follower, ignoring", "actorID", report.ActorID)
			return
		}
		slog.Info("Follower unfollowed us", "actorID", report.ActorID)
		if err := ctrl.RepoActor.RemoveFollower(rq.Context(), report.ActorID); err != nil {
			slog.Error("Failed to remove follower", "actorID", report.ActorID, "err", err)
			return
		}

	case apports.FollowReport:
		_, err := fediverse.RequestActorByID(report.ActorID)
		if err != nil {
			slog.Error("Failed to fetch actor", "err", err)
			return
		}
		if signedOK := signing.VerifyRequestSignature(rq, data); !signedOK {
			slog.Error("Failed to verify signature", "actorID", report.ActorID)
			return
		}

		if report.ObjectID == fediverse.OurID() {
			slog.Info("Someone asked to follow us", "actorID", report.ActorID)
			jobs.ScheduleJSON(jobtype.SendAcceptFollow, report)
		} else {
			slog.Info("Follow request for someone else", "actorID", report.ActorID, "objectID", report.ObjectID)
			jobs.ScheduleJSON(jobtype.ReceiveRejectFollow, report)
		}

	case apports.AcceptReport:
		_, err := fediverse.RequestActorByID(report.ActorID)
		if err != nil {
			slog.Error("Failed to fetch actor", "err", err)
			return
		}
		if signedOK := signing.VerifyRequestSignature(rq, data); !signedOK {
			slog.Error("Failed to verify signature", "actorID", report.ActorID)
			return
		}

		switch report.Object["type"] {
		case "Follow":
			report := apports.FollowReport{
				ActorID:          bxstr.StringifyAnything(report.Object["actor"]),
				ObjectID:         bxstr.StringifyAnything(report.Object["object"]),
				OriginalActivity: report.Object,
			}
			jobs.ScheduleJSON(jobtype.ReceiveAcceptFollow, report)
		}

	case apports.RejectReport:
		_, err := fediverse.RequestActorByID(report.ActorID)
		if err != nil {
			slog.Error("Failed to fetch actor", "err", err)
			return
		}
		if signedOK := signing.VerifyRequestSignature(rq, data); !signedOK {
			slog.Error("Failed to verify signature", "actorID", report.ActorID)
			return
		}

		switch report.Object["type"] {
		case "Follow":
			report := apports.FollowReport{
				ActorID:          bxstr.StringifyAnything(report.Object["actor"]),
				ObjectID:         bxstr.StringifyAnything(report.Object["object"]),
				OriginalActivity: report.Object,
			}
			jobs.ScheduleJSON(jobtype.ReceiveRejectFollow, report)
		}

	case apports.LikeReport:
		_, err := fediverse.RequestActorByID(report.ActorID)
		if err != nil {
			slog.Error("Failed to fetch actor", "err", err)
			return
		}
		if signedOK := signing.VerifyRequestSignature(rq, data); !signedOK {
			slog.Error("Failed to verify signature", "actorID", report.ActorID)
			return
		}

		event := likingports.EventLike{
			LikeID:        report.ID,
			ActorID:       report.ActorID,
			LikedObjectID: report.ObjectID,
			Activity:      report.Activity,
		}
		err = ctrl.SvcLiking.ReceiveLike(rq.Context(), event)
		if err != nil {
			event.Activity = nil
			slog.Error("Failed to receive Like",
				"event", event, "err", err, "activity", report.Activity)
			return
		}

	case apports.UndoLikeReport:
		_, err := fediverse.RequestActorByID(report.Object.ActorID)
		if err != nil {
			slog.Error("Failed to fetch actor", "err", err)
			return
		}
		if signedOK := signing.VerifyRequestSignature(rq, data); !signedOK {
			slog.Error("Failed to verify signature", "actorID", report.Object.ActorID)
			return
		}

		event := likingports.EventUndoLike{
			UndoLikeID: report.ID,
			ActorID:    report.Object.ActorID,
			LikeID:     report.Object.ID,
			Activity:   report.Activity,
		}
		err = ctrl.SvcLiking.ReceiveUndoLike(rq.Context(), event)
		if err != nil {
			event.Activity = nil
			slog.Error("Failed to receive Undo{Like}",
				"event", event, "err", err, "activity", report.Activity)
			return
		}

	default:
		// Not meant to happen
		slog.Error("Invalid report type; this is a bug")
	}
}
