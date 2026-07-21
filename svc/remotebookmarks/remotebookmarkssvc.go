// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package remotebookmarkssvc

import (
	"fmt"

	apports "git.sr.ht/~bouncepaw/betula/ports/activitypub"
	likingports "git.sr.ht/~bouncepaw/betula/ports/liking"
	remotebookmarksports "git.sr.ht/~bouncepaw/betula/ports/remotebookmarks"
	wwwports "git.sr.ht/~bouncepaw/betula/ports/www"
)

type Service struct {
	sanitizer          wwwports.HTMLSanitizer
	localBookmarkRepo  likingports.LocalBookmarkRepository
	remoteBookmarkRepo remotebookmarksports.RemoteBookmarkRepository
	activityPub        apports.ActivityPub

	siteURLFn       func() string
	adminUsernameFn func() string
	siteDomainFn    func() string
}

var _ remotebookmarksports.Service = (*Service)(nil)

func New(
	sanitizer wwwports.HTMLSanitizer,
	localBookmarkRepo likingports.LocalBookmarkRepository,
	remoteBookmarkRepo remotebookmarksports.RemoteBookmarkRepository,
	activityPub apports.ActivityPub,
	siteURLFn func() string,
	adminUsernameFn func() string,
	siteDomainFn func() string,
) *Service {
	return &Service{
		sanitizer:          sanitizer,
		localBookmarkRepo:  localBookmarkRepo,
		remoteBookmarkRepo: remoteBookmarkRepo,
		activityPub:        activityPub,
		siteURLFn:          siteURLFn,
		adminUsernameFn:    adminUsernameFn,
		siteDomainFn:       siteDomainFn,
	}
}

type ownActor struct {
	id       string
	username string
	domain   string
}

var _ apports.Actor = ownActor{}

func (a ownActor) ID() string                { return a.id }
func (a ownActor) Acct() string              { return fmt.Sprintf("@%s@%s", a.username, a.domain) }
func (a ownActor) PreferredUsername() string { return a.username }
func (a ownActor) DisplayedName() string     { return a.username }
func (a ownActor) SendSerializedActivity([]byte) error {
	return fmt.Errorf("cannot send an activity to our own actor")
}
