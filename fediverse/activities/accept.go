// SPDX-FileCopyrightText: 2024 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2026 Danila Gorelko
//
// SPDX-License-Identifier: AGPL-3.0-only

package activities

type AcceptReport struct {
	ActorID  string
	ObjectID string
	Object   Dict
}

func guessAccept(activity Dict) (any, error) {
	report := AcceptReport{
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
		case Dict:
			report.Object = v
		}
	}

	return report, nil
}
