package jobs

import (
	"encoding/json"
	"fmt"
	"git.sr.ht/~bouncepaw/betula/activities"
	"git.sr.ht/~bouncepaw/betula/db"
	"git.sr.ht/~bouncepaw/betula/readpage"
	"git.sr.ht/~bouncepaw/betula/settings"
	"git.sr.ht/~bouncepaw/betula/types"
	"log"
	"strconv"
	"strings"
)

func notifyJob(job types.Job) {
	var postId int
	switch v := job.Payload.(type) {
	case int64:
		postId = int(v)
	default:
		log.Printf("Unexpected payload for NotifyAboutMyRepost of type %T: %v\n", v, v)
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

func verifyJob(job types.Job) {
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

func receiveUnrepostJob(job types.Job) {
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

func notifyAboutMyUnrepost(job types.Job) {
	panic("todo")
	// TODO: implement
}
