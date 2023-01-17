// Package myco wraps Mycomarkup invocation.
package myco

import (
	"git.sr.ht/~bouncepaw/mycomarkup/v5"
	"git.sr.ht/~bouncepaw/mycomarkup/v5/mycocontext"
	"git.sr.ht/~bouncepaw/mycomarkup/v5/options"
	"html/template"
)

var opts = options.Options{
	HyphaName:                        "",
	WebSiteURL:                       "",
	TransclusionSupported:            false,
	RedLinksSupported:                false,
	InterwikiSupported:               false,
	HyphaExists:                      func(_ string) bool { return true },
	IterateHyphaNamesWith:            func(_ func(string)) {},
	HyphaHTMLData:                    nil,
	LocalTargetCanonicalName:         func(s string) string { return s },
	LocalLinkHref:                    func(s string) string { return "/" + s },
	LocalImgSrc:                      func(s string) string { return s },
	LinkHrefFormatForInterwikiPrefix: nil,
	ImgSrcFormatForInterwikiPrefix:   nil,
}

func MarkupToHTML(text string) template.HTML {
	ctx, _ := mycocontext.ContextFromStringInput(text, opts)
	return template.HTML(mycomarkup.BlocksToHTML(ctx, mycomarkup.BlockTree(ctx)))
}
