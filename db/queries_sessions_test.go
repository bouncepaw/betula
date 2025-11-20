// SPDX-FileCopyrightText: 2022-2025 Betula contributors
//
// SPDX-License-Identifier: AGPL-3.0-only

package db

import "testing"

// testing AddSession, SessionExists, StopSession
func TestSessionOps(t *testing.T) {
	InitInMemoryDB()
	token := pufferfish
	AddSession(token, "")
	if !SessionExists(token) {
		t.Errorf("Existing token not found")
	}
	StopSession(token)
	if SessionExists(token) {
		t.Errorf("Non-existent token found")
	}
}

func TestSetCredentials(t *testing.T) {
	InitInMemoryDB()
	SetCredentials(pufferfish, tropicfish)
	if MetaEntry[string](BetulaMetaAdminUsername) != pufferfish {
		t.Errorf("Wrong username returned")
	}
	if MetaEntry[string](BetulaMetaAdminPasswordHash) != tropicfish {
		t.Errorf("Wrong password hash returned")
	}
}
