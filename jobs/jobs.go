// Package jobs handles behind-the-scenes scheduled stuff.
//
// It makes sense to call all functions here in a separate goroutine.
package jobs

import (
	"bytes"
	"fmt"
	"git.sr.ht/~bouncepaw/betula/db"
	"git.sr.ht/~bouncepaw/betula/signing"
	"git.sr.ht/~bouncepaw/betula/stricks"
	"git.sr.ht/~bouncepaw/betula/types"
	"log"
	"net/http"
	"time"
)

var jobch = make(chan types.Job)

var client = http.Client{
	Timeout: time.Second,
}

var catmap = map[types.JobCategory]func(job types.Job){
	types.NotifyAboutMyRepost:   notifyJob,
	types.VerifyTheirRepost:     verifyJob,
	types.ReceiveUnrepost:       receiveUnrepostJob,
	types.NotifyAboutMyUnrepost: notifyAboutMyUnrepost,
}

func ListenAndWhisper() {
	lateJobs := db.LoadAllJobs()
	go func() {
		for job := range jobch {
			log.Printf("Received job no. %d ‘%s’\n", job.ID, job.Category)
			if jobber, ok := catmap[job.Category]; !ok {
				fmt.Printf("An unhandled job category came in: %s\n", job.Category)
			} else {
				jobber(job)
			}
			db.DropJob(job.ID)
		}
	}()
	for _, job := range lateJobs {
		jobch <- job
	}
}

func SendActivityToInbox(activity []byte, inbox string) error {
	rq, err := http.NewRequest(http.MethodPost, inbox, bytes.NewReader(activity))
	if err != nil {
		log.Println(err)
		return err
	}

	rq.Header.Set("Content-Type", types.ActivityType)
	signing.SignRequest(rq, activity)

	log.Printf("Sending activity %s\n", string(activity))
	resp, err := client.Do(rq)
	if err != nil {
		log.Println(err)
		return err
	}
	log.Printf("Activity sent to %s returned %d status\n", inbox, resp.StatusCode)
	return nil
}

func sendActivity(uri string, activity []byte) error {
	url := stricks.ParseValidURL(uri)
	inbox := fmt.Sprintf("%s://%s/inbox", url.Scheme, url.Host)
	return SendActivityToInbox(activity, inbox)
}

func planJob(category types.JobCategory, data any) {
	job := types.Job{
		Category: category,
		Payload:  data,
	}
	id := db.PlanJob(job)
	job.ID = id
	jobch <- job
}
