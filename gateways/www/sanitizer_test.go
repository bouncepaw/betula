// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package wwwgw

import (
	"html/template"
	"testing"

	"github.com/nalgeon/be"
)

func sanitize(in string) string {
	return string((&Sanitizer{}).Sanitize(template.HTML(in)))
}

func TestSanitize(t *testing.T) {
	t.Parallel()

	t.Run("Mention qualified with host and classed", func(t *testing.T) {
		t.Parallel()
		be.Equal(t,
			sanitize(`<span class="h-card"><a href="https://merveilles.example/@ink" class="u-url mention">@<span>ink</span></a></span>`),
			`<span class="h-card"><a href="/@ink@merveilles.example" class="u-url mention wikilink">@<span>ink</span></a></span>`)
	})

	t.Run("Hashtag re-pointed and classed", func(t *testing.T) {
		t.Parallel()
		be.Equal(t,
			sanitize(`<a href="https://merveilles.example/tags/FediDev" class="mention hashtag" rel="tag">#<span>FediDev</span></a>`),
			`<a href="/tag/fedidev" class="mention hashtag wikilink" rel="tag">#<span>FediDev</span></a>`)
	})

	t.Run("External link gets wikilink classes", func(t *testing.T) {
		t.Parallel()
		be.Equal(t,
			sanitize(`<a href="https://example.example/page">plain</a>`),
			`<a href="https://example.example/page" class="wikilink wikilink_external wikilink_https">plain</a>`)
	})

	t.Run("Mailto link gets wikilink classes", func(t *testing.T) {
		t.Parallel()
		be.Equal(t,
			sanitize(`<a href="mailto:a@b.example">mail</a>`),
			`<a href="mailto:a@b.example" class="wikilink wikilink_external wikilink_mailto">mail</a>`)
	})

	t.Run("Dangerous markup stripped by underlying policy", func(t *testing.T) {
		t.Parallel()
		be.Equal(t,
			sanitize(`<p>hi <a href="javascript:alert(1)" onclick="x()">x</a><script>bad()</script></p>`),
			`<p>hi <a class="wikilink wikilink_external wikilink_javascript">x</a></p>`)
	})
}

func TestProfilePath(t *testing.T) {
	t.Parallel()

	t.Run("Mastodon-style @user path", func(t *testing.T) {
		t.Parallel()
		be.Equal(t, profilePath("https://merveilles.example/@ink"), "/@ink@merveilles.example")
	})

	t.Run("Pleroma-style users path", func(t *testing.T) {
		t.Parallel()
		be.Equal(t, profilePath("https://pleroma.example/users/bob"), "/@bob@pleroma.example")
	})

	t.Run("Trailing slash tolerated", func(t *testing.T) {
		t.Parallel()
		be.Equal(t, profilePath("https://social.example/@alice/"), "/@alice@social.example")
	})

	t.Run("Empty when no host", func(t *testing.T) {
		t.Parallel()
		be.Equal(t, profilePath("/@ink"), "")
	})

	t.Run("Empty when no user segment", func(t *testing.T) {
		t.Parallel()
		be.Equal(t, profilePath("https://social.example"), "")
	})
}
