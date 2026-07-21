// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package remarkingsvc

import (
	apports "git.sr.ht/~bouncepaw/betula/ports/activitypub"
	notifports "git.sr.ht/~bouncepaw/betula/ports/notif"
	remarkingports "git.sr.ht/~bouncepaw/betula/ports/remarking"
	remotebookmarksports "git.sr.ht/~bouncepaw/betula/ports/remotebookmarks"
)

type Service struct {
	activityPub         apports.ActivityPub
	remarkRepo          remarkingports.Repository
	remoteBookmarksRepo remotebookmarksports.RemoteBookmarkRepository
	actorRepo           apports.ActorRepository
	assembly            apports.Assembly
	notifRepo           notifports.Repository

	federationEnabledFn func() bool
}

var _ remarkingports.Service = &Service{}

func New(
	activityPub apports.ActivityPub,
	remarkRepo remarkingports.Repository,
	remoteBookmarksRepo remotebookmarksports.RemoteBookmarkRepository,
	actorRepo apports.ActorRepository,
	assembly apports.Assembly,
	notifRepo notifports.Repository,
	federationEnabledFn func() bool,
) *Service {
	return &Service{
		activityPub:         activityPub,
		remarkRepo:          remarkRepo,
		remoteBookmarksRepo: remoteBookmarksRepo,
		actorRepo:           actorRepo,
		assembly:            assembly,
		notifRepo:           notifRepo,
		federationEnabledFn: federationEnabledFn,
	}
}
