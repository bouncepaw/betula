// SPDX-FileCopyrightText: 2023 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2024 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package activities

import (
	apports "git.sr.ht/~bouncepaw/betula/ports/activitypub"
)

func guessUndo(activity Dict) (reportMaybe any, err error) {
	var (
		report    apports.UndoAnnounceReport
		objectMap Dict
	)

	if err := mustHaveSuchField(
		activity, "object", ErrNoObject,
		func(v map[string]any) {
			objectMap = v
		},
	); err != nil {
		return nil, err
	}

	switch objectMap["type"] {
	case "Announce":
		switch repost := objectMap["id"].(type) {
		case string:
			report.AnnounceID = repost
		}
		switch original := objectMap["object"].(type) {
		case string:
			report.ObjectID = original
		}
		switch actor := objectMap["actor"].(type) {
		case Dict:
			switch username := actor["preferredUsername"].(type) {
			case string:
				report.ActorID = username
			}
		}
		return report, nil

	case "Follow":
		if objectMap == nil {
			return nil, ErrNoObject
		}
		followReport, err := guessFollow(objectMap)
		if err != nil {
			return nil, err
		}
		return apports.UndoFollowReport{FollowReport: followReport.(apports.FollowReport)}, nil

	case "Like":
		if objectMap == nil {
			return nil, ErrNoObject
		}
		likeReport, err := guessLike(objectMap)
		if err != nil {
			return nil, err
		}
		return apports.UndoLikeReport{
			ID:       getIDSomehow(activity, "id"),
			Object:   likeReport.(apports.LikeReport),
			Activity: activity["original activity"].([]byte),
		}, nil

	default:
		return nil, ErrUnknownType
	}
}
