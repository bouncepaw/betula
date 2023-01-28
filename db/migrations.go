package db

import (
	"database/sql"
	"fmt"
	"log"
)

const expectedVersion = 1

const currentSchema = `
create table Posts (
    ID integer primary key autoincrement,
    URL text not null check ( URL <> '' ),
    Title text not null check ( Title <> '' ),
    Description text not null,
    Visibility integer check ( Visibility = 0 or Visibility = 1 ),
    CreationTime integer not null default current_timestamp,
    DeletionTime integer                
);

create view Categories as
select distinct CatName from CategoriesToPosts;

create table CategoriesToPosts (
    CatName text not null,
    PostID integer not null,
    unique (CatName, PostID) on conflict ignore,
    check ( CatName <> '' )
);

create table BetulaMeta (
    Key text primary key not null,
    Value text
);

insert or ignore into BetulaMeta values
	('DB version', 1),
	('Admin username', null),
	('Admin password hash', null);

create table Sessions (
    Token text primary key not null,
    CreationTime integer not null
);`

func handleMigrations() {
	curver, found := currentVersion()
	if !found {
		mustExec(currentSchema)
		return
	}

	if curver == expectedVersion {
		return
	}

	if curver > expectedVersion {
		log.Fatalf("The database file specifies version %d, but this version of Betula only supports versions up to %d. Please update Betula or fix your database somehow.\n", curver, expectedVersion)
	}

	switch curver {
	case 0:
		migrate0To1()
	default:
		panic(fmt.Sprintf("unimplemented migration from %d to %d", curver, expectedVersion))
	}
}

func currentVersion() (version int64, found bool) {
	const qMetaExists = `
select name from sqlite_master
where type='table' and name='BetulaMeta' limit 1;
`
	name := querySingleValue[sql.NullString](qMetaExists)
	if !name.Valid {
		return 0, false
	}

	const qVersion = `
select Value from BetulaMeta
where Key = 'DB version';
`
	v := querySingleValue[sql.NullInt64](qVersion)
	if !v.Valid {
		return 0, false
	}

	return v.Int64, true
}

func migrate0To1() {
	log.Println("Migrating from 0 to 1")
	/*
		--This was the definition in the past:
		create table if not exists Posts (
			 ID integer primary key autoincrement not null,
			 URL text not null,
			 Title text not null,
			 Description text not null,
			 Visibility integer not null,
			 CreationTime integer not null
		);
	*/
	// See https://www.sqlite.org/lang_altertable.html for recommendations.
	// Especially section no. 7. They suggest this order:
	// 1. Create new table
	// 2. Copy data
	// 3. Drop old table
	// 4. Rename new into old
	// So be it!
	const q = `
create table NewPosts (
    ID integer primary key autoincrement,
    URL text not null check ( URL <> '' ),
    Title text not null check ( Title <> '' ),
    Description text not null,
    Visibility integer check ( Visibility = 0 or Visibility = 1 ),
    CreationTime integer not null default current_timestamp,
    DeletionTime integer                 
);

insert into NewPosts (ID, URL, Title, Description, Visibility, CreationTime)
select ID, URL, Title, Description, Visibility, CreationTime
from Posts;

drop table Posts;

alter table NewPosts rename to Posts;

replace into BetulaMeta (Key, Value) values ('DB version', 1);
`
	mustExec(q)

	log.Println("Migrated from 0 to 1")
}
