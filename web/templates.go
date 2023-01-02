package web

import (
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"time"
)

/*
HTML pages in Betula all have a common template, a skeleton, which is
stored in skeleton.gohtml. It expects several templates to be defines
beforehand. They include:

    * title, which is the <title> of the page.
    * body, which is the main part of the page, usually <main> and
      its contents.

For every view, a corresponding .gohtml and the skeleton are parsed
together. This file collects utilities for constructing such templates
and running them, as well as all such templates.
*/

func templateFrom(filename string, funcMap template.FuncMap) *template.Template {
	return template.Must(template.New("skeleton.gohtml").Funcs(funcMap).ParseFS(fs, filename, "skeleton.gohtml"))
}

func templateExec(temp *template.Template, data any, w http.ResponseWriter) {
	err := temp.ExecuteTemplate(w, "skeleton.gohtml", data)
	if err != nil {
		log.Fatalln(err)
	}
}

var templateAddLink = templateFrom("add-link.gohtml", nil)
var templateAddLinkInvalidURL = templateFrom("add-link-invalid-url.gohtml", nil)
var templatePost = templateFrom("post.gohtml", funcMapForPosts)
var templateFeed = templateFrom("feed.gohtml", funcMapForPosts)

var funcMapForPosts = template.FuncMap{
	"randomGlobe": func() string {
		return string([]rune{[]rune("🌍🌎🌏")[rand.Intn(3)]})
	},
	"timestampToHuman": func(stamp int64) string {
		t := time.Unix(stamp, 0)
		return t.Format("2006-01-02 15:04")
	},
}
