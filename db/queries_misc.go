package db

import (
	"database/sql"
	"git.sr.ht/~bouncepaw/betula/types"
	"log"
	"time"
)

func MetaEntry[T any](key BetulaMetaKey) T {
	const q = `select Value from BetulaMeta where Key = ? limit 1;`
	return querySingleValue[T](q, key)
}

func SetMetaEntry[T any](key BetulaMetaKey, val T) {
	const q = `insert or replace into BetulaMeta values (?, ?);`
	mustExec(q, key, val)
}

func OldestTime(authorized bool) *time.Time {
	const q = `
select min(CreationTime)
from Posts
where DeletionTime is null and (Visibility = 1 or ?);
`
	stamp := querySingleValue[sql.NullString](q, authorized)
	if stamp.Valid {
		val, err := time.Parse("2006-01-02 15:04:05", stamp.String)
		if err != nil {
			log.Fatalln(err)
		}
		return &val
	}
	return nil
}

func NewestTime(authorized bool) *time.Time {
	const q = `
select max(CreationTime)
from Posts
where DeletionTime is null and (Visibility = 1 or ?);
`
	stamp := querySingleValue[sql.NullString](q, authorized)
	if stamp.Valid {
		val, err := time.Parse(types.TimeLayout, stamp.String)
		if err != nil {
			log.Fatalln(err)
		}
		return &val
	}
	return nil
}
