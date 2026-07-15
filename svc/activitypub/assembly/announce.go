// SPDX-FileCopyrightText: 2023 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2024 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package assembly

import (
	"encoding/json"
)

func (asm *Assembler) NewAnnounce(originalURL string, repostURL string) (json.RawMessage, error) {
	activity := map[string]any{
		"@context": atContext,
		"type":     "Announce",
		"actor": map[string]string{
			"id":                asm.actor(),
			"preferredUsername": asm.adminUsernameFn(),
		},
		"id":     repostURL,
		"object": originalURL,
	}
	return json.Marshal(activity)
}
