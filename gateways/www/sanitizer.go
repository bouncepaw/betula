// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package wwwgw

import (
	"html/template"
	"net/url"
	"slices"
	"strings"

	"github.com/go-shiori/dom"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"git.sr.ht/~bouncepaw/betula/pkg/htmlesc"
	wwwports "git.sr.ht/~bouncepaw/betula/ports/www"
	"git.sr.ht/~bouncepaw/betula/types"
)

type Sanitizer struct {
}

var _ wwwports.HTMLSanitizer = &Sanitizer{}

func NewSanitizer() *Sanitizer {
	return &Sanitizer{}
}

func (s *Sanitizer) Sanitize(raw template.HTML) template.HTML {
	return htmlesc.Escape(raw, mentionRule{}, hashtagRule{}, linkRule{})
}

type mentionRule struct{}

func (mentionRule) Matches(n *html.Node) bool {
	return n.DataAtom == atom.A && hasClass(n, "mention") && !hasClass(n, "hashtag")
}

func (mentionRule) Rewrite(n *html.Node) *html.Node {
	if path := profilePath(dom.GetAttribute(n, "href")); path != "" {
		dom.SetAttribute(n, "href", path)
	}
	return n
}

func profilePath(rawHref string) string {
	u, err := url.Parse(rawHref)
	if err != nil || u.Host == "" {
		return ""
	}

	segments := strings.Split(strings.Trim(u.Path, "/"), "/")
	user := strings.TrimPrefix(segments[len(segments)-1], "@")
	if user == "" {
		return ""
	}
	return "/@" + user + "@" + u.Host
}

type hashtagRule struct{}

func (hashtagRule) Matches(n *html.Node) bool {
	return n.DataAtom == atom.A && hasClass(n, "hashtag")
}

func (hashtagRule) Rewrite(n *html.Node) *html.Node {
	name := types.CanonicalTagName(strings.TrimSpace(dom.TextContent(n)))
	if name != "" {
		dom.SetAttribute(n, "href", "/tag/"+name)
	}
	return n
}

type linkRule struct{}

func (linkRule) Matches(n *html.Node) bool {
	return n.DataAtom == atom.A && dom.GetAttribute(n, "href") != ""
}

func (linkRule) Rewrite(n *html.Node) *html.Node {
	addClass(n, "wikilink")
	if u, err := url.Parse(dom.GetAttribute(n, "href")); err == nil && u.Scheme != "" {
		addClass(n, "wikilink_external")
		addClass(n, "wikilink_"+u.Scheme)
	}
	return n
}

func hasClass(n *html.Node, class string) bool {
	return slices.Contains(strings.Fields(dom.ClassName(n)), class)
}

func addClass(n *html.Node, class string) {
	if hasClass(n, class) {
		return
	}
	if existing := dom.ClassName(n); existing != "" {
		dom.SetAttribute(n, "class", existing+" "+class)
	} else {
		dom.SetAttribute(n, "class", class)
	}
}
