package db

import (
	"time"

	"git.sr.ht/~bouncepaw/betula/types"
)

func AddSession(token, userAgent string) {
	mustExec(`insert into Sessions values (?, ?, ?);`,
		token, time.Now(), userAgent)
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
		var session types.Session

		mustScan(rows, &session.Token, &timestamp, &session.UserAgent)
		session.CreationTime, err = time.Parse(types.TimeLayout+"Z07:00", timestamp)
		if err != nil {
			continue
		}
		sessions = append(sessions, session)
	}
	return sessions
}
