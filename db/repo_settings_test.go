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

	settingsports "git.sr.ht/~bouncepaw/betula/ports/settings"
)

func TestSetCredentials(t *testing.T) {
	InitInMemoryDB()
	ctx := t.Context()
	repo := &SettingsRepo{}
	be.Err(t, repo.SetCredentials(ctx, pufferfish, tropicfish), nil)

	name, err := repo.MetaEntryString(ctx, settingsports.BetulaMetaAdminUsername)
	be.Err(t, err, nil)
	be.Equal(t, name, pufferfish)

	hash, err := repo.MetaEntryString(ctx, settingsports.BetulaMetaAdminPasswordHash)
	be.Err(t, err, nil)
	be.Equal(t, hash, tropicfish)
}
