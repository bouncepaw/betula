// SPDX-FileCopyrightText: 2025 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2026 Iaroslav Angliuster <https://mysh.dev>
//
// SPDX-License-Identifier: AGPL-3.0-only

package notifsvc

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"html/template"
	"log/slog"

	"git.sr.ht/~bouncepaw/betula/db"
	wwwgw "git.sr.ht/~bouncepaw/betula/gateways/www"
	apports "git.sr.ht/~bouncepaw/betula/ports/activitypub"
	remotebookmarkssvc "git.sr.ht/~bouncepaw/betula/svc/remotebookmarks"
	"git.sr.ht/~bouncepaw/betula/types"
	notiftypes "git.sr.ht/~bouncepaw/betula/types/notif"
)

var (
	actorRepo = db.NewActorRepo()
	sanitizer = wwwgw.NewSanitizer()
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
// TODO: Architecture is a mess around here.

func renderAuthor(actorID string) (template.HTML, error) {
	actor, err := actorRepo.GetActorByID(context.Background(), actorID, apports.GetActorsOpts{})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}
	return types.RenderedAuthorLink(actorID, actor, err == nil), nil
}

type renderedNotification notiftypes.Notification

type notificationTemplateData struct {
	Author     template.HTML
	RemarkURL  string
	RemarkText template.HTML
	BookmarkID int
}

var (
	likeNotificationTemplate = template.Must(template.New("like notification").Parse(`<div class="notif" notif-cat="like">
	<span class="actor-link">{{.Author}}</span> liked bookmark <a href="/{{.BookmarkID}}">{{.BookmarkID}}.</a>
</div>`))
	followNotificationTemplate = template.Must(template.New("follow notification").Parse(`<div class="notif" notif-cat="follow">
	<span class="actor-link">{{.Author}}</span> followed you!
</div>`))
	remarkNotificationTemplate = template.Must(template.New("remark notification").Parse(`<div class="notif" notif-cat="remark">
	<span class="actor-link">{{.Author}}</span> <a href="{{.RemarkURL}}">remarked</a> <a href="/{{.BookmarkID}}">{{.BookmarkID}}.</a>{{if .RemarkText}}<blockquote>{{.RemarkText}}</blockquote>{{end}}
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

	author, err := renderAuthor(payload.ActorID)
	if err != nil {
		return "", err
	}
	return renderTemplate(
		likeNotificationTemplate,
		notificationTemplateData{
			Author:     author,
			BookmarkID: payload.BookmarkID,
		},
	)
}

func (n *renderedNotification) followAsHTML() (template.HTML, error) {
	var payload notiftypes.FollowPayload
	if err := json.Unmarshal(n.Payload, &payload); err != nil {
		return "", err
	}

	author, err := renderAuthor(payload.ActorID)
	if err != nil {
		return "", err
	}
	return renderTemplate(
		followNotificationTemplate,
		notificationTemplateData{
			Author: author,
		},
	)
}

func (n *renderedNotification) remarkAsHTML() (template.HTML, error) {
	var payload notiftypes.RemarkPayload
	if err := json.Unmarshal(n.Payload, &payload); err != nil {
		return "", err
	}

	author, err := renderAuthor(payload.ActorID)
	if err != nil {
		return "", err
	}

	return renderTemplate(
		remarkNotificationTemplate,
		notificationTemplateData{
			Author:    author,
			RemarkURL: payload.RemarkURL,
			RemarkText: remotebookmarkssvc.RenderRemoteDescription(
				sanitizer,
				payload.Source,
				payload.SourceType,
				payload.DescriptionHTML,
			),
			BookmarkID: payload.BookmarkID,
		},
	)
}
