// SPDX-FileCopyrightText: 2025 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2026 Iaroslav Angliuster <https://mysh.dev>
//
// SPDX-License-Identifier: AGPL-3.0-only

package notifsvc

import (
	"bytes"
	"encoding/json"
	"errors"
	"html/template"
	"log/slog"

	"git.sr.ht/~bouncepaw/betula/db"
	notiftypes "git.sr.ht/~bouncepaw/betula/types/notif"
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

type notificationTemplateData struct {
	Acct          string
	DisplayedName string
	RemarkURL     string
	BookmarkID    int
}

var (
	likeNotificationTemplate = template.Must(template.New("like notification").Parse(`<div class="notif" notif-cat="like">
	<a href="/{{.Acct}}">{{.DisplayedName}}</a> liked bookmark <a href="/{{.BookmarkID}}">{{.BookmarkID}}.</a>
</div>`))
	followNotificationTemplate = template.Must(template.New("follow notification").Parse(`<div class="notif" notif-cat="follow">
	<a href="/{{.Acct}}">{{.DisplayedName}}</a> followed you!
</div>`))
	remarkNotificationTemplate = template.Must(template.New("remark notification").Parse(`<div class="notif" notif-cat="remark">
	<a href="/{{.Acct}}">{{.DisplayedName}}</a> <a href="{{.RemarkURL}}">reposted</a> <a href="/{{.BookmarkID}}">{{.BookmarkID}}.</a>
</div>`))
)

func renderTemplate(tmpl *template.Template, data notificationTemplateData) (template.HTML, error) {
	var html bytes.Buffer
	if err := tmpl.Execute(&html, data); err != nil {
		return "", err
	}
	return template.HTML(html.String()), nil
}

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
	return renderTemplate(
		likeNotificationTemplate,
		notificationTemplateData{
			Acct:          actor.Acct(),
			DisplayedName: actor.DisplayedName,
			BookmarkID:    payload.BookmarkID,
		},
	)
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
	return renderTemplate(
		followNotificationTemplate,
		notificationTemplateData{
			Acct:          actor.Acct(),
			DisplayedName: actor.DisplayedName,
		},
	)
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
	return renderTemplate(
		remarkNotificationTemplate,
		notificationTemplateData{
			Acct:          actor.Acct(),
			DisplayedName: actor.DisplayedName,
			RemarkURL:     payload.RemarkURL,
			BookmarkID:    payload.BookmarkID,
		},
	)
}
