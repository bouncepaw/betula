package db

import (
	"fmt"
	"git.sr.ht/~bouncepaw/betula/types"
)

// JobNotifyAboutMyRepost schedules Betula to notify the original post creator of a new repost.
func JobNotifyAboutMyRepost(postID int) {
	if postID <= 0 {
		panic(fmt.Sprintf("Very bad post ID: %d", postID))
	}

	q := `insert into Jobs (Category, Payload) values (?, ?)`
	mustExec(q, types.NotifyAboutMyRepost, postID)
}

// JobCheckTheirRepost schedules Betula to check out the new incoming repost.
func JobCheckTheirRepost(url string) {
	q := `insert into Jobs (Category, Payload) values (?, ?)`
	mustExec(q, types.CheckTheirRepost, url)
}
