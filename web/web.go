// SPDX-FileCopyrightText: 2022 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2023 Danila Gorelko
// SPDX-FileCopyrightText: 2023 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2023 ninedraft
// SPDX-FileCopyrightText: 2024 Danila Gorelko
// SPDX-FileCopyrightText: 2024 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2025 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2026 Danila Gorelko
//
// SPDX-License-Identifier: AGPL-3.0-only

// Package web provides web capabilities. Import this package to initialize the handlers and the templates.
package web

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"

	"git.sr.ht/~bouncepaw/betula/db"
	"git.sr.ht/~bouncepaw/betula/types"

	"git.sr.ht/~bouncepaw/betula/auth"
	"git.sr.ht/~bouncepaw/betula/settings"
)

var serverRestartChannel = make(chan struct{})

func StartServer(c Controller) {
	ctrl = c
	go restartServer()
	var srv = &http.Server{}
	for range serverRestartChannel {
		if err := srv.Close(); err != nil {
			// Is it important? Does it matter?
			slog.Info("Closing server", "err", err)
		}
		srv = &http.Server{
			Addr:    listenAddr(),
			Handler: &auther{mux},
		}
		slog.Info("Running HTTP server", "addr", srv.Addr)
		go func() {
			if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				slog.Error("HTTP server failed", "err", err)
				os.Exit(1)
			}
		}()
	}
}

func listenAddr() string {
	return fmt.Sprintf("%s:%d", settings.NetworkHost(), settings.NetworkPort())
}

func restartServer() {
	serverRestartChannel <- struct{}{}
}

type auther struct {
	http.Handler
}

type dataAuthorized struct {
	*dataCommon
	Status string
}

func (a *auther) ServeHTTP(w http.ResponseWriter, rq *http.Request) {
	// Auth is OK if it is set up or the user wants to set it up or they request static data.
	authOK := auth.Ready() ||
		strings.HasPrefix(rq.URL.Path, "/static/") ||
		strings.HasPrefix(rq.URL.Path, "/register")

	// We don't support anything else.
	// A thought for a future Bouncepaw: maybe we should support HEAD?
	allowedMethod := rq.Method == http.MethodGet || rq.Method == http.MethodPost

	if !allowedMethod {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write(
			fmt.Appendf(nil, "Method %s is not supported by this server. Use POST and GET.", rq.Method))
		return
	}

	if !authOK {
		templateExec(w, rq, templateRegisterForm, dataAuthorized{
			dataCommon: emptyCommon(),
		})
		return
	}

	a.Handler.ServeHTTP(w, rq)
}

func extractPage(rq *http.Request) (currentPage uint) {
	if page, err := strconv.Atoi(rq.FormValue("page")); err != nil || page == 0 {
		currentPage = 1
	} else {
		currentPage = uint(page)
	}
	return
}

func extractBookmark(w http.ResponseWriter, rq *http.Request) (*types.Bookmark, bool) {
	id, ok := extractBookmarkID(w, rq)
	if !ok {
		return nil, false
	}

	bookmark, found := db.GetBookmarkByID(id)
	if !found {
		slog.Info("Bookmark not found", "path", rq.URL.Path, "id", id)
		handlerNotFound(w, rq)
		return nil, false
	}

	authed := auth.AuthorizedFromRequest(rq)
	if bookmark.Visibility == types.Private && !authed {
		slog.Info("Unauthorized attempt to access", "path", rq.URL.Path, "status", http.StatusUnauthorized)
		handlerUnauthorized(w, rq)
		return nil, false
	}

	bookmark.Tags = db.TagsForBookmarkByID(bookmark.ID)

	return &bookmark, true
}

// returns id, found.
func extractBookmarkID(w http.ResponseWriter, rq *http.Request) (int, bool) {
	id, err := strconv.Atoi(rq.PathValue("id"))
	if err != nil {
		slog.Info("Extracting bookmark id: wrong format", "path", rq.URL.Path)
		handlerNotFound(w, rq)
		return 0, false
	}
	return id, true
}

// Wrap handlers that only make sense for the admin with this thingy in init().
func adminOnly(next func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, rq *http.Request) {
		authed := auth.AuthorizedFromRequest(rq)
		if !authed {
			slog.Info("Unauthorized attempt to access", "path", rq.URL.Path, "status", http.StatusUnauthorized)
			handlerUnauthorized(w, rq)
			return
		}
		next(w, rq)
	}
}

func federatedOnly(next func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, rq *http.Request) {
		federated := settings.FederationEnabled()
		if !federated {
			slog.Info("Attempt to access failed: Betula is not federated", "path", rq.URL.Path, "status", http.StatusUnauthorized)
			handlerNotFederated(w, rq)
			return
		}
		next(w, rq)
	}
}

func fediverseWebFork(
	nextFedi func(http.ResponseWriter, *http.Request),
	nextWeb func(http.ResponseWriter, *http.Request),
) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, rq *http.Request) {
		if strings.HasPrefix(rq.URL.Path, "/@") {
			handlerAt(w, rq)
			return
		}
		wantsActivity := strings.Contains(rq.Header.Get("Accept"), types.ActivityType) || strings.Contains(rq.Header.Get("Accept"), types.OtherActivityType)
		if wantsActivity && nextFedi != nil {
			federatedOnly(nextFedi)(w, rq)
		} else if nextWeb != nil {
			nextWeb(w, rq)
		} else {
			handlerNotFound(w, rq)
		}
	}
}
