// SPDX-FileCopyrightText: 2022-2025 Betula contributors
//
// SPDX-License-Identifier: AGPL-3.0-only

package activities

import "testing"

// This one was handled wrong at some point. Making a test here to fix it.
var undoFollowJSON = []byte(`
{
  "@context": "https://www.w3.org/ns/activitystreams",
  "actor": "https://bob.bouncepaw.com/@bob",
  "id": "https://bob.bouncepaw.com/unfollow?account=https://alice.bouncepaw.com/@alice",
  "object": {
    "actor": "https://bob.bouncepaw.com/@bob",
    "id": "https://bob.bouncepaw.com/follow?account=https://alice.bouncepaw.com/@alice",
    "object": "https://alice.bouncepaw.com/@alice",
    "type": "Follow"
  },
  "type": "Undo"
}`)

func TestGuessUndoFollow(t *testing.T) {
	report, err := Guess(undoFollowJSON)
	if err != nil {
		t.Error(err)
		t.Logf("%q", report)
		return
	}
	undoFollowReport, ok := report.(UndoFollowReport)
	if !ok {
		t.Error("wrong type")
		t.Logf("%q", report)
		return
	}
	// and just a little check
	if undoFollowReport.ActorID != "https://bob.bouncepaw.com/@bob" {
		t.Error("it's all messed up")
		t.Logf("%q", report)
		return
	}
}
