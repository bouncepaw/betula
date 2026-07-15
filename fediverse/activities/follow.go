// SPDX-FileCopyrightText: 2024 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package activities

type FollowReport struct {
	ActorID          string
	ObjectID         string
	OriginalActivity Dict
}

func guessFollow(activity Dict) (any, error) {
	report := FollowReport{
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
