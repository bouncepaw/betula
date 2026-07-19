// SPDX-FileCopyrightText: 2024 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package activities

import (
	apports "git.sr.ht/~bouncepaw/betula/ports/activitypub"
)

func guessFollow(activity Dict) (any, error) {
	report := apports.FollowReport{
		ActorID:          getIDSomehow(activity, "actor"),
		ObjectID:         getIDSomehow(activity, "object"),
		OriginalActivity: activity,
	}
	if report.ActorID == "" {
		return nil, ErrNoActor
	}
	if report.ObjectID == "" {
		return nil, ErrNoObject
	}
	return report, nil
}
