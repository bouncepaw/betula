// Package jobs handles behind-the-scenes scheduled stuff
package jobs

import (
	"git.sr.ht/~bouncepaw/betula/db"
	"log"
	"net/url"
)

func ListenAndWhisper() {
	for {
		select {
		case rowid := <-db.JobChannel:
			_ = rowid
		}
	}
}

func CheckThisRepostLater(iri *url.URL) {
	log.Printf("Look what we got! %s\n", iri.String())
}

func NotifyAboutMyRepost(postId int) {

}
