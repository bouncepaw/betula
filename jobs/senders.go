package jobs

import (
	"encoding/json"
	"git.sr.ht/~bouncepaw/betula/activities"
	"git.sr.ht/~bouncepaw/betula/jobs/jobtype"
	"log"
)

// CheckThisRepostLater plans a job to check the specified announce if it's true.
func CheckThisRepostLater(announce activities.AnnounceReport) {
	data, err := json.Marshal(announce)
	if err != nil {
		log.Printf("While scheduling repost checking: %v\n", err)
		return
	}
	planJob(jobtype.ReceiveAnnounce, data)
}

func NotifyAboutMyRepost(postId int64) {
	planJob(jobtype.SendAnnounce, postId)
}

func ReceiveUnrepost(report activities.UndoAnnounceReport) {
	data, err := json.Marshal(report)
	if err != nil {
		log.Printf("While scheduling unrepost checking: %v\n", err)
		return
	}
	planJob(jobtype.ReceiveUndoAnnounce, data)
}

func NotifyAboutMyUnrepost(report activities.UndoAnnounceReport) {
	data, err := json.Marshal(report)
	if err != nil {
		log.Printf("While scheduling repost cancel notification: %v\n", err)
		return
	}
	planJob(jobtype.SendUndoAnnounce, data)
}

func SendAcceptFollow(report activities.FollowReport) {
	data, err := json.Marshal(report)
	if err != nil {
		log.Printf("While scheduling %s: %v\n", jobtype.SendAcceptFollow, err)
		return
	}
	planJob(jobtype.SendAcceptFollow, data)
}

func SendRejectFollow(report activities.FollowReport) {
	data, err := json.Marshal(report)
	if err != nil {
		log.Printf("While scheduling %s: %v\n", jobtype.SendRejectFollow, err)
		return
	}
	planJob(jobtype.SendRejectFollow, data)
}

func ReceiveAcceptFollow(report activities.FollowReport) {

}

func ReceiveReceiveFollow(report activities.FollowReport) {

}
