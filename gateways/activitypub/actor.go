// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package apgw

import (
	"fmt"
	"git.sr.ht/~bouncepaw/betula/jobs"
	apports "git.sr.ht/~bouncepaw/betula/ports/activitypub"
	"git.sr.ht/~bouncepaw/betula/types"
)

type Actor struct {
	id                string
	inboxURL          string
	preferredUsername string
	displayedName     string
	summary           string
	publicKey         struct {
		id           string
		owner        string
		publicKeyPEM string
	}

	domain string
}

var _ apports.Actor = &Actor{}

func newActor(dto types.Actor) *Actor {
	return &Actor{
		id:                dto.ID,
		inboxURL:          dto.Inbox,
		preferredUsername: dto.PreferredUsername,
		displayedName:     dto.DisplayedName,
		summary:           dto.Summary,
		publicKey: struct {
			id           string
			owner        string
			publicKeyPEM string
		}{
			id:           dto.PublicKey.ID,
			owner:        dto.PublicKey.Owner,
			publicKeyPEM: dto.PublicKey.PublicKeyPEM,
		},

		domain: dto.Domain,
	}
}

func (a *Actor) ID() string {
	return a.id
}

func (a *Actor) Acct() string {
	return fmt.Sprintf("@%s@%s", a.preferredUsername, a.domain)
}

func (a *Actor) DisplayedName() string {
	return a.displayedName
}

func (a *Actor) SendSerializedActivity(activity []byte) error {
	// TODO: that function shall be this function.
	return jobs.SendActivityToInbox(activity, a.inboxURL)
}
