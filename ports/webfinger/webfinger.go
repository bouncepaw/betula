// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package webfingerports

import (
	"fmt"

	"git.sr.ht/~bouncepaw/betula/pkg/bxstr"
	"git.sr.ht/~bouncepaw/betula/types"
)

type WebFinger interface {
	DereferenceAcct(Acct) (id string, err error)
}

type (
	Acct struct {
		User, Host string
	}
	Document struct {
		Links []struct {
			Rel  string `json:"rel"`
			Type string `json:"type"`
			Href string `json:"href"`
		} `json:"links"`
	}
)

func (a Acct) String() string {
	return fmt.Sprintf("acct:%s@%s", a.User, a.Host)
}

func (d Document) ActivityPubActorID() string {
	for _, link := range d.Links {
		if link.Rel == "self" && link.Type == types.OtherActivityType && bxstr.IsValidURL(link.Href) {
			return link.Href
		}
	}
	return ""
}
