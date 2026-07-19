// SPDX-FileCopyrightText: 2022-2025 Betula contributors
// SPDX-FileCopyrightText: 2026 Danila Gorelko
//
// SPDX-License-Identifier: AGPL-3.0-only

package parsing

import (
	"testing"

	apports "git.sr.ht/~bouncepaw/betula/ports/activitypub"
	"github.com/nalgeon/be"
)

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
	report, err := testGuesser.Guess(undoFollowJSON)
	be.Err(t, err, nil)
	undoFollowReport, ok := report.(apports.UndoFollowReport)
	be.True(t, ok)
	// and just a little check
	be.Equal(t, undoFollowReport.ActorID, "https://bob.bouncepaw.com/@bob")
}
