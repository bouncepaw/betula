// SPDX-FileCopyrightText: 2023 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2024 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package assembly

import (
	"encoding/json"
	"fmt"
	apports "git.sr.ht/~bouncepaw/betula/ports/activitypub"
)

func (asm *Assembler) NewUndoFollowFromUs(objectID string) (json.RawMessage, error) {
	activity := apports.Dict{
		"@context": atContext,
		"id":       fmt.Sprintf("%s/unfollow?account=%s", asm.siteURLFn(), objectID),
		"type":     "Undo",
		"actor":    asm.actor(),
		"object": apports.Dict{
			"id":     fmt.Sprintf("%s/follow?account=%s", asm.siteURLFn(), objectID),
			"type":   "Follow",
			"actor":  asm.actor(),
			"object": objectID,
		},
	}
	return json.Marshal(activity)
}

func (asm *Assembler) NewFollowFromUs(objectID string) (json.RawMessage, error) {
	activity := apports.Dict{
		"@context": atContext,
		"id":       fmt.Sprintf("%s/follow?account=%s", asm.siteURLFn(), objectID),
		"type":     "Follow",
		"actor":    asm.actor(),
		"object":   objectID,
	}
	return json.Marshal(activity)
}
