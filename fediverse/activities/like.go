// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package activities

import (
	apports "git.sr.ht/~bouncepaw/betula/ports/activitypub"
)

func guessLike(activity Dict) (any, error) {
	report := apports.LikeReport{
		ID:       getIDSomehow(activity, "id"),
		ActorID:  getIDSomehow(activity, "actor"),
		ObjectID: getIDSomehow(activity, "object"),
	}
	if activity["original activity"] != nil {
		report.Activity = activity["original activity"].([]byte)
	}
	if err := report.Valid(); err != nil {
		return nil, err
	}

	return report, nil
}
