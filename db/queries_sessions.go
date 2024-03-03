package db

import "time"

func AddSession(token string) {
	mustExec(`insert into Sessions values (?, ?);`,
		token, time.Now())
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

func SetCredentials(name, hash string) {
	mustExec(`
insert or replace into BetulaMeta values
	('Admin username', ?),
	('Admin password hash', ?);
`, name, hash)
}
