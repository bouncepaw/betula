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

	"git.sr.ht/~bouncepaw/betula/pkg/bxstr"
)

func (asm *Assembler) NewReject(rejectedActivity apports.Dict) (json.RawMessage, error) {
	delete(rejectedActivity, "@context")
	activity := apports.Dict{
		"@context": atContext,
		"id":       fmt.Sprintf("%s/temp/%s", asm.siteURLFn(), bxstr.RandomWhatever()),
		"type":     "Reject",
		"actor":    asm.actor(),
		"object":   rejectedActivity,
	}
	return json.Marshal(activity)
}
