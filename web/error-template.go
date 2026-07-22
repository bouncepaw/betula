// SPDX-FileCopyrightText: 2023 Danila Gorelko
// SPDX-FileCopyrightText: 2023 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2024 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package web

import (
	"git.sr.ht/~bouncepaw/betula/types"
	"net/http"
)

type errorTemplate interface {
	emptyUrl(bookmark types.Bookmark, data *dataCommon, w http.ResponseWriter, rq *http.Request)
	invalidUrl(bookmark types.Bookmark, data *dataCommon, w http.ResponseWriter, rq *http.Request)
	titleNotFound(bookmark types.Bookmark, data *dataCommon, w http.ResponseWriter, rq *http.Request)
}

/* Error templates for edit link currentPage */

func (d dataEditLink) emptyUrl(bookmark types.Bookmark, data *dataCommon, w http.ResponseWriter, rq *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	templateExec(w, rq, templateEditLink, dataEditLink{
		Bookmark:      bookmark,
		dataCommon:    data,
		ErrorEmptyURL: true,
	})
}

func (d dataEditLink) invalidUrl(bookmark types.Bookmark, data *dataCommon, w http.ResponseWriter, rq *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	templateExec(w, rq, templateEditLink, dataEditLink{
		Bookmark:        bookmark,
		dataCommon:      data,
		ErrorInvalidURL: true,
	})
}

func (d dataEditLink) titleNotFound(bookmark types.Bookmark, data *dataCommon, w http.ResponseWriter, rq *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
	templateExec(w, rq, templateEditLink, dataEditLink{
		Bookmark:           bookmark,
		dataCommon:         data,
		ErrorTitleNotFound: true,
	})
}

/* Error templates for save link currentPage */

func (d dataSaveLink) emptyUrl(bookmark types.Bookmark, data *dataCommon, w http.ResponseWriter, rq *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	templateExec(w, rq, templateSaveLink, dataSaveLink{
		Bookmark:      bookmark,
		dataCommon:    data,
		ErrorEmptyURL: true,
	})
}

func (d dataSaveLink) invalidUrl(bookmark types.Bookmark, data *dataCommon, w http.ResponseWriter, rq *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	templateExec(w, rq, templateSaveLink, dataSaveLink{
		Bookmark:        bookmark,
		dataCommon:      data,
		ErrorInvalidURL: true,
	})
}

func (d dataSaveLink) titleNotFound(bookmark types.Bookmark, data *dataCommon, w http.ResponseWriter, rq *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
	templateExec(w, rq, templateSaveLink, dataSaveLink{
		Bookmark:           bookmark,
		dataCommon:         data,
		ErrorTitleNotFound: true,
	})
}
