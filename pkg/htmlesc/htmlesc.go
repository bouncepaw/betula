// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

// Package htmlesc provides the HTML-escaping function.
package htmlesc

import "html/template"

type Options struct {
	// If true, remote hashtag links will be replaced with local hashtag links.
	LocalizeHashtags bool
	// If true, mentions will be replaced with local remote actor profiles links.
	LocalizeMentions bool
}

var DefaultOptions = Options{
	LocalizeHashtags: true,
	LocalizeMentions: true,
}

// Escape returns raw verbatim. It will escape in the future.
func Escape(raw template.HTML, opts ...Options) template.HTML {
	opt := append(opts, DefaultOptions)[0]
	_ = opt

	// TODO: https://codeberg.org/bouncepaw/betula/issues/255
	return raw
}
