// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package wwwports

// WorldWideWeb fetches information from the web.
type WorldWideWeb interface {
	// TitleOfPage returns <title> value for the given web page.
	TitleOfPage(addr string) (string, error)
}
