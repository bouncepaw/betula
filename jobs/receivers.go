package jobs

import (
	"encoding/json"
	"fmt"
	"git.sr.ht/~bouncepaw/betula/db"
	"git.sr.ht/~bouncepaw/betula/fediverse/activities"
	"git.sr.ht/~bouncepaw/betula/jobs/jobtype"
	"git.sr.ht/~bouncepaw/betula/readpage"
	"git.sr.ht/~bouncepaw/betula/settings"
	"git.sr.ht/~bouncepaw/betula/stricks"
	"git.sr.ht/~bouncepaw/betula/types"
	"log"
	"strconv"
	"strings"
)

var catmap = map[jobtype.JobCategory]func(job jobtype.Job){
	jobtype.SendAnnounce:        notifyJob,
	jobtype.ReceiveAnnounce:     verifyJob,
	jobtype.ReceiveUndoAnnounce: receiveUnrepostJob,
	jobtype.SendUndoAnnounce:    notifyAboutMyUnrepost,
	jobtype.SendAcceptFollow:    sendAcceptFollow,
	jobtype.SendRejectFollow:    sendRejectFollow,
	jobtype.ReceiveAcceptFollow: receiveAcceptFollow,
	jobtype.ReceiveRejectFollow: receiveRejectFollow,
}

func sendAcceptFollow(job jobtype.Job) {
	data, ok := job.Payload.([]byte)
	if !ok {
		log.Printf("Unexpected payload for NotifyAboutMyRepost of type %T: %v\n", job.Payload, job.Payload)
		return
	}

	var report activities.FollowReport
	err := json.Unmarshal(data, &report)
	if err != nil {
		log.Printf("When unmarshaling payload for job: %s\n", err)
		return
	}

	if !stricks.ValidURL(report.ActorID) {
		log.Printf("Invaling actor ID: %s. Dropping activity.\n", report.ActorID)
	}

	db.AddFollower(report.ObjectID)
	finish
}

func notifyJob(job jobtype.Job) {
	var postId int
	switch v := job.Payload.(type) {
	case int64:
		postId = int(v)
	default:
		log.Printf("Unexpected payload for notifyJob of type %T: %v\n", v, v)
		return
	}

	post, found := db.PostForID(postId)
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

	err = sendActivity(*post.RepostOf, activity)
	if err != nil {
		log.Println(err)
		return
	}
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
		err := sendActivity(report.OriginalPage, v)
		if err != nil {
			log.Printf("While sending unrepost notification: %s\n", err)
		}
	default:
		log.Printf("Bad payload for ReceiveUnrepost job: %v\n", v)
		return
	}
}
