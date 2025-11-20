// SPDX-FileCopyrightText: 2022-2025 Betula contributors
//
// SPDX-License-Identifier: AGPL-3.0-only

package notifsvc

import (
	"encoding/json"
	"errors"
	"fmt"
	"git.sr.ht/~bouncepaw/betula/db"
	"git.sr.ht/~bouncepaw/betula/types/notif"
	"html/template"
	"log/slog"
)

// Render returns an HTML representation of the notification
// that is ready to be inserted on the notifications page.
//
// This didn't fit well into the templates well, so it's
// a separate function.
func Render(notif notiftypes.Notification) template.HTML {
	rendered := renderedNotification(notif)
	return rendered.AsHTML()
}

// TODO: Might introduce caching in the future, maybe.

var errActorNotFound = errors.New("actor not found")

type renderedNotification notiftypes.Notification

func (n *renderedNotification) AsHTML() template.HTML {
	var (
		html template.HTML
		err  error
	)
	switch n.Kind {
	case notiftypes.KindLike:
		html, err = n.likeAsHTML()
	case notiftypes.KindFollow:
		html, err = n.followAsHTML()
	case notiftypes.KindRemark:
		html, err = n.remarkAsHTML()
	}
	if err != nil {
		slog.Error("Failed to render notification",
			"notifPayload", string(n.Payload), "notifID", n.ID, "err", err)
		return ""
	}

	return html
}

func (n *renderedNotification) likeAsHTML() (template.HTML, error) {
	var payload notiftypes.LikePayload
	if err := json.Unmarshal(n.Payload, &payload); err != nil {
		return "", err
	}

	actor, found := db.ActorByID(payload.ActorID)
	if !found {
		return "", errActorNotFound
	}
	return template.HTML(fmt.Sprintf(
		`<div class="notif" notif-cat="like">
	<a href="/%s">%s</a> liked bookmark <a href="/%d">%d.</a>
</div>`,
		actor.Acct(), actor.DisplayedName, payload.BookmarkID, payload.BookmarkID,
	)), nil
}

func (n *renderedNotification) followAsHTML() (template.HTML, error) {
	var payload notiftypes.FollowPayload
	if err := json.Unmarshal(n.Payload, &payload); err != nil {
		return "", err
	}

	actor, found := db.ActorByID(payload.ActorID)
	if !found {
		return "", errActorNotFound
	}
	return template.HTML(fmt.Sprintf(
		`<div class="notif" notif-cat="follow">
	<a href="/%s">%s</a> followed you!
</div>`,
		actor.Acct(), actor.DisplayedName)), nil
}

func (n *renderedNotification) remarkAsHTML() (template.HTML, error) {
	var payload notiftypes.RemarkPayload
	if err := json.Unmarshal(n.Payload, &payload); err != nil {
		return "", err
	}

	actor, found := db.ActorByID(payload.ActorID)
	if !found {
		return "", errActorNotFound
	}

	// TODO: s/repost/remark when the time comes
	// TODO: link/show local representation of the remark after the big refac
	// TODO: support the case with remark text
	return template.HTML(fmt.Sprintf(
		`<div class="notif" notif-cat="remark">
	<a href="/%s">%s</a> <a href="%s">reposted</a> <a href="/%d">%d.</a>
</div>`,
		actor.Acct(), actor.DisplayedName, payload.RemarkURL, payload.BookmarkID, payload.BookmarkID,
	)), nil
}
