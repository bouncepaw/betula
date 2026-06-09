// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package web

import (
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	imexports "git.sr.ht/~bouncepaw/betula/ports/imex"
	"git.sr.ht/~bouncepaw/betula/types"
)

type dataImport struct {
	*dataCommon
}

func getImport(w http.ResponseWriter, rq *http.Request) {
	templateExec(w, rq, templateImport, dataImport{
		dataCommon: emptyCommon(),
	})
}

func postImport(w http.ResponseWriter, rq *http.Request) {
	file, _, err := rq.FormFile("file")
	if err != nil {
		templateExec(w, rq, templateImport, dataImport{
			dataCommon: emptyCommon().withSystemNotifications(SystemNotification{
				Category: NotificationFailure,
				Body:     template.HTML(fmt.Sprintf("Failed to read the uploaded file: %s.", err)),
			}),
		})
		return
	}
	defer file.Close()

	var params imexports.ImportParams
	for _, tag := range types.SplitTags(rq.FormValue("assign-tags")) {
		params.AddTags = append(params.AddTags, tag.Name)
	}
	params.KeepDuplicate = rq.FormValue("keep-duplicate") == "true"
	params.MakePublic = rq.FormValue("make-public") == "true"

	count, err := ctrl.SvcImEx.Import(rq.Context(), params, file)

	var notif SystemNotification
	if err != nil {
		slog.Error("Bookmark import failed", "err", err)
		notif = SystemNotification{
			Category: NotificationFailure,
			Body:     template.HTML(fmt.Sprintf("Import failed: %s.", err)),
		}
	} else {
		notif = SystemNotification{
			Category: NotificationSuccess,
			Body:     template.HTML(fmt.Sprintf("Imported %d bookmarks.", count)),
		}
	}

	templateExec(w, rq, templateImport, dataImport{
		dataCommon: emptyCommon().withSystemNotifications(notif),
	})
}

type dataExport struct {
	*dataCommon
}

func getExport(w http.ResponseWriter, rq *http.Request) {
	common := emptyCommon()
	if errTxt := rq.FormValue("err"); errTxt != "" {
		common.withSystemNotifications(SystemNotification{
			Category: NotificationFailure,
			Body:     template.HTML(fmt.Sprintf("Export failed: %s.", errTxt)),
		})
	}
	templateExec(w, rq, templateExport, dataExport{
		dataCommon: common,
	})
}

func postExport(w http.ResponseWriter, rq *http.Request) {
	params := imexports.ExportParams{
		Format:         imexports.ExportFormat(rq.FormValue("format")),
		IncludePrivate: rq.FormValue("include-private") == "true",
	}

	filename := fmt.Sprintf(
		"%s Betula bookmarks.%s",
		time.Now().UTC().Format(time.DateOnly),
		params.Format.FileExtension())
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

	err := ctrl.SvcImEx.Export(rq.Context(), params, w)
	if err != nil {
		slog.Error("Bookmark export failed", "err", err)
		http.Redirect(w, rq, fmt.Sprintf("/export?err=%s", url.QueryEscape(err.Error())), http.StatusSeeOther)
	}
}
