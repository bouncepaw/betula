// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package htmlesc

import (
	"strings"
	"testing"

	"github.com/go-shiori/dom"
	"github.com/nalgeon/be"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func parseFirst(t *testing.T, s string) *html.Node {
	t.Helper()
	context := &html.Node{Type: html.ElementNode, Data: "div", DataAtom: atom.Div}
	nodes, err := html.ParseFragment(strings.NewReader(s), context)
	be.Err(t, err, nil)
	be.True(t, len(nodes) > 0)
	return nodes[0]
}

func render(t *testing.T, n *html.Node) string {
	t.Helper()
	var b strings.Builder
	be.Err(t, html.Render(&b, n), nil)
	return b.String()
}

func TestHeadingRule(t *testing.T) {
	t.Parallel()

	t.Run("Matches h1 through h6", func(t *testing.T) {
		t.Parallel()
		for _, tag := range []string{"h1", "h2", "h3", "h4", "h5", "h6"} {
			n := parseFirst(t, "<"+tag+">x</"+tag+">")
			be.True(t, headingRule{}.Matches(n))
		}
	})

	t.Run("Ignores non-headings", func(t *testing.T) {
		t.Parallel()
		be.True(t, !headingRule{}.Matches(parseFirst(t, `<p>x</p>`)))
		be.True(t, !headingRule{}.Matches(parseFirst(t, `<strong>x</strong>`)))
	})

	t.Run("Rewrites to p>strong keeping inline children", func(t *testing.T) {
		t.Parallel()
		got := render(t, headingRule{}.Rewrite(parseFirst(t, `<h2>My <em>post</em></h2>`)))
		be.Equal(t, got, `<p><strong>My <em>post</em></strong></p>`)
	})

	t.Run("Rewrites empty heading", func(t *testing.T) {
		t.Parallel()
		got := render(t, headingRule{}.Rewrite(parseFirst(t, `<h1></h1>`)))
		be.Equal(t, got, `<p><strong></strong></p>`)
	})
}

func TestDropClassRule(t *testing.T) {
	t.Parallel()

	t.Run("Matches quote-inline among other tokens", func(t *testing.T) {
		t.Parallel()
		be.True(t, dropClassRule{}.Matches(parseFirst(t, `<span class="foo quote-inline">x</span>`)))
		be.True(t, !dropClassRule{}.Matches(parseFirst(t, `<span class="quote-inlines">x</span>`)))
		be.True(t, !dropClassRule{}.Matches(parseFirst(t, `<span>x</span>`)))
	})

	t.Run("Rewrites to empty text node dropping contents", func(t *testing.T) {
		t.Parallel()
		got := render(t, dropClassRule{}.Rewrite(parseFirst(t, `<span class="quote-inline"><br/>RE: <a href="/x">x</a></span>`)))
		be.Equal(t, got, ``)
	})
}

func TestClassFilterRule(t *testing.T) {
	t.Parallel()

	t.Run("Matches only when class present", func(t *testing.T) {
		t.Parallel()
		be.True(t, classFilterRule{}.Matches(parseFirst(t, `<span class="x">y</span>`)))
		be.True(t, !classFilterRule{}.Matches(parseFirst(t, `<span>y</span>`)))
	})

	t.Run("Keeps allowed tokens dropping the rest", func(t *testing.T) {
		t.Parallel()
		n := parseFirst(t, `<span class="evil h-card bad mention">x</span>`)
		classFilterRule{}.Rewrite(n)
		be.Equal(t, dom.GetAttribute(n, "class"), "h-card mention")
	})

	t.Run("Removes class attr when nothing survives", func(t *testing.T) {
		t.Parallel()
		n := parseFirst(t, `<span class="evil bad">x</span>`)
		classFilterRule{}.Rewrite(n)
		be.Equal(t, dom.GetAttribute(n, "class"), "")
		be.Equal(t, len(n.Attr), 0)
	})
}

func TestClassAllowed(t *testing.T) {
	t.Parallel()

	t.Run("Allows microformats prefixes", func(t *testing.T) {
		t.Parallel()
		be.True(t, classAllowed("h-card"))
		be.True(t, classAllowed("u-url"))
		be.True(t, classAllowed("dt-published"))
	})

	t.Run("Allows semantic and wikilink classes", func(t *testing.T) {
		t.Parallel()
		be.True(t, classAllowed("mention"))
		be.True(t, classAllowed("hashtag"))
		be.True(t, classAllowed("wikilink_external"))
	})

	t.Run("Rejects unknown and near-miss tokens", func(t *testing.T) {
		t.Parallel()
		be.True(t, !classAllowed("evil"))
		be.True(t, !classAllowed("hashtagline"))
		be.True(t, !classAllowed(""))
	})
}
