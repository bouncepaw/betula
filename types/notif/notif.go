// SPDX-FileCopyrightText: 2025 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

// Package notiftypes provides types that represent
// different kinds of user notifications.
package notiftypes

import (
	"database/sql"
	"encoding/json"
	"html/template"
	"time"

	"git.sr.ht/~bouncepaw/betula/types"
)

type (
	ID   int64
	Kind string

	LikePayload struct {
		ActorID    string `json:"actor_id"`
		BookmarkID int    `json:"bookmark_id"`
	}
	RemarkPayload struct {
		ActorID    string `json:"actor_id"`
		BookmarkID int    `json:"bookmark_id"`
		RemarkURL  string `json:"remark_url"`

		Source          sql.NullString   `json:"source"`
		SourceType      types.SourceType `json:"source_type"`
		DescriptionHTML template.HTML    `json:"description_html"`
	}
	FollowPayload struct {
		ActorID string `json:"actor_id"`
	}

	Notification struct {
		ID        ID
		Kind      Kind
		CreatedAt time.Time
		Payload   json.RawMessage
	}
	NotificationGroup struct {
		Title         string
		Notifications []Notification
	}
)

const (
	KindLike   Kind = "like"
	KindRemark Kind = "remark"
	KindFollow Kind = "follow"
)
