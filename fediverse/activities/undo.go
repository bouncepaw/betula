// SPDX-FileCopyrightText: 2023 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2024 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package activities

import (
	"encoding/json"
)

type UndoAnnounceReport struct {
	AnnounceReport
}

type UndoFollowReport struct {
	FollowReport
}

type UndoLikeReport struct {
	ID       string
	Object   LikeReport
	Activity json.RawMessage
}

func guessUndo(activity Dict) (reportMaybe any, err error) {
	var (
		report    UndoAnnounceReport
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
			report.RepostPage = repost
		}
		switch original := objectMap["object"].(type) {
		case string:
			report.OriginalPage = original
		}
		switch actor := objectMap["actor"].(type) {
		case Dict:
			switch username := actor["preferredUsername"].(type) {
			case string:
				report.ReposterUsername = username
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
		return UndoFollowReport{followReport.(FollowReport)}, nil

	case "Like":
		if objectMap == nil {
			return nil, ErrNoObject
		}
		likeReport, err := guessLike(objectMap)
		if err != nil {
			return nil, err
		}
		return UndoLikeReport{
			ID:       getIDSomehow(activity, "id"),
			Object:   likeReport.(LikeReport),
			Activity: activity["original activity"].([]byte),
		}, nil

	default:
		return nil, ErrUnknownType
	}
}
