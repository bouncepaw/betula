package db

import (
	"database/sql"
	"fmt"
	"log"
)

const expectedVersion = 3

/*
Wishes for schema version 4:

1.

Write more here. Implement all when there is an actual need to have a new schema.
*/

const currentSchema = `
create table Posts (
    ID integer primary key autoincrement,
    URL text not null check ( URL <> '' ),
    Title text not null check ( Title <> '' ),
    Description text not null,
    Visibility integer check ( Visibility = 0 or Visibility = 1 ),
    CreationTime text not null default current_timestamp,
    DeletionTime text                
);

create table TagsToPosts (
    TagName text not null,
    PostID integer not null,
    unique (Tag, PostID) on conflict ignore,
    check ( TagName <> '' )
);

create table BetulaMeta (
    Key text primary key not null,
    Value text
);

insert or ignore into BetulaMeta values
	('DB version', 3),
	('Admin username', null),
	('Admin password hash', null);

create table Sessions (
    Token text primary key not null,
    CreationTime not null default current_timestamp
);

create table TagDescriptions (
   TagName text primary key,
   Description text not null
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
	case 2:
		migrate2To3()
	case 1:
		migrate1To2()
		migrate2To3()
	case 0:
		migrate0To1()
		migrate1To2()
		migrate2To3()
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

func migrate2To3() {
	log.Println("Migrating from 2 to 3...")
	/* Past is as such
	create table CategoriesToPosts (
	    CatName text not null,
	    PostID integer not null,
	    unique (CatName, PostID) on conflict ignore,
	    check ( CatName <> '' )
	);

	create table CategoryDescriptions (
	   CatName text primary key,
	   Description text not null
	);
	*/
	const q = `
-- It's not like we needed it in the first place
drop table CategoryImplications;

-- Rename CategoriesToPosts to TagsToPosts
create table TagsToPosts (
    TagName text not null,
    PostID integer not null,
    unique (TagName, PostID) on conflict ignore,
    check ( TagName <> '' )
);

insert into TagsToPosts (TagName, PostID)
select CatName, PostID
from CategoriesToPosts;

drop table CategoriesToPosts;


-- Rename CategoryDescriptions to TagDescriptions
create table TagDescriptions (
   TagName text primary key,
   Description text not null
);

insert into TagDescriptions (TagName, Description)
select CatName, Description
from CategoryDescriptions;

drop table CategoryDescriptions;


--- Taking notes...
replace into BetulaMeta (Key, Value) values ('DB version', 3);
`
	mustExec(q)

	log.Println("Migrated from 2 to 3")
}

func migrate1To2() {
	log.Println("Migrating from 1 to 2...")
	/*-- Past is as such:
	create table Posts (
		ID integer primary key autoincrement,
		URL text not null check ( URL <> '' ),
		Title text not null check ( Title <> '' ),
		Description text not null,
		Visibility integer check ( Visibility = 0 or Visibility = 1 ),
		CreationTime integer not null default (strftime('%s', 'now')),
		DeletionTime integer
	);
	create table Sessions (
		Token text primary key not null,
		CreationTime integer not null
	);
	*/
	const q = `
-- New tables
create table CategoryDescriptions (
   CatName text primary key,
   Description text not null
);

create table CategoryImplications (
   IfCat text not null,
   ThenCat text not null
);

-- Going from UNIX time to stringular time
--- Posts
create table NewPosts (
	ID integer primary key autoincrement,
	URL text not null check ( URL <> '' ),
	Title text not null check ( Title <> '' ),
	Description text not null,
	Visibility integer check ( Visibility = 0 or Visibility = 1 ),
	CreationTime text not null default current_timestamp,
	DeletionTime text
);

insert into NewPosts (ID, URL, Title, Description, Visibility, CreationTime, DeletionTime)
select ID, URL, Title, Description, Visibility,
   datetime(CreationTime, 'unixepoch'), datetime(DeletionTime, 'unixepoch')
from Posts;

drop table Posts;
alter table NewPosts rename to Posts;

--- Sessions
create table NewSessions (
    Token text primary key not null,
    CreationTime text not null default current_timestamp
);

insert into NewSessions (Token, CreationTime)
select Token, datetime(CreationTime, 'unixepoch')
from Sessions;

drop table Sessions;
alter table NewSessions rename to Sessions;

--- Taking notes...
replace into BetulaMeta (Key, Value) values ('DB version', 2);
`
	mustExec(q)
	log.Println("Migrated from 1 to 2")
}

func migrate0To1() {
	log.Println("Migrating from 0 to 1...")
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
    CreationTime integer not null default (strftime('%s', 'now')),
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
