// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package jobsports

import (
	"context"

	"git.sr.ht/~bouncepaw/betula/jobs/jobtype"
)

type Repository interface {
	// PlanJob puts a new job into the database and returns the id of the new job.
	PlanJob(ctx context.Context, job jobtype.Job) (int64, error)
	// DropJob removes the job specified by id from the database.
	// Call after the job is done.
	DropJob(ctx context.Context, id int64) error
	// LoadAllJobs reads all jobs in the database. Call on startup once.
	LoadAllJobs(ctx context.Context) ([]jobtype.Job, error)
}
