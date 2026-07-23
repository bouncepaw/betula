// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package htmlesc

import (
	"html/template"
	"slices"
	"strings"

	"github.com/go-shiori/dom"
	"github.com/microcosm-cc/bluemonday"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type Rule interface {
	Matches(*html.Node) bool
	Rewrite(*html.Node) *html.Node
}

var defaultRules = []Rule{
	headingRule{},
	dropClassRule{},
	classFilterRule{},
}

var policy = bluemonday.NewPolicy()

func init() {
	policy.AllowElements(
		"p", "span", "br", "del", "pre", "code", "em", "strong",
		"b", "i", "u", "ul", "ol", "li", "blockquote", "mark", "sub", "sup",
	)

	policy.AllowAttrs("class").OnElements("span")
	policy.AllowAttrs("href", "rel", "class").OnElements("a")
	policy.AllowAttrs("start", "reversed").OnElements("ol")
	policy.AllowAttrs("value").OnElements("li")

	policy.AllowElements("img")
	policy.AllowAttrs("src", "alt", "title").OnElements("img")

	policy.AllowRelativeURLs(true)
	policy.AllowURLSchemes(
		"http", "https", "dat", "dweb", "ipfs", "ipns",
		"ssb", "gopher", "xmpp", "magnet", "gemini", "mailto",
	)
}

func Escape(raw template.HTML, additionalRules ...Rule) template.HTML {
	rules := slices.Concat(additionalRules, defaultRules)

	context := &html.Node{Type: html.ElementNode, Data: "div", DataAtom: atom.Div}
	nodes, err := html.ParseFragment(strings.NewReader(string(raw)), context)
	if err != nil {
		return fallback(raw)
	}
	for _, n := range nodes {
		context.AppendChild(n)
	}

	escapeNode(context, rules)

	var buf strings.Builder
	for c := context.FirstChild; c != nil; c = c.NextSibling {
		if err = html.Render(&buf, c); err != nil {
			return fallback(raw)
		}
	}

	return template.HTML(policy.Sanitize(buf.String()))
}

func escapeNode(node *html.Node, rules []Rule) {
	node = applyRules(node, rules)

	var next *html.Node
	for c := node.FirstChild; c != nil; c = next {
		next = c.NextSibling
		escapeNode(c, rules)
	}
}

func applyRules(node *html.Node, rules []Rule) *html.Node {
	if node.Type != html.ElementNode {
		return node
	}
	for _, rule := range rules {
		if !rule.Matches(node) {
			continue
		}
		replacement := rule.Rewrite(node)
		if replacement != node {
			dom.ReplaceChild(node.Parent, replacement, node)
			node = replacement
		}
	}
	return node
}

func fallback(raw template.HTML) template.HTML {
	return template.HTML(template.HTMLEscapeString(string(raw)))
}
