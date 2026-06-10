// SPDX-FileCopyrightText: 2023 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2024 Danila Gorelko
// SPDX-FileCopyrightText: 2024 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package db

import (
	"context"
	"time"

	ua "github.com/mileusna/useragent"

	"git.sr.ht/~bouncepaw/betula/pkg/bxtime"
	sessionsports "git.sr.ht/~bouncepaw/betula/ports/sessions"
	"git.sr.ht/~bouncepaw/betula/types"
)

type SessionsRepo struct {
}

var _ sessionsports.Repository = (*SessionsRepo)(nil)

func NewSessionsRepo() *SessionsRepo {
	return &SessionsRepo{}
}

func (repo *SessionsRepo) AddSession(ctx context.Context, token, userAgent string) error {
	_, err := db.ExecContext(ctx, `insert into Sessions(Token, UserAgent) values (?, ?);`, token, userAgent)
	return err
}

func (repo *SessionsRepo) SessionExists(ctx context.Context, token string) (bool, error) {
	var has bool
	err := db.QueryRowContext(ctx, `select exists(select 1 from Sessions where Token = ?);`, token).Scan(&has)
	return has, err
}

func (repo *SessionsRepo) StopSession(ctx context.Context, token string) error {
	_, err := db.ExecContext(ctx, `delete from Sessions where Token = ?;`, token)
	return err
}

func (repo *SessionsRepo) StopAllSessions(ctx context.Context, excludeToken string) error {
	_, err := db.ExecContext(ctx, `delete from Sessions where Token <> ?;`, excludeToken)
	return err
}

func (repo *SessionsRepo) Sessions(ctx context.Context) ([]types.Session, error) {
	rows, err := db.QueryContext(ctx, `select Token, CreationTime, coalesce(UserAgent, '') from Sessions`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []types.Session
	for rows.Next() {
		var (
			timestamp string
			userAgent string
			session   types.Session
		)
		if err := rows.Scan(&session.Token, &timestamp, &userAgent); err != nil {
			return nil, err
		}

		session.UserAgent = ua.Parse(userAgent)
		creationTime, err := time.Parse(types.TimeLayout, timestamp)
		if err != nil {
			creationTime, err = time.Parse(types.TimeLayout+"Z07:00", timestamp)
			if err != nil {
				continue
			}
		}
		session.LastSeen = bxtime.LastSeen(creationTime, time.Now())
		sessions = append(sessions, session)
	}
	return sessions, rows.Err()
}
