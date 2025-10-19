package notif

import (
	"encoding/json"
	"errors"
	"fmt"
	"git.sr.ht/~bouncepaw/betula/db"
	notiftypes "git.sr.ht/~bouncepaw/betula/types/notif"
	"html/template"
	"log/slog"
)

// Things that didn't fit in the template.
// TODO: Might introduce caching in the future, maybe.

var ErrActorNotFound = errors.New("actor not found")

type RenderedNotification notiftypes.Notification

func (n *RenderedNotification) AsHTML() template.HTML {
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

func (n *RenderedNotification) likeAsHTML() (template.HTML, error) {
	var payload notiftypes.LikePayload
	if err := json.Unmarshal(n.Payload, &payload); err != nil {
		return "", err
	}

	actor, found := db.ActorByID(payload.ActorID)
	if !found {
		return "", ErrActorNotFound
	}
	return template.HTML(fmt.Sprintf(
		`<div class="notif" notif-cat="like">
	<a href="/%s">%s</a> liked bookmark <a href="/%d">%d.</a>
</div>`,
		actor.Acct(), actor.DisplayedName, payload.BookmarkID, payload.BookmarkID,
	)), nil
}

func (n *RenderedNotification) followAsHTML() (template.HTML, error) {
	var payload notiftypes.FollowPayload
	if err := json.Unmarshal(n.Payload, &payload); err != nil {
		return "", err
	}

	actor, found := db.ActorByID(payload.ActorID)
	if !found {
		return "", ErrActorNotFound
	}
	return template.HTML(fmt.Sprintf(
		`<div class="notif" notif-cat="follow">
	<a href="/%s">%s</a> followed you!
</div>`,
		actor.Acct(), actor.DisplayedName)), nil
}

func (n *RenderedNotification) remarkAsHTML() (template.HTML, error) {
	var payload notiftypes.RemarkPayload
	if err := json.Unmarshal(n.Payload, &payload); err != nil {
		return "", err
	}

	actor, found := db.ActorByID(payload.ActorID)
	if !found {
		return "", ErrActorNotFound
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
