// SPDX-FileCopyrightText: 2023 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2024 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2026 Danila Gorelko
// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package parsing

import (
	apports "git.sr.ht/~bouncepaw/betula/ports/activitypub"
)

type FollowParser struct{}

var _ apports.FollowParser = (*FollowParser)(nil)

func NewFollowParser() *FollowParser {
	return &FollowParser{}
}

func (p *FollowParser) GuessFollow(activity apports.Dict) (any, error) {
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

func (p *FollowParser) GuessAccept(activity apports.Dict) (any, error) {
	report := apports.AcceptReport{
		ActorID:  getIDSomehow(activity, "actor"),
		ObjectID: getIDSomehow(activity, "object"),
	}
	if report.ActorID == "" {
		return nil, ErrNoActor
	}
	if report.ObjectID == "" {
		return nil, ErrNoObject
	}
	if obj, ok := activity["object"]; ok {
		switch v := obj.(type) {
		case apports.Dict:
			report.Object = v
		}
	}

	return report, nil
}

func (p *FollowParser) GuessReject(activity apports.Dict) (any, error) {
	report := apports.RejectReport{
		ActorID:  getIDSomehow(activity, "actor"),
		ObjectID: getIDSomehow(activity, "object"),
	}
	if report.ActorID == "" {
		return nil, ErrNoActor
	}
	if report.ObjectID == "" {
		return nil, ErrNoObject
	}
	if obj, ok := activity["object"]; ok {
		switch v := obj.(type) {
		case apports.Dict:
			report.Object = v
		}
	}

	return report, nil
}

func (p *FollowParser) GuessUndoFollow(object apports.Dict) (any, error) {
	followReport, err := p.GuessFollow(object)
	if err != nil {
		return nil, err
	}
	return apports.UndoFollowReport{FollowReport: followReport.(apports.FollowReport)}, nil
}
