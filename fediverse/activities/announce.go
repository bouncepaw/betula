// SPDX-FileCopyrightText: 2023 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2024 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package activities

import (
	"git.sr.ht/~bouncepaw/betula/pkg/bxstr"
)

type AnnounceReport struct {
	ActorID    string
	AnnounceID string // id of the repost
	ObjectID   string // object that was reposted
}

func mustHaveSuchField[T any](activity Dict, field string, errOnLack error, lambdaOnPresence func(T)) error {
	if val, ok := activity[field]; !ok {
		return errOnLack
	} else {
		switch v := val.(type) {
		case T:
			lambdaOnPresence(v)
			return nil
		default:
			return errOnLack
		}
	}
}

func guessAnnounce(activity Dict) (reportMaybe any, err error) {
	report := AnnounceReport{
		ActorID:    getIDSomehow(activity, "actor"),
		AnnounceID: getIDSomehow(activity, "id"),
		ObjectID:   getIDSomehow(activity, "object"),
	}

	if !bxstr.IsValidURL(report.ObjectID) {
		return nil, ErrNoObject
	}

	if !bxstr.IsValidURL(report.AnnounceID) {
		return nil, ErrNoId
	}

	return report, nil
}
