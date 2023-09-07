// Package jobs handles behind-the-scenes scheduled stuff.
//
// It makes sense to call all functions here in a separate goroutine.
package jobs

import (
	"git.sr.ht/~bouncepaw/betula/db"
	"git.sr.ht/~bouncepaw/betula/types"
	"log"
)

var jobch = make(chan types.Job)

func ListenAndWhisper() {
	lateJobs := db.LoadAllJobs()
	go func() {
		for job := range jobch {
			log.Printf("Received job no. %d ‘%s: %v’\n", job.ID, job.Category, job.Payload)
			switch job.Category {
			case types.NotifyAboutMyRepost:
				// TODO: handle job
			case types.VerifyTheirRepost:
				// TODO: handle job
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

// CheckThisRepostLater plans a job to check if the repost at iri is fine.
// The iri is expected to be valid.
func CheckThisRepostLater(iri string) {
	job := types.Job{
		Category: types.VerifyTheirRepost,
		Payload:  iri,
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
