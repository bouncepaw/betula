// SPDX-FileCopyrightText: 2023 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2024 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

// Package assembly builds outgoing ActivityPub activities and objects.
// The actor is derived from the current settings on each call.
package assembly

import (
	apports "git.sr.ht/~bouncepaw/betula/ports/activitypub"
)

const atContext = "https://www.w3.org/ns/activitystreams"
const publicAudience = "https://www.w3.org/ns/activitystreams#Public"

type Dict = apports.Dict

type Assembler struct {
	siteURLFn       func() string
	adminUsernameFn func() string
}

var _ apports.Assembly = (*Assembler)(nil)

func New(siteURLFn, adminUsernameFn func() string) *Assembler {
	return &Assembler{
		siteURLFn:       siteURLFn,
		adminUsernameFn: adminUsernameFn,
	}
}

// actor is the actor to use for outgoing activities. It is derived from the
// current settings on every call, so settings changes take effect at once.
func (asm *Assembler) actor() string {
	username := asm.adminUsernameFn()
	if username == "" {
		username = "betula"
	}
	return asm.siteURLFn() + "/@" + username
}
