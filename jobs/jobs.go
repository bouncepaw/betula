// Package jobs handles behind-the-scenes scheduled stuff
package jobs

import "git.sr.ht/~bouncepaw/betula/db"

func ListenAndWhisper() {
	for {
		select {
		case rowid := <-db.JobChannel:
			_ = rowid
		}
	}
}
