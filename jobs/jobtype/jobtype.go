// Package jobtype holds types for jobs and their categories.
package jobtype

import "time"

// If you make something drastic to this file, reflect the changes in Adding a new job.md

type JobCategory string

const (
	SendAnnounce        JobCategory = "notify about my repost"
	ReceiveAnnounce     JobCategory = "verify their repost"
	ReceiveUndoAnnounce JobCategory = "receive unrepost"
	SendUndoAnnounce    JobCategory = "notify about my unrepost"

	/* I changed the style from now. The new style is below. */

	SendAcceptFollow    JobCategory = "Send Accept{Follow}"
	SendRejectFollow    JobCategory = "Send Reject{Follow}"
	ReceiveAcceptFollow JobCategory = "Receive Accept{Follow}"
	ReceiveRejectFollow JobCategory = "Receive Reject{Follow}"
	SendCreateNote      JobCategory = "Send Create{Note}"
	SendUpdateNote      JobCategory = "Send Update{Note}"
	SendDeleteNote      JobCategory = "Send Delete{Note}"
	ReceiveCreateNote   JobCategory = "Receive Create{Note}"
	ReceiveUpdateNote   JobCategory = "Receive Update{Note}"
	ReceiveDeleteNote   JobCategory = "Receive Delete{Note}"
)

// Job is a task for Betula to do later.
type Job struct {
	// ID is a unique identifier for the Job. You get it when reading from the database. Do not set it when issuing a new job.
	ID       int64
	Category JobCategory
	Due      time.Time
	// Payload is some data.
	Payload any
}
