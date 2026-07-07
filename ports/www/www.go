// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package wwwports

import (
	"errors"
	"html/template"
	"net/url"
)

var (
	ErrTimeout      = errors.New("request timed out")
	ErrNoTitleFound = errors.New("no title found in the document")
)

// WorldWideWeb fetches information from the web.
type WorldWideWeb interface {
	// TitleOfPage returns <title> value for the given web page.
	TitleOfPage(addr string) (string, error)
	// RelAlternates returns all <link rel="alternate"> found on the web page.
	RelAlternates(addr string) ([]RelAlternate, error)
}

type RelAlternate struct {
	Type  string
	Href  string
	Title string
}

func (a RelAlternate) ResolveHref(base string) string {
	b, bErr := url.Parse(base)
	ref, refErr := url.Parse(a.Href)
	if bErr != nil || refErr != nil {
		return a.Href
	}
	return b.ResolveReference(ref).String()
}

type HTMLSanitizer interface {
	Sanitize(html template.HTML) template.HTML
}
