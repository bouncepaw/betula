// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package wwwgw

import (
	wwwports "git.sr.ht/~bouncepaw/betula/ports/www"
	"git.sr.ht/~bouncepaw/betula/readpage"
)

type WWW struct{}

var _ wwwports.WorldWideWeb = &WWW{}

func New() *WWW {
	return &WWW{}
}

func (www *WWW) TitleOfPage(addr string) (string, error) {
	// TODO: https://codeberg.org/bouncepaw/betula/issues/153
	return readpage.FindTitle(addr)
}
