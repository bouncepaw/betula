package web

import (
	"git.sr.ht/~bouncepaw/betula/types"
	"net/http"
)

type errorTemplate interface {
	emptyUrl(post types.Post, data *dataCommon, w http.ResponseWriter, rq *http.Request)
	invalidUrl(post types.Post, data *dataCommon, w http.ResponseWriter, rq *http.Request)
	titleNotFound(post types.Post, data *dataCommon, w http.ResponseWriter, rq *http.Request)
}

/* Error templates for edit link currentPage */

func (d dataEditLink) emptyUrl(post types.Post, data *dataCommon, w http.ResponseWriter, rq *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	templateExec(w, rq, templateEditLink, dataEditLink{
		Post:          post,
		dataCommon:    data,
		ErrorEmptyURL: true,
	})
}

func (d dataEditLink) invalidUrl(post types.Post, data *dataCommon, w http.ResponseWriter, rq *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	templateExec(w, rq, templateEditLink, dataEditLink{
		Post:            post,
		dataCommon:      data,
		ErrorInvalidURL: true,
	})
}

func (d dataEditLink) titleNotFound(post types.Post, data *dataCommon, w http.ResponseWriter, rq *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
	templateExec(w, rq, templateEditLink, dataEditLink{
		Post:               post,
		dataCommon:         data,
		ErrorTitleNotFound: true,
	})
}

/* Error templates for save link currentPage */

func (d dataSaveLink) emptyUrl(post types.Post, data *dataCommon, w http.ResponseWriter, rq *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	templateExec(w, rq, templateSaveLink, dataSaveLink{
		Post:          post,
		dataCommon:    data,
		ErrorEmptyURL: true,
	})
}

func (d dataSaveLink) invalidUrl(post types.Post, data *dataCommon, w http.ResponseWriter, rq *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	templateExec(w, rq, templateSaveLink, dataSaveLink{
		Post:            post,
		dataCommon:      data,
		ErrorInvalidURL: true,
	})
}

func (d dataSaveLink) titleNotFound(post types.Post, data *dataCommon, w http.ResponseWriter, rq *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
	templateExec(w, rq, templateSaveLink, dataSaveLink{
		Post:               post,
		dataCommon:         data,
		ErrorTitleNotFound: true,
	})
}
