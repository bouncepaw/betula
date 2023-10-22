// Package jobs handles behind-the-scenes scheduled stuff.
//
// It makes sense to call all functions here in a separate goroutine.
package jobs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"git.sr.ht/~bouncepaw/betula/activities"
	"git.sr.ht/~bouncepaw/betula/db"
	"git.sr.ht/~bouncepaw/betula/readpage"
	"git.sr.ht/~bouncepaw/betula/settings"
	"git.sr.ht/~bouncepaw/betula/stricks"
	"git.sr.ht/~bouncepaw/betula/types"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var jobch = make(chan types.Job)

var client = http.Client{
	Timeout: time.Second,
}

func sendActivity(uri string, activity []byte) error {
	url := stricks.ParseValidURL(uri)
	inbox := fmt.Sprintf("%s://%s/inbox", url.Scheme, url.Host)
	resp, err := client.Post(
		inbox,
		"application/ld+json; profile=\"https://www.w3.org/ns/activitystreams\"",
		bytes.NewReader(activity),
	)
	if err != nil {
		log.Println(err)
		return err
	}
	log.Printf("Sending activity %s\n", string(activity))
	log.Printf("Activity sent to %s returned %d status\n", inbox, resp.StatusCode)
	return nil
}

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

func ListenAndWhisper() {
	lateJobs := db.LoadAllJobs()
	go func() {
		for job := range jobch {
			log.Printf("Received job no. %d ‘%s’\n", job.ID, job.Category)
			switch job.Category {
			case types.NotifyAboutMyRepost:
				notifyJob(job)
			case types.VerifyTheirRepost:
				verifyJob(job)
			case types.ReceiveUnrepost:
				receiveUnrepostJob(job)
			default:
				panic("unhandled job type")
			}
			db.DropJob(job.ID)
		}
	}()
	for _, job := range lateJobs {
		jobch <- job
	}
}

// CheckThisRepostLater plans a job to check the specified announce if it's true.
func CheckThisRepostLater(announce activities.AnnounceReport) {
	data, err := json.Marshal(announce)
	if err != nil {
		log.Printf("While scheduling repost checking: %v\n", err)
		return
	}
	job := types.Job{
		Category: types.VerifyTheirRepost,
		Payload:  data,
	}
	id := db.PlanJob(job)
	job.ID = id
	jobch <- job
}

func NotifyAboutMyRepost(postId int64) {
	job := types.Job{
		Category: types.NotifyAboutMyRepost,
		Payload:  postId,
	}
	id := db.PlanJob(job)
	job.ID = id
	jobch <- job
}

func ReceiveUnrepost(report activities.UndoAnnounceReport) {
	data, err := json.Marshal(report)
	if err != nil {
		log.Printf("While scheduling unrepost checking: %v\n", err)
		return
	}
	job := types.Job{
		Category: types.ReceiveUnrepost,
		Payload:  data,
	}
	id := db.PlanJob(job)
	job.ID = id
	jobch <- job
}
