// SPDX-FileCopyrightText: 2023 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

// Package myco wraps Mycomarkup invocation.
package myco

import (
	"html/template"

	"git.sr.ht/~bouncepaw/mycomarkup/v5"
	"git.sr.ht/~bouncepaw/mycomarkup/v5/mycocontext"
	"git.sr.ht/~bouncepaw/mycomarkup/v5/options"
)

var opts = options.Options{
	HyphaName:             "",
	WebSiteURL:            "",
	TransclusionSupported: false,
	RedLinksSupported:     false,
	InterwikiSupported:    false,
}.FillTheRest()

func MarkupToHTML(text string) template.HTML {
	ctx, _ := mycocontext.ContextFromStringInput(text, opts)
	return template.HTML(mycomarkup.BlocksToHTML(ctx, mycomarkup.BlockTree(ctx)))
}
