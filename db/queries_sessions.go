// SPDX-FileCopyrightText: 2022-2025 Betula contributors
//
// SPDX-License-Identifier: AGPL-3.0-only

package db

import (
	"git.sr.ht/~bouncepaw/betula/tools"
	ua "github.com/mileusna/useragent"
	"time"

	"git.sr.ht/~bouncepaw/betula/types"
)

func AddSession(token, userAgent string) {
	mustExec(`insert into Sessions(Token, UserAgent) values (?, ?);`, token, userAgent)
}

func SessionExists(token string) (has bool) {
	rows := mustQuery(`select exists(select 1 from Sessions where Token = ?);`, token)
	rows.Next()
	mustScan(rows, &has)
	_ = rows.Close()
	return has
}

func StopSession(token string) {
	mustExec(`delete from Sessions where Token = ?;`, token)
}

func StopAllSessions(excludeToken string) {
	mustExec(`delete from Sessions where Token <> ?;`, excludeToken)
}

func SetCredentials(name, hash string) {
	mustExec(`
insert or replace into BetulaMeta values
	('Admin username', ?),
	('Admin password hash', ?);
`, name, hash)
}

func Sessions() (sessions []types.Session) {
	rows := mustQuery(`select Token, CreationTime, coalesce(UserAgent, '') from Sessions`)
	for rows.Next() {
		var err error
		var timestamp string
		var creationTime time.Time
		var session types.Session
		var userAgent string

		mustScan(rows, &session.Token, &timestamp, &userAgent)
		session.UserAgent = ua.Parse(userAgent)
		creationTime, err = time.Parse(types.TimeLayout, timestamp)
		if err != nil {
			creationTime, err = time.Parse(types.TimeLayout+"Z07:00", timestamp)
			if err != nil {
				continue
			}
		}
		session.LastSeen = tools.LastSeen(creationTime, time.Now())
		sessions = append(sessions, session)
	}
	return sessions
}
