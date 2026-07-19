// SPDX-FileCopyrightText: 2023 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2024 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

// Package activities provides activity data extraction from JSON.
package activities

import (
	"errors"

	apports "git.sr.ht/~bouncepaw/betula/ports/activitypub"
)

func getString(activity Dict, field string) string {
	m := activity[field]
	switch v := m.(type) {
	case string:
		return v
	}
	return ""
}

type Dict = apports.Dict

var (
	ErrNoType      = errors.New("activities: type absent or invalid")
	ErrUnknownType = errors.New("activities: unknown activity type")
	ErrNoObject    = errors.New("activities: object absent or invalid")
)
