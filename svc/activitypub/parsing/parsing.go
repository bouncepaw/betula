// SPDX-FileCopyrightText: 2023 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2024 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

// Package parsing parses incoming ActivityPub objects and activities into
// Betula's domain types. It is the inbound counterpart to package assembly.
package parsing

import (
	"errors"
	"time"

	"git.sr.ht/~bouncepaw/betula/pkg/bxstr"
	apports "git.sr.ht/~bouncepaw/betula/ports/activitypub"
	"git.sr.ht/~bouncepaw/betula/types"
)

var (
	ErrNotNote      = errors.New("not a Note")
	ErrEmptyField   = errors.New("empty field")
	ErrHostMismatch = errors.New("host mismatch")
	ErrNoObject     = errors.New("object absent or invalid")
	ErrNoActor      = errors.New("actor absent or invalid")
	ErrNoId         = errors.New("id absent or invalid")
)

func getIDSomehow(activity apports.Dict, field string) string {
	m, ok := activity[field]
	if !ok {
		return ""
	}
	switch v := m.(type) {
	case string:
		if bxstr.IsValidURL(v) {
			return v
		}
		return ""
	}
	for k, v := range m.(apports.Dict) {
		if k != "id" {
			continue
		}
		switch v := v.(type) {
		case string:
			return v
		default:
			return ""
		}
	}
	return ""
}

func getTime(object apports.Dict, field string) string {
	rfc3339 := getString(object, field)
	t, err := time.Parse(time.RFC3339, rfc3339)
	if err != nil {
		return ""
	}
	return t.Format(types.TimeLayout)
}

func getString(activity apports.Dict, field string) string {
	m := activity[field]
	switch v := m.(type) {
	case string:
		return v
	}
	return ""
}
