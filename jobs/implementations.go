package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"git.sr.ht/~bouncepaw/betula/db"
	"git.sr.ht/~bouncepaw/betula/fediverse"
	"git.sr.ht/~bouncepaw/betula/fediverse/activities"
	"git.sr.ht/~bouncepaw/betula/jobs/jobtype"
	"git.sr.ht/~bouncepaw/betula/readpage"
	"git.sr.ht/~bouncepaw/betula/settings"
	"git.sr.ht/~bouncepaw/betula/stricks"
	"git.sr.ht/~bouncepaw/betula/types"
	notiftypes "git.sr.ht/~bouncepaw/betula/types/notif"
	"log"
	"log/slog"
	"strconv"
	"strings"
)

var repoNotif = db.New()

func callForJSON[T any](jobcat jobtype.JobCategory, next func(T)) func(jobtype.Job) {
	return func(job jobtype.Job) {
		data, ok := job.Payload.([]byte)
		if !ok {
			log.Printf("Unexpected payload for %s job of type %T: %v\n", jobcat, job.Payload, job.Payload)
			return
		}

		var report T
		err := json.Unmarshal(data, &report)
		if err != nil {
			log.Printf("When unmarshaling payload for job %s: %s\n", jobcat, err)
			return
		}

		next(report)
	}
}

var catmap = map[jobtype.JobCategory]func(job jobtype.Job){
	jobtype.SendAnnounce:        notifyAboutMyRepost,
	jobtype.ReceiveAnnounce:     verifyJob,
	jobtype.ReceiveUndoAnnounce: receiveUnrepostJob,
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
		log.Printf("Unexpected payload of type %T for %s: %v\n", payload, job.Category, payload)
		return
	}

	followers := db.GetFollowers()
	if len(followers) == 0 {
		log.Println("Nobody to broadcast to :-(")
		return
	}

	log.Printf("Broadcasting to followers: %s\n", job.Payload)

	succSends := len(followers)

	// This loop might take some time (n = len(followers)) because we don't parallelize it.
	// I don't we should parallelize it.
	for _, follower := range followers {
		err := SendQuietActivityToInbox(payload, follower.Inbox)
		if err != nil {
			log.Printf("While sending to %s: %s\n", follower.Inbox, err)
			succSends--
		}
	}

	log.Printf("Sent %s to %d out of %d followers.\n", job.Category, succSends, len(followers))
}

func receiveAcceptFollow(report activities.FollowReport) {
	// We assume that they are actually talking about us, because we filtered out wrong activities in the inbox.

	if status := db.SubscriptionStatus(report.ObjectID); status.IsPending() {
		log.Printf("Received Accept{Follow} to %s\n", report.ObjectID)
		db.MarkAsSurelyFollowing(report.ObjectID)
	} else {
		log.Printf("Received an invalid Accept{Follow}, status is %s. Ignoring. Activity: %s\n", status, report.OriginalActivity)
	}
}

func receiveRejectFollow(report activities.FollowReport) {
	// We assume that they are actually talking about us, because we filtered out wrong activities in the inbox.

	if status := db.SubscriptionStatus(report.ObjectID); status.IsPending() {
		log.Printf("Received Reject{Follow} to %s\n", report.ObjectID)
		db.StopFollowing(report.ObjectID)
	} else {
		log.Printf("Received an invalid Reject{Follow}, status is %s. Ignoring. Activity: %s\n", status, report.OriginalActivity)
	}
}

func sendRejectFollow(report activities.FollowReport) {
	if !stricks.ValidURL(report.ActorID) {
		log.Printf("Invaling actor ID: %s. Dropping activity.\n", report.ActorID)
	}

	activity, err := activities.NewReject(report.OriginalActivity)
	if err = SendActivityToInbox(activity, fediverse.RequestActorInboxByID(report.ActorID)); err != nil {
		log.Println(err)
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
		log.Printf("Unexpected payload for notifyJob of type %T: %v\n", v, v)
		return
	}

	post, found := db.GetBookmarkByID(postId)
	if !found {
		log.Printf("Can't notify about non-existent repost no. %d\n", postId)
		return
	}

	if post.RepostOf == nil {
		log.Printf("Post %d is not a repost\n", postId)
		return
	}

	if post.Visibility != types.Public {
		log.Printf("Repost %d is not public, not notifying\n", postId)
		return
	}

	activity, err := activities.NewAnnounce(
		*post.RepostOf,
		fmt.Sprintf("%s/%d", settings.SiteURL(), postId),
	)
	if err != nil {
		log.Println(err)
		return
	}

	// TODO: this will have to change. Avoid sending twice if reposting a follower
	err = sendActivity(*post.RepostOf, activity)
	if err != nil {
		log.Println(err)
		return
	}

	broadcastToFollowers(jobtype.Job{
		Category: jobtype.SendAnnounce,
		Payload:  activity,
	})
}

func verifyJob(job jobtype.Job) {
	var report activities.AnnounceReport
	switch v := job.Payload.(type) {
	case []byte:
		if err := json.Unmarshal(v, &report); err != nil {
			log.Printf("While unmarshaling announce report %v: %v\n", v, err)
			return
		}
	case string:
		if err := json.Unmarshal([]byte(v), &report); err != nil {
			log.Printf("While unmarshaling announce report %v: %v\n", v, err)
			return
		}
	default:
		log.Printf("Bad payload for VerifyTheirRepost job: %v\n", v)
		return
	}

	valid, err := readpage.IsThisValidRepost(report)
	if err != nil {
		log.Printf("While parsing repost page %s: %v\n", report.RepostPage, err)
		return
	}

	if !valid {
		log.Printf("There is no repost of %s at %s\n", report.OriginalPage, report.RepostPage)
		return
	}

	// getting postId
	parts := strings.Split(report.OriginalPage, "/")
	postId, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		log.Println(err)
		return
	}

	log.Println("Saving repost")
	db.SaveRepost(postId, types.RepostInfo{
		URL:  report.RepostPage,
		Name: report.ReposterUsername,
	})
}

func receiveUnrepostJob(job jobtype.Job) {
	var report activities.UndoAnnounceReport

	switch v := job.Payload.(type) {
	case []byte:
		if err := json.Unmarshal(v, &report); err != nil {
			log.Printf("While unmarshaling UndoAnnounceReport %v: %v\n", v, err)
			return
		}
	default:
		log.Printf("Bad payload for ReceiveUnrepost job: %v\n", v)
		return
	}

	valid, err := readpage.IsThisValidRepost(report.AnnounceReport)
	if err != nil {
		log.Printf("While parsing repost page %s: %v\n", report.RepostPage, err)
		return
	}

	if valid {
		log.Printf("There is still a repost of %s at %s\n", report.OriginalPage, report.RepostPage)
		return
	}

	parts := strings.Split(report.OriginalPage, "/")
	postId, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		log.Println(err)
		return
	}

	log.Printf("Deleting %s's repost of %s at %s\n", report.ReposterUsername, report.OriginalPage, report.RepostPage)
	db.DeleteRepost(postId, report.RepostPage)
}

func notifyAboutMyUnrepost(job jobtype.Job) {
	var report activities.UndoAnnounceReport

	switch v := job.Payload.(type) {
	case []byte:
		if err := json.Unmarshal(v, &report); err != nil {
			log.Printf("While unmarshaling UndoAnnounceReport %v: %v\n", v, err)
			return
		}

		// TODO: avoid sending twice if unreposted from follower
		err := sendActivity(report.OriginalPage, v)
		if err != nil {
			log.Printf("While sending unrepost notification: %s\n", err)
		}
		broadcastToFollowers(jobtype.Job{
			Category: jobtype.SendUndoAnnounce,
			Payload:  v,
		})
	default:
		log.Printf("Bad payload for ReceiveUnrepost job: %v\n", v)
		return
	}
}
