// SPDX-FileCopyrightText: 2023 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2024 Danila Gorelko
// SPDX-FileCopyrightText: 2024 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2026 Danila Gorelko
// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
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
	ctx := t.Context()
	repo := NewSessionsRepo()
	token := pufferfish

	be.Err(t, repo.AddSession(ctx, token, ""), nil)

	exists, err := repo.SessionExists(ctx, token)
	be.Err(t, err, nil)
	be.True(t, exists)

	be.Err(t, repo.StopSession(ctx, token), nil)

	exists, err = repo.SessionExists(ctx, token)
	be.Err(t, err, nil)
	be.True(t, !exists)
}
