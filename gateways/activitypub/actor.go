// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package apgw

import (
	"git.sr.ht/~bouncepaw/betula/jobs"
	apports "git.sr.ht/~bouncepaw/betula/ports/activitypub"
)

type Actor struct {
	id       string
	inboxURL string
}

var _ apports.Actor = &Actor{}

func NewActor(id, inboxURL string) *Actor {
	return &Actor{
		id:       id,
		inboxURL: inboxURL,
	}
}

func (a *Actor) ID() string {
	return a.id
}

func (a *Actor) SendSerializedActivity(activity []byte) error {
	// TODO: that function shall be this function.
	return jobs.SendActivityToInbox(activity, a.inboxURL)
}
