package web

import (
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"strings"
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

func templateFrom(funcMap template.FuncMap, filenames ...string) *template.Template {
	filenames = append(filenames, "skeleton.gohtml")
	return template.Must(template.New("skeleton.gohtml").Funcs(funcMap).ParseFS(fs, filenames...))
}

func templateExec(temp *template.Template, data any, w http.ResponseWriter) {
	err := temp.ExecuteTemplate(w, "skeleton.gohtml", data)
	if err != nil {
		log.Fatalln(err)
	}
}

// Auth views:
var templateRegisterForm = templateFrom(nil, "register-form.gohtml")
var templateLoginForm = templateFrom(nil, "login-form.gohtml")
var templateLogoutForm = templateFrom(nil, "logout-form.gohtml")

// Sad views:
var template404 = templateFrom(nil, "404.gohtml")

// Meaningful views:
var templateSaveLink = templateFrom(nil, "link-form-fragment.gohtml", "save-link.gohtml")
var templateEditLink = templateFrom(nil, "link-form-fragment.gohtml", "edit-link.gohtml")
var templateAddLinkInvalidURL = templateFrom(nil, "link-form-fragment.gohtml", "save-link-invalid-url.gohtml")
var templatePost = templateFrom(funcMapForPosts, "post-fragment.gohtml", "post.gohtml")
var templateFeed = templateFrom(funcMapForPosts, "post-fragment.gohtml", "feed.gohtml")
var templateCategories = templateFrom(nil, "categories.gohtml")
var templateCategory = templateFrom(funcMapForPosts, "post-fragment.gohtml", "category.gohtml")
var templateAbout = templateFrom(funcMapForTime, "about.gohtml")

var funcMapForPosts = template.FuncMap{
	"randomGlobe": func() string {
		return string([]rune{[]rune("🌍🌎🌏")[rand.Intn(3)]})
	},
	"timestampToHuman": func(stamp int64) string {
		t := time.Unix(stamp, 0)
		return t.Format("2006-01-02 15:04")
	},
	"stripCommonProtocol": func(a string) string {
		b := strings.TrimPrefix(a, "https://")
		c := strings.TrimPrefix(b, "http://")
		// Gemini, Gopher, FTP, Mail are not stripped, to emphasize them, when they are.
		return c
	},
}

var funcMapForTime = template.FuncMap{
	"timeToHuman": func(t time.Time) string {
		return t.Format("2006-01-02 15:04")
	},
}
