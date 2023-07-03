package web

import (
	"git.sr.ht/~bouncepaw/betula/auth"
	"git.sr.ht/~bouncepaw/betula/myco"
	"git.sr.ht/~bouncepaw/betula/settings"
	"git.sr.ht/~bouncepaw/betula/types"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"time"
)

/*
HTML paginator in Betula all have a common template, a skeleton, which is
stored in skeleton.gohtml. It expects several templates to be defines
beforehand. They include:

    * title, which is the <title> of the currentPage.
    * body, which is the main part of the currentPage, usually <main> and
      its contents.

For every view, a corresponding .gohtml and the skeleton are parsed
together. This file collects utilities for constructing such templates
and running them, as well as all such templates.
*/

func templateFrom(funcMap template.FuncMap, filenames ...string) *template.Template {
	for i, filename := range filenames {
		filenames[i] = filename + ".gohtml"
	}
	filenames = append(filenames, "skeleton.gohtml")
	return template.Must(template.New("skeleton.gohtml").Funcs(funcMap).ParseFS(fs, filenames...))
}

func templateExec(w http.ResponseWriter, temp *template.Template, data viewData, rq *http.Request) {
	common := dataCommon{
		authorized: auth.AuthorizedFromRequest(rq),
		siteTitle:  settings.SiteTitle(),
		siteName:   settings.SiteName(),
	}
	data.Fill(common)
	err := temp.ExecuteTemplate(w, "skeleton.gohtml", data)
	if err != nil {
		log.Fatalln(err)
	}
}

// Auth views:
var templateRegisterForm = templateFrom(nil, "register-form")
var templateLoginForm = templateFrom(nil, "login-form")
var templateLogoutForm = templateFrom(nil, "logout-form")
var templateSettings = templateFrom(nil, "settings")

// Sad views:
var templateStatus = templateFrom(nil, "status")

// Meaningful views:
var templateSaveLink = templateFrom(funcMapForForm, "link-form-fragment", "save-link", "submit-another")
var templateEditLink = templateFrom(funcMapForForm, "link-form-fragment", "edit-link")
var templatePost = templateFrom(funcMapForPosts, "post-fragment", "post")
var templateFeed = templateFrom(funcMapForPosts, "paginator-fragment", "post-fragment", "feed")
var templateSearch = templateFrom(funcMapForPosts, "paginator-fragment", "post-fragment", "search")
var templateTag = templateFrom(funcMapForPosts, "paginator-fragment", "post-fragment", "tag")
var templateTags = templateFrom(nil, "tags")
var templateDay = templateFrom(funcMapForPosts, "post-fragment", "day")
var templateEditTag = templateFrom(funcMapForForm, "edit-tag")

var templateAbout = templateFrom(funcMapForTime, "about")

var funcMapForPosts = template.FuncMap{
	"randomGlobe": func() string {
		return string([]rune{[]rune("🌍🌎🌏")[rand.Intn(3)]})
	},
	"timestampToHuman": func(stamp string) string {
		t, err := time.Parse(types.TimeLayout, stamp)
		if err != nil {
			// Sorry for party rocking...
			log.Fatalln(err)
		}
		return t.Format("2006-01-02 15:04")
	},
	"stripCommonProtocol": types.StripCommonProtocol,
	"mycomarkup":          myco.MarkupToHTML,
	"timestampToDayStamp": func(stamp string) string {
		// len("2000-00-00") == 10
		return stamp[:10] // Pray 🙏
	},
}

var funcMapForForm = template.FuncMap{
	"catsTogether": types.JoinTags,
}

var funcMapForTime = template.FuncMap{
	"timeToHuman": func(t *time.Time) string {
		return t.Format("2006-01-02 15:04")
	},
}

// Do not bother to fill it.
type dataCommon struct {
	authorized  bool
	siteTitle   template.HTML
	siteName    string
	head        template.HTML
	searchQuery string

	paginator []types.Page
}

type viewData interface {
	SiteTitleHTML() template.HTML
	Authorized() bool
	Fill(dataCommon)
	Head() template.HTML
	SearchQuery() string
	Pages() []types.Page
}

func (c *dataCommon) SearchQuery() string {
	return c.searchQuery
}

func (c *dataCommon) SiteTitleHTML() template.HTML {
	return c.siteTitle
}

func (c *dataCommon) SiteName() string {
	return c.siteName
}

func (c *dataCommon) Authorized() bool {
	return c.authorized
}

func (c *dataCommon) Head() template.HTML {
	return c.head
}

func (c *dataCommon) Pages() []types.Page {
	return c.paginator
}

func (c *dataCommon) Fill(C dataCommon) {
	if c == nil {
		panic("common data is nil! Initialize it at templateExec invocation.")
	}
	c.siteTitle = C.siteTitle
	c.authorized = C.authorized
	c.siteName = C.siteName
}

func emptyCommon() *dataCommon {
	return &dataCommon{}
}
