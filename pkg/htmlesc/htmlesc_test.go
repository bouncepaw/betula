// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package htmlesc

import (
	"embed"
	"html/template"
	"testing"

	"github.com/go-shiori/dom"
	"github.com/nalgeon/be"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func escape(in string) string {
	return string(Escape(template.HTML(in)))
}

type localizeLinks struct{}

func (localizeLinks) Matches(n *html.Node) bool { return n.DataAtom == atom.A }

func (localizeLinks) Rewrite(n *html.Node) *html.Node {
	dom.SetAttribute(n, "href", "/local")
	return n
}

func TestEscape(t *testing.T) {
	t.Parallel()

	t.Run("Strips script", func(t *testing.T) {
		t.Parallel()
		be.Equal(t, escape(`<p>hi</p><script>alert(1)</script>`), `<p>hi</p>`)
	})

	t.Run("Drops event handler and javascript url", func(t *testing.T) {
		t.Parallel()
		be.Equal(t, escape(`<a href="javascript:alert(1)" onclick="x()">x</a>`), `x`)
	})

	t.Run("Heading becomes p>strong keeping inline markup", func(t *testing.T) {
		t.Parallel()
		be.Equal(t, escape(`<h2>My <em>post</em></h2>`), `<p><strong>My <em>post</em></strong></p>`)
	})

	t.Run("Unknown class token dropped, semantic kept", func(t *testing.T) {
		t.Parallel()
		be.Equal(t, escape(`<span class="evil h-card">x</span>`), `<span class="h-card">x</span>`)
	})

	t.Run("External link kept as-is", func(t *testing.T) {
		t.Parallel()
		be.Equal(t, escape(`<a href="https://example.example">x</a>`), `<a href="https://example.example">x</a>`)
	})

	t.Run("Mailto allowed", func(t *testing.T) {
		t.Parallel()
		be.Equal(t, escape(`<a href="mailto:a@b.example">mail</a>`), `<a href="mailto:a@b.example">mail</a>`)
	})

	t.Run("Whole note preserves mention and hashtag markup", func(t *testing.T) {
		t.Parallel()
		be.Equal(t,
			escape(`<h2>My <em>post</em></h2>`+
				`<p>Hey <span class="h-card"><a href="https://social.example/@bob" class="u-url mention">@<span>bob</span></a></span>, `+
				`see <a href="https://example.example/x">this</a> `+
				`<a href="https://merveilles.example/tags/FediDev" class="mention hashtag" rel="tag">#<span>FediDev</span></a></p>`),
			`<p><strong>My <em>post</em></strong></p>`+
				`<p>Hey <span class="h-card"><a href="https://social.example/@bob" class="u-url mention">@<span>bob</span></a></span>, `+
				`see <a href="https://example.example/x">this</a> `+
				`<a href="https://merveilles.example/tags/FediDev" class="mention hashtag" rel="tag">#<span>FediDev</span></a></p>`)
	})

	t.Run("Nested lists preserved", func(t *testing.T) {
		t.Parallel()
		be.Equal(t,
			escape(`<ul><li>one<ul><li>one-a</li><li>one-b</li></ul></li><li>two</li></ul>`),
			`<ul><li>one<ul><li>one-a</li><li>one-b</li></ul></li><li>two</li></ul>`)
	})

	t.Run("Blockquote keeps nested inline formatting, strips unknown", func(t *testing.T) {
		t.Parallel()
		be.Equal(t,
			escape(`<blockquote><p>quote <strong>bold <i>italic</i></strong> <marquee>no</marquee></p></blockquote>`),
			`<blockquote><p>quote <strong>bold <i>italic</i></strong> no</p></blockquote>`)
	})

	t.Run("Ordered list keeps allowed attrs, drops handler", func(t *testing.T) {
		t.Parallel()
		be.Equal(t,
			escape(`<ol start="3" reversed onclick="x()"><li value="5">item</li></ol>`),
			`<ol start="3" reversed=""><li value="5">item</li></ol>`)
	})

	t.Run("Injected rule runs before default rules", func(t *testing.T) {
		t.Parallel()
		got := string(Escape(template.HTML(`<a href="https://example.example">x</a>`), localizeLinks{}))
		be.Equal(t, got, `<a href="/local">x</a>`)
	})
}

//go:embed testdata/mastodon_note.html testdata/mastodon_note.golden.html
var testdataFS embed.FS

func TestEscapeRealMastodonNote(t *testing.T) {
	t.Parallel()

	in, err := testdataFS.ReadFile("testdata/mastodon_note.html")
	be.Err(t, err, nil)
	want, err := testdataFS.ReadFile("testdata/mastodon_note.golden.html")
	be.Err(t, err, nil)

	got := string(Escape(template.HTML(in)))
	be.Equal(t, got, string(want))
}
