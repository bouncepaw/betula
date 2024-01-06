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

func SendAcceptFollow(report activities.FollowReport) {
	data, err := json.Marshal(report)
	if err != nil {
		log.Printf("While scheduling %s: %v\n", types.SendAcceptFollow, err)
		return
	}
	planJob(types.SendAcceptFollow, data)
}

func SendRejectFollow(report activities.FollowReport) {
	data, err := json.Marshal(report)
	if err != nil {
		log.Printf("While scheduling %s: %v\n", types.SendRejectFollow, err)
		return
	}
	planJob(types.SendRejectFollow, data)
}

func ReceiveAcceptFollow(report activities.FollowReport) {

}

func ReceiveReceiveFollow(report activities.FollowReport) {

}
