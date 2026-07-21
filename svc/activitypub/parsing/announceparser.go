// SPDX-FileCopyrightText: 2023 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2024 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package parsing

import (
	"git.sr.ht/~bouncepaw/betula/pkg/bxstr"
	apports "git.sr.ht/~bouncepaw/betula/ports/activitypub"
)

type AnnounceParser struct{}

var _ apports.AnnounceParser = (*AnnounceParser)(nil)

func NewAnnounceParser() *AnnounceParser {
	return &AnnounceParser{}
}

func (p *AnnounceParser) GuessAnnounce(activity apports.Dict) (any, error) {
	report := apports.AnnounceReport{
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

func (p *AnnounceParser) GuessUndoAnnounce(object apports.Dict) (any, error) {
	var report apports.UndoAnnounceReport
	switch remark := object["id"].(type) {
	case string:
		report.AnnounceID = remark
	}
	switch original := object["object"].(type) {
	case string:
		report.ObjectID = original
	}
	switch actor := object["actor"].(type) {
	case apports.Dict:
		switch username := actor["preferredUsername"].(type) {
		case string:
			report.ActorID = username
		}
	}
	return report, nil
}
