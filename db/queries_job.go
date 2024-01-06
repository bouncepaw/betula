package db

import (
	"git.sr.ht/~bouncepaw/betula/jobs/jobtype"
	"log"
)

// PlanJob puts a new job into the Jobs table and returns the id of the new job.
func PlanJob(job jobtype.Job) int64 {
	const q = `insert into Jobs (Category, Payload) values (?, ?)`

	// mustExec not used because res needed
	res, err := db.Exec(q, job.Category, job.Payload)
	if err != nil {
		log.Fatalln(err)
	}

	// It never fails
	// https://github.com/mattn/go-sqlite3/blob/v1.14.17/sqlite3.go#L2008
	id, _ := res.LastInsertId()
	return id
}

// DropJob removes the job specified by id from the database.
// Call after the job is done.
func DropJob(id int64) {
	mustExec(`delete from Jobs where ID = ?`, id)
}

// LoadAllJobs reads all jobs in the database. Call on startup once.
func LoadAllJobs() (jobs []jobtype.Job) {
	const q = `select id, category, payload from Jobs`
	rows := mustQuery(q)
	for rows.Next() {
		var job jobtype.Job
		mustScan(rows, &job.ID, &job.Category, &job.Payload)
		jobs = append(jobs, job)
	}
	return jobs
}
