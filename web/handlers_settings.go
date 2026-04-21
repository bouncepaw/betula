// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package web

import (
	"fmt"
	"html/template"
	"log/slog"
	"net/http"

	"git.sr.ht/~bouncepaw/betula/ports/settings"
)

// TODO: move other settings endpoints here along with copyright notices.

type dataLoggingSettings struct {
	*dataCommon
	Method   string
	URL      string
	Username string
	Token    string
}

func (data dataLoggingSettings) withLoggingSettings(ls settingsports.LoggingSettings) dataLoggingSettings {
	if ls.Method != nil {
		data.Method = string(*ls.Method)
	}
	if ls.URL != nil {
		data.URL = *ls.URL
	}
	if ls.Username != nil {
		data.Username = *ls.Username
	}
	if ls.Token != nil {
		data.Token = *ls.Token
	}
	return data
}

func (data dataLoggingSettings) withCoolCSS() dataLoggingSettings {
	data.head = `<style>
form:has(select[name="method"] option[value=""]:checked) [data-logging-show] { display: none; }
form:has(select[name="method"] option[value="ECS + No Auth"]:checked) [data-logging-show="auth"],
form:has(select[name="method"] option[value="ECS + No Auth"]:checked) [data-logging-show="basic-auth"] { display: none; }
form:has(select[name="method"] option[value="ECS + Bearer"]:checked) [data-logging-show="basic-auth"] { display: none; }
</style>`
	return data
}

func getLoggingSettings(w http.ResponseWriter, rq *http.Request) {
	ls, err := ctrl.SvcSettings.GetLoggingSettings(rq.Context())
	if err != nil {
		slog.Error("Failed to get logging settings", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data := dataLoggingSettings{
		dataCommon: emptyCommon(),
	}.withLoggingSettings(ls).withCoolCSS()
	templateExec(w, rq, templateLoggingSettings, data)
}

func postLoggingSettings(w http.ResponseWriter, rq *http.Request) {
	ls := settingsports.LoggingSettings{}

	method := settingsports.LoggingMethod(rq.FormValue("method"))
	if method != settingsports.LoggingMethodDefault {
		ls.Method = &method
	}
	if url := rq.FormValue("url"); url != "" {
		ls.URL = &url
	}
	if username := rq.FormValue("username"); username != "" {
		ls.Username = &username
	}
	if token := rq.FormValue("token"); token != "" {
		ls.Token = &token
	}

	var notif SystemNotification
	if err := ctrl.SvcSettings.SaveLoggingSettings(rq.Context(), ls); err != nil {
		slog.Error("Failed to save logging settings", "err", err)
		notif = SystemNotification{
			Category: NotificationFailure,
			Body:     template.HTML(fmt.Sprintf("Failed to save logging settings: %s.", err)),
		}
	} else {
		notif = SystemNotification{
			Category: NotificationSuccess,
			Body:     "Logging settings saved.",
		}
	}

	data := dataLoggingSettings{
		dataCommon: emptyCommon().withSystemNotifications(notif),
	}.withLoggingSettings(ls).withCoolCSS()
	templateExec(w, rq, templateLoggingSettings, data)
}
