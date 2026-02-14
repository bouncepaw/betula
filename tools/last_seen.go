// SPDX-FileCopyrightText: 2024 Danila Gorelko
//
// SPDX-License-Identifier: AGPL-3.0-only

package tools

import (
	"fmt"
	"time"
)

func LastSeen(from, to time.Time) string {

	format := func(s string, count int) string {
		return fmt.Sprintf("%s ago", pluralize(s, count))
	}

	diff := to.Sub(from)

	if diff.Seconds() < 1 {
		return "just now"
	}
	if diff.Minutes() < 1 {
		return format("second", int(diff.Seconds()))
	}
	if diff.Hours() < 1 {
		return format("minute", int(diff.Minutes()))
	}
	if diff.Hours() < 24 {
		return format("hour", int(diff.Hours()))
	}
	if diff.Hours() < 24*7 {
		return format("day", int(diff.Hours()/24))
	}
	if diff.Hours() == 24*7 {
		return "a week ago"
	}
	return fmt.Sprintf("on %s", from.Format("Monday, January 2, 2006"))
}

func pluralize(s string, count int) string {
	if count == 1 {
		if s == "hour" {
			return fmt.Sprintf("an %s", s)
		}
		return fmt.Sprintf("a %s", s)
	}
	return fmt.Sprintf("%d %ss", count, s)
}
