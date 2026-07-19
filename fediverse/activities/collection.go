// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package activities

import (
	"encoding/json"

	apports "git.sr.ht/~bouncepaw/betula/ports/activitypub"
)

func collectionFromDict(dict Dict) (*apports.Collection, error) {
	// A bit ineffective innit.
	j, err := json.Marshal(dict)
	if err != nil {
		return nil, err
	}

	var collection apports.Collection
	err = json.Unmarshal(j, &collection)
	if err != nil {
		return nil, err
	}

	if err = collection.Valid(); err != nil {
		return nil, err
	}

	return &collection, nil
}
