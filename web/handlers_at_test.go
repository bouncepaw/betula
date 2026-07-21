// SPDX-FileCopyrightText: 2026 Danila Gorelko
//
// SPDX-License-Identifier: AGPL-3.0-only

package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"git.sr.ht/~bouncepaw/betula/auth"
	"git.sr.ht/~bouncepaw/betula/db"
	"git.sr.ht/~bouncepaw/betula/settings"
	"git.sr.ht/~bouncepaw/betula/types"
	"github.com/nalgeon/be"
)

func TestHandlerAtRedirectsOwnRemoteProfileToIndex(t *testing.T) {
	db.InitInMemoryDB()
	auth.SetCredentials("bob", "password")
	settings.SetSettings(types.Settings{SiteURL: "https://example.com"})
	r := httptest.NewRequest(http.MethodGet, "/@bob@example.com", nil)
	w := httptest.NewRecorder()
	handlerAt(w, r)
	res := w.Result()
	be.Equal(t, res.StatusCode, http.StatusSeeOther)
	be.Equal(t, res.Header.Get("Location"), "/")
}
