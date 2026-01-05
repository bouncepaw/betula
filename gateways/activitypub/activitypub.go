// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

// Package apgw provides the ActivityPub gateway.
package apgw

import (
	"fmt"
	"git.sr.ht/~bouncepaw/betula/fediverse"
	apports "git.sr.ht/~bouncepaw/betula/ports/activitypub"
	"git.sr.ht/~bouncepaw/betula/settings"
	"git.sr.ht/~bouncepaw/betula/stricks"
	"strconv"
	"strings"
)

type ActivityPub struct {
	remoteBookmarkRepo apports.RemoteBookmarkRepository
}

var _ apports.ActivityPub = &ActivityPub{}

func NewActivityPub(
	remoteBookmarkRepo apports.RemoteBookmarkRepository,
) *ActivityPub {
	return &ActivityPub{
		remoteBookmarkRepo: remoteBookmarkRepo,
	}
}

func (ap *ActivityPub) AuthorOfRemoteBookmark(remoteBookmarkID string) (apports.Actor, error) {
	actorID, err := ap.remoteBookmarkRepo.GetActorIDFor(remoteBookmarkID)
	if err != nil {
		return nil, err
	}

	inboxURL := fediverse.RequestActorInboxByID(actorID)
	return NewActor(actorID, inboxURL), nil
}

func (ap *ActivityPub) KnowsRemoteBookmark(remoteBookmarkID string) (bool, error) {
	if !stricks.ValidURL(remoteBookmarkID) {
		return false, fmt.Errorf("not url, thus not remote bookmark id: %s", remoteBookmarkID)
	}

	return ap.remoteBookmarkRepo.Exists(remoteBookmarkID)
}

func (ap *ActivityPub) LocalBookmarkIDFromActivityPubID(id string) (int, error) {
	if !strings.HasPrefix(id, settings.SiteURL()) {
		return 0, fmt.Errorf("url %s does not start with our address %s",
			id, settings.SiteURL())
	}

	parts := strings.Split(id, "/")
	lastPart := parts[len(parts)-1]
	return strconv.Atoi(lastPart)
}
