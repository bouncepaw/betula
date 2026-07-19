// SPDX-FileCopyrightText: 2023 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2024 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package parsing

import (
	apports "git.sr.ht/~bouncepaw/betula/ports/activitypub"
)

type LikeParser struct{}

var _ apports.LikeParser = (*LikeParser)(nil)

func NewLikeParser() *LikeParser {
	return &LikeParser{}
}

func (p *LikeParser) GuessLike(activity apports.Dict) (any, error) {
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

func (p *LikeParser) GuessUndoLike(activity, object apports.Dict) (any, error) {
	likeReport, err := p.GuessLike(object)
	if err != nil {
		return nil, err
	}
	return apports.UndoLikeReport{
		ID:       getIDSomehow(activity, "id"),
		Object:   likeReport.(apports.LikeReport),
		Activity: activity["original activity"].([]byte),
	}, nil
}
