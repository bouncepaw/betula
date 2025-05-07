package web

import (
	"fmt"
	"git.sr.ht/~bouncepaw/betula/auth"
	"git.sr.ht/~bouncepaw/betula/myco"
	"git.sr.ht/~bouncepaw/betula/settings"
	"git.sr.ht/~bouncepaw/betula/types"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"strings"
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
	filenames = append(filenames, "skeleton")
	for i, filename := range filenames {
		filenames[i] = "views/" + filename + ".gohtml"
	}
	return template.Must(template.New("skeleton.gohtml").Funcs(funcMap).ParseFS(fs, filenames...))
}

func templateExec(w http.ResponseWriter, rq *http.Request, temp *template.Template, data viewData) {
	common := dataCommon{
		authorized: auth.AuthorizedFromRequest(rq),
		siteTitle:  settings.SiteTitle(),
		siteName:   settings.SiteName(),
	}

	if settings.FederationEnabled() && common.authorized {
		var givenHost = rq.Host
		var expectedHost = settings.SiteDomain()

		if expectedHost != givenHost {
			var notif = Notification{
				Category: NotificationHostMismatch,
				Body: template.HTML(fmt.Sprintf(
					`<b>[BET-113]</b> Configured to use the domain “%s”, but this request has Host header “%s”. Federation might not work. Is your reverse proxy misconfigured? Check <a href="/settings">Settings</a>. See <a href="/help/en/errors#BET_113">Help</a>.`,
					expectedHost, givenHost,
				)),
			}
			common.SystemNotifications = append(common.SystemNotifications, notif)
		}

		if strings.HasPrefix(settings.SiteURL(), "http://") {
			var notif = Notification{
				Category: NotificationWrongProtocol,
				Body: template.HTML(fmt.Sprintf(
					`<b>[BET-114]</b> Configured to use the address “%s”, which uses HTTP. Federation will not work. Check <a href="/settings">Settings</a>. See <a href="/help/en/errors#BET_114">Help</a>.`,
					settings.SiteURL())),
			}
			common.SystemNotifications = append(common.SystemNotifications, notif)
		}
	}

	data.Fill(common)
	err := temp.ExecuteTemplate(w, "skeleton.gohtml", data)
	if err != nil {
		log.Printf("While writing template: %s\n", err)
	}
}

// Auth views:
var templateRegisterForm = templateFrom(nil, "register-form")
var templateLoginForm = templateFrom(nil, "login-form")
var templateLogoutForm = templateFrom(nil, "logout-form")
var templateSettings = templateFrom(nil, "settings")
var templateSessions = templateFrom(funcMapForTime, "sessions")

// Sad views:
var templateStatus = templateFrom(nil, "status")

// Meaningful views:
var templateSaveLink = templateFrom(funcMapForForm, "link-form-fragment", "save-link")
var templateEditLink = templateFrom(funcMapForForm, "link-form-fragment", "edit-link")
var templateRepost = templateFrom(funcMapForForm, "repost")
var templatePost = templateFrom(funcMapForBookmarks, "post-fragment", "post")
var templateRepostsFor = templateFrom(funcMapForTime, "reposts-of")
var templateFeed = templateFrom(funcMapForBookmarks, "paginator-fragment", "post-fragment", "feed")
var templateSearch = templateFrom(funcMapForBookmarks, "paginator-fragment", "post-fragment", "search")
var templateTag = templateFrom(funcMapForBookmarks, "paginator-fragment", "post-fragment", "tag")
var templateTags = templateFrom(nil, "tags")
var templateDay = templateFrom(funcMapForBookmarks, "post-fragment", "day")
var templateEditTag = templateFrom(funcMapForForm, "edit-tag")
var templateHelp = templateFrom(nil, "help")
var templateAbout = templateFrom(funcMapForTime, "about")
var templateBookmarklet = templateFrom(nil, "bookmarklet")
var templateMyProfile = templateFrom(funcMapForTime, "my-profile")

// Fedded verse views:
var templateRemoteProfile = templateFrom(funcMapForBookmarks, "paginator-fragment", "timeline", "remote-profile")
var templateFollowing = templateFrom(nil, "following")
var templateFollowers = templateFrom(nil, "followers")
var templateTimeline = templateFrom(funcMapForBookmarks, "paginator-fragment", "timeline")
var templateFedisearch = templateFrom(funcMapForBookmarks, "fedisearch")

var funcMapForBookmarks = template.FuncMap{
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
	"shortenLink": func(a string) template.HTML {
		result, pathPart := types.CleanerLinkParts(a)
		result += fmt.Sprintf(`<span class="url-path">%s</span>`, pathPart)
		return template.HTML(result)
	},
	"mycomarkup": myco.MarkupToHTML,
	"timestampToDayStamp": func(stamp string) string {
		// len("2000-00-00") == 10
		return stamp[:10] // Pray 🙏
	},
}

var funcMapForForm = template.FuncMap{
	"catsTogether": types.JoinTags,
	"catStringsTogether": func(tags []string) string {
		return types.JoinTags(types.TagsFromStringSlice(tags))
	},
}

var funcMapForTime = template.FuncMap{
	"timeToHuman": func(t *time.Time) string {
		return t.Format("2006-01-02 15:04")
	},
}

type NotificationCategory string

const (
	NotificationHostMismatch  NotificationCategory = "Host mismatch"
	NotificationWrongProtocol NotificationCategory = "Wrong protocol"
	NotificationSuccess       NotificationCategory = "Success"
	NotificationFailure       NotificationCategory = "Failure"
)

type Notification struct {
	Category NotificationCategory
	Body     template.HTML
}

// Do not bother to fill it.
type dataCommon struct {
	authorized  bool
	siteTitle   template.HTML
	siteName    string
	head        template.HTML
	searchQuery string

	paginator           []types.Page
	SystemNotifications []Notification
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

func (c *dataCommon) MultiplePages() bool {
	return len(c.Pages()) > 1
}

func (c *dataCommon) FederationEnabled() bool {
	return settings.FederationEnabled()
}

func (c *dataCommon) AdminUsername() string {
	return settings.AdminUsername()
}

func (c *dataCommon) Fill(C dataCommon) {
	if c == nil {
		panic("common data is nil! Initialize it at templateExec invocation.")
	}
	c.siteTitle = C.siteTitle
	c.authorized = C.authorized
	c.siteName = C.siteName
	c.SystemNotifications = append(c.SystemNotifications, C.SystemNotifications...)
}

func emptyCommon() *dataCommon {
	return &dataCommon{}
}

func commonWithAutoCompletion() *dataCommon {
	common := emptyCommon()
	common.head = `<script defer src="/static/autocompletion.js"></script>`
	return common
}
