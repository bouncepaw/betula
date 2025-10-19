// Package notiftypes provides types that represent different kinds of user notifications.
package notiftypes

import (
	"encoding/json"
	"errors"
	"html/template"
	"log/slog"
	"time"
)

var (
	ErrWrongKind = errors.New("wrong kind")
)

type (
	ID   int64
	Kind string

	LikePayload struct {
		ActorID    string `json:"actor_id"`
		BookmarkID int    `json:"bookmark_id"`
	}
	RemarkPayload struct {
		ActorID    string        `json:"actor_id"`
		BookmarkID int           `json:"bookmark_id"`
		RemarkURL  string        `json:"remark_url"`
		RemarkText template.HTML `json:"remark_text"`
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

func (n *Notification) GetRemarkPayload() *RemarkPayload {
	if n.Kind != KindRemark {
		return nil
	}
	var payload RemarkPayload
	if err := json.Unmarshal(n.Payload, &payload); err != nil {
		slog.Error("Failed to unmarshal remark payload",
			"err", err, "payload", string(n.Payload))
		return nil
	}
	return &payload
}

func (n *Notification) GetFollowPayload() *FollowPayload {
	if n.Kind != KindFollow {
		return nil
	}
	var payload FollowPayload
	if err := json.Unmarshal(n.Payload, &payload); err != nil {
		slog.Error("Failed to unmarshal follow payload",
			"err", err, "payload", string(n.Payload))
		return nil
	}
	return &payload
}
