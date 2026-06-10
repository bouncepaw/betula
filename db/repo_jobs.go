// SPDX-FileCopyrightText: 2023 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2024 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package db

import (
	"context"

	"git.sr.ht/~bouncepaw/betula/jobs/jobtype"
	jobsports "git.sr.ht/~bouncepaw/betula/ports/jobs"
)

type JobsRepo struct {
}

var _ jobsports.Repository = (*JobsRepo)(nil)

func NewJobsRepo() *JobsRepo {
	return &JobsRepo{}
}

func (repo *JobsRepo) PlanJob(ctx context.Context, job jobtype.Job) (int64, error) {
	res, err := db.ExecContext(ctx, `insert into Jobs (Category, Payload) values (?, ?)`, job.Category, job.Payload)
	if err != nil {
		return 0, err
	}

	// It never fails
	// https://github.com/mattn/go-sqlite3/blob/v1.14.17/sqlite3.go#L2008
	id, _ := res.LastInsertId()
	return id, nil
}

func (repo *JobsRepo) DropJob(ctx context.Context, id int64) error {
	_, err := db.ExecContext(ctx, `delete from Jobs where ID = ?`, id)
	return err
}

func (repo *JobsRepo) LoadAllJobs(ctx context.Context) ([]jobtype.Job, error) {
	rows, err := db.QueryContext(ctx, `select id, category, payload from Jobs`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []jobtype.Job
	for rows.Next() {
		var job jobtype.Job
		if err := rows.Scan(&job.ID, &job.Category, &job.Payload); err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	return jobs, rows.Err()
}
