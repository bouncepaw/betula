// SPDX-FileCopyrightText: 2023 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2024 Danila Gorelko
// SPDX-FileCopyrightText: 2024 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2026 Danila Gorelko
//
// SPDX-License-Identifier: AGPL-3.0-only

package db

import (
	"testing"

	"github.com/nalgeon/be"
)

// testing AddSession, SessionExists, StopSession.
func TestSessionOps(t *testing.T) {
	InitInMemoryDB()
	token := pufferfish
	AddSession(token, "")
	be.True(t, SessionExists(token))
	StopSession(token)
	be.True(t, !SessionExists(token))
}

func TestSetCredentials(t *testing.T) {
	InitInMemoryDB()
	SetCredentials(pufferfish, tropicfish)
	be.Equal(t, MetaEntry[string](BetulaMetaAdminUsername), pufferfish)
	be.Equal(t, MetaEntry[string](BetulaMetaAdminPasswordHash), tropicfish)
}
