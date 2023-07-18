module git.sr.ht/~bouncepaw/betula

go 1.19

require (
	git.sr.ht/~bouncepaw/mycomarkup/v5 v5.4.0
	github.com/gorilla/feeds v1.1.1
	github.com/mattn/go-sqlite3 v1.14.16
	golang.org/x/crypto v0.5.0
	golang.org/x/net v0.5.0
)

require github.com/kr/pretty v0.3.1 // indirect

// Temporary fix for musl! We use musl on builds.sr.ht.
// Should be removed when https://github.com/mattn/go-sqlite3/pull/1177 gets merged
// or https://github.com/mattn/go-sqlite3/pull/1177 gets fixed in any way.
replace github.com/mattn/go-sqlite3 => github.com/leso-kn/go-sqlite3 v0.0.0-20230710125852-03158dc838ed
