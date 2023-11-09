package jobs

import (
	"encoding/json"
	"git.sr.ht/~bouncepaw/betula/activities"
	"git.sr.ht/~bouncepaw/betula/types"
	"log"
)

// CheckThisRepostLater plans a job to check the specified announce if it's true.
func CheckThisRepostLater(announce activities.AnnounceReport) {
	data, err := json.Marshal(announce)
	if err != nil {
		log.Printf("While scheduling repost checking: %v\n", err)
		return
	}
	planJob(types.VerifyTheirRepost, data)
}

func NotifyAboutMyRepost(postId int64) {
	planJob(types.NotifyAboutMyRepost, postId)
}

func ReceiveUnrepost(report activities.UndoAnnounceReport) {
	data, err := json.Marshal(report)
	if err != nil {
		log.Printf("While scheduling unrepost checking: %v\n", err)
		return
	}
	planJob(types.ReceiveUnrepost, data)
}

func NotifyAboutMyUnrepost(report activities.UndoAnnounceReport) {
	data, err := json.Marshal(report)
	if err != nil {
		log.Printf("While scheduling repost cancel notification: %v\n", err)
		return
	}
	planJob(types.NotifyAboutMyUnrepost, data)
}
