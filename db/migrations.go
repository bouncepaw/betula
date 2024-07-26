package db

import (
	"cmp"
	"database/sql"
	"embed"
	"fmt"
	"log"
	"slices"
	"strconv"
	"strings"
)

//go:embed scripts/*.sql
var scriptsFS embed.FS
var expectedVersion int64

func init() {
	expectedVersion = getExpectedVersion()
}

func getExpectedVersion() int64 {
	var version int64

	scriptsDir, err := scriptsFS.ReadDir("scripts")
	if err != nil {
		log.Fatalln(err)
	}

	scripts := make([]string, 0, len(scriptsDir))
	for _, script := range scriptsDir {
		scripts = append(scripts, script.Name())
	}

	getVersionNum := func(scriptName string) (int64, error) {
		return strconv.ParseInt(strings.TrimSuffix(scriptName, ".sql"), 10, 64)
	}

	slices.SortFunc(scripts, func(i, j string) int {
		verNum1, err1 := getVersionNum(i)
		verNum2, err2 := getVersionNum(j)
		if err1 != nil || err2 != nil {
			log.Fatalln(err1, err2)
		}
		return cmp.Compare(verNum1, verNum2)
	})

	// Get last version
	version, err = getVersionNum(scripts[len(scripts)-1])
	if err != nil {
		log.Fatalln(err)
	}

	return version
}

func getScript(name string) string {
	data, err := scriptsFS.ReadFile("scripts/" + name + ".sql")
	if err != nil {
		log.Fatalln(err)
	}
	return string(data)
}

// Never update the SQL here. Add comments maybe.
const schemaV6 string = `
create table Posts (
    ID integer primary key autoincrement,
    URL text not null check ( URL <> '' ),
    Title text not null check ( Title <> '' ),
    Description text not null,
    Visibility integer check ( Visibility = 0 or Visibility = 1 ),
    CreationTime text not null default current_timestamp,
    DeletionTime text,
    RepostOf text
);

create table TagsToPosts (
    TagName text not null,
    PostID integer not null,
    unique (TagName, PostID) on conflict ignore,
    check ( TagName <> '' )
);

create table BetulaMeta (
    Key text primary key not null,
    Value text
);

insert or ignore into BetulaMeta values
	('DB version', 6),
	('Admin username', null),
	('Admin password hash', null);

create table Sessions (
    Token text primary key not null,
    CreationTime not null default current_timestamp
);

create table TagDescriptions (
   TagName text primary key,
   Description text not null
);

create table Jobs (
	ID integer primary key autoincrement,
	Due not null default current_timestamp,
	Category text not null,
	Payload
);

create table KnownReposts (
   RepostURL text primary key,
   ReposterName text,
   RepostedAt text not null default current_timestamp,
   PostID integer
);
`

func handleMigrations() {
	curver, found := currentVersion()

	// DB was never populated! Let's write the latest schema we have.
	if !found {
		mustExec(schemaV6) // Up to 6
		curver = 6
		goto past6 // And newer
	}

	// Seems to be update, we're done here.
	if curver == expectedVersion {
		return
	}

	// Whoa, a db from a newer Betula? We better get out of here.
	if curver > expectedVersion {
		log.Fatalf("The database file specifies database version %d, but this version of Betula only supports database versions up to %d. Please update your Betula.\n", curver, expectedVersion)
	}

	// Here, curver < expectedVersion

	// Migration up to 6 were made using a different mechanism. Let's get up to 6 first.
	if curver < 6 {
		migrators := []func(){migrate0To1, migrate1To2, migrate2To3, migrate3To4, migrate4To5, migrate5To6}
		for _, migrator := range migrators[curver:] {
			log.Printf("Migrating from DB schema version %d to %d...\n", curver, curver+1)
			migrator()
			curver++
		}
	}

past6:
	for curver < expectedVersion {
		if found {
			log.Printf("Migrating from DB schema version %d to %d...\n", curver, curver+1)
		}
		mustExec(getScript(fmt.Sprintf("%d", curver+1)))
		mustExec(fmt.Sprintf(`replace into BetulaMeta (Key, Value) values ('DB version', %d);`, curver+1))
		curver++
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

func migrate5To6() {
	/*
		-- Old
		create table KnownReposts (
		   RepostURL text primary key,
		   PostID integer
		);
	*/
	const q = `
alter table KnownReposts add column ReposterName text;
alter table KnownReposts add column RepostedAt text not null default current_timestamp;

replace into BetulaMeta (Key, Value) values ('DB version', 6);
`
	mustExec(q)
}

func migrate4To5() {
	/* -- Old jobs
	create table Jobs (
		ID integer primary key autoincrement,
		Category text not null,
		Payload
	);
	*/
	const q = `
alter table Jobs add column Due not null default current_timestamp;

replace into BetulaMeta (Key, Value) values ('DB version', 5);
`
	mustExec(q)
}

func migrate3To4() {
	const q = `
-- The new tables.
create table Jobs (
	ID integer primary key autoincrement,
	Category text not null,
	Payload
);

create table KnownReposts (
   RepostURL text primary key,
   PostID integer
);

-- New Posts!
alter table Posts add column RepostOf text;

-- Taking notes...
replace into BetulaMeta (Key, Value) values ('DB version', 4);
`
	/* The past was as such:
	create table Posts (
	    ID integer primary key autoincrement,
	    URL text not null check ( URL <> '' ),
	    Title text not null check ( Title <> '' ),
	    Description text not null,
	    Visibility integer check ( Visibility = 0 or Visibility = 1 ),
	    CreationTime text not null default current_timestamp,
	    DeletionTime text
	);
	*/
	mustExec(q)
}

func migrate2To3() {
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
}

func migrate1To2() {
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
}

func migrate0To1() {
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
}
