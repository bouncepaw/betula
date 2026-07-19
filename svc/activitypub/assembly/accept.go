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

// NewAccept wraps the acceptedActivity in an Accept activity.
// The @context of the wrapped activity is deleted.
func (asm *Assembler) NewAccept(acceptedActivity apports.Dict) (json.RawMessage, error) {
	delete(acceptedActivity, "@context")
	return json.Marshal(apports.Dict{
		"@context": atContext,
		"id":       fmt.Sprintf("%s/temp/%s", asm.siteURLFn(), bxstr.RandomWhatever()),
		"type":     "Accept",
		"actor":    asm.actor(),
		"object":   acceptedActivity,
	})
}
