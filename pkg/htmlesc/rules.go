// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package htmlesc

import (
	"slices"
	"strings"

	"github.com/go-shiori/dom"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type headingRule struct{}

var headingAtoms = []atom.Atom{atom.H1, atom.H2, atom.H3, atom.H4, atom.H5, atom.H6}

func (headingRule) Matches(n *html.Node) bool {
	return slices.Contains(headingAtoms, n.DataAtom)
}

func (headingRule) Rewrite(n *html.Node) *html.Node {
	p := dom.CreateElement("p")
	strong := dom.CreateElement("strong")
	dom.AppendChild(p, strong)

	for _, c := range dom.ChildNodes(n) {
		dom.AppendChild(strong, c)
	}
	return p
}

type dropClassRule struct{}

func (dropClassRule) Matches(n *html.Node) bool {
	const fep044fQuoteReNotice = "quote-inline"
	return slices.Contains(strings.Fields(dom.GetAttribute(n, "class")), fep044fQuoteReNotice)
}

func (dropClassRule) Rewrite(n *html.Node) *html.Node {
	return dom.CreateTextNode("")
}

type classFilterRule struct{}

func (classFilterRule) Matches(n *html.Node) bool {
	return dom.GetAttribute(n, "class") != ""
}

func (classFilterRule) Rewrite(n *html.Node) *html.Node {
	var kept []string
	for c := range strings.FieldsSeq(dom.GetAttribute(n, "class")) {
		if classAllowed(c) {
			kept = append(kept, c)
		}
	}
	if len(kept) == 0 {
		dom.RemoveAttribute(n, "class")
	} else {
		dom.SetAttribute(n, "class", strings.Join(kept, " "))
	}
	return n
}

var (
	classPrefixes = []string{"h-", "p-", "u-", "dt-", "e-", "wikilink"}
	classExact    = []string{"mention", "hashtag", "ellipsis", "invisible"}
)

func classAllowed(c string) bool {
	hasPrefix := func(p string) bool { return strings.HasPrefix(c, p) }
	return slices.ContainsFunc(classPrefixes, hasPrefix) || slices.Contains(classExact, c)
}
