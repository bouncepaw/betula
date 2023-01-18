// Package myco wraps Mycomarkup invocation.
package myco

import (
	"git.sr.ht/~bouncepaw/mycomarkup/v5"
	"git.sr.ht/~bouncepaw/mycomarkup/v5/mycocontext"
	"git.sr.ht/~bouncepaw/mycomarkup/v5/options"
	"html/template"
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
