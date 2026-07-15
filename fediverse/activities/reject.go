// SPDX-FileCopyrightText: 2023 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2024 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package activities

type RejectReport struct {
	ActorID  string
	ObjectID string
	Object   Dict
}

func guessReject(activity Dict) (any, error) {
	report := RejectReport{
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
