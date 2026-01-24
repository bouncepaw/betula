// SPDX-FileCopyrightText: 2022-2025 Betula contributors
//
// SPDX-License-Identifier: AGPL-3.0-only

// Package jobs handles behind-the-scenes scheduled stuff.
//
// It makes sense to call all functions here in a separate goroutine.
package jobs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"git.sr.ht/~bouncepaw/betula/db"
	"git.sr.ht/~bouncepaw/betula/fediverse/signing"
	"git.sr.ht/~bouncepaw/betula/jobs/jobtype"
	"git.sr.ht/~bouncepaw/betula/pkg/stricks"
	"git.sr.ht/~bouncepaw/betula/settings"
	"git.sr.ht/~bouncepaw/betula/types"
)

var jobch = make(chan jobtype.Job)

var client = http.Client{
	Timeout: time.Second * 5,
}

// ScheduleDatum schedules a job with the given category and data of any type, which will be saved as is.
//
// TODO: get rid of it.
func ScheduleDatum(category jobtype.JobCategory, data any) {
	job := jobtype.Job{
		Category: category,
		Payload:  data,
	}
	id := db.PlanJob(job)
	job.ID = id
	jobch <- job
}

// ScheduleJSON schedules a job with the given category and data, which will be marshaled into JSON before saving to database. This is the one you should use, unlike ScheduleDatum.
func ScheduleJSON(category jobtype.JobCategory, dataJSON any) {
	data, err := json.Marshal(dataJSON)
	if err != nil {
		log.Printf("While scheduling a %s job: %s\n", category, err)
		return
	}
	ScheduleDatum(category, data)
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

// TODO: Move to a proper place
func SendActivityToInbox(activity []byte, inbox string) error {
	rq, err := http.NewRequest(http.MethodPost, inbox, bytes.NewReader(activity))
	if err != nil {
		log.Println(err)
		return err
	}

	rq.Header.Set("User-Agent", settings.UserAgent())
	rq.Header.Set("Content-Type", types.ActivityType)
	signing.SignRequest(rq, activity)

	log.Printf("Sending activity %s to %s\n", string(activity), inbox)
	resp, err := client.Do(rq)
	if err != nil {
		log.Println(err)
		return err
	}
	if resp.StatusCode != http.StatusOK {
		log.Printf("Activity sent to %s returned %d status\n", inbox, resp.StatusCode)
	}
	return nil
}

func SendQuietActivityToInbox(activity []byte, inbox string) error {
	rq, err := http.NewRequest(http.MethodPost, inbox, bytes.NewReader(activity))
	if err != nil {
		log.Println(err)
		return err
	}

	rq.Header.Set("Content-Type", types.ActivityType)
	signing.SignRequest(rq, activity)

	log.Printf("Sending activity to %s\n", inbox)
	resp, err := client.Do(rq)
	if err != nil {
		log.Println(err)
		return err
	}
	if resp.StatusCode != http.StatusOK {
		log.Printf("Activity sent to %s returned %d status\n", inbox, resp.StatusCode)
	}
	return nil
}

func sendActivity(uri string, activity []byte) error {
	url := stricks.ParseValidURL(uri)
	inbox := fmt.Sprintf("%s://%s/inbox", url.Scheme, url.Host)
	return SendActivityToInbox(activity, inbox)
}
