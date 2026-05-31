// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package bxtime

import (
	"strconv"
	"time"
)

// ParseUnixTimestamp parses a decimal Unix timestamp string into a time.Time.
// Returns zero time on parse failure.
func ParseUnixTimestamp(s string) time.Time {
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return time.Time{}
	}
	return time.Unix(n, 0)
}

// FormatUnixTimestamp formats t as a decimal Unix timestamp string.
func FormatUnixTimestamp(t time.Time) string {
	return strconv.FormatInt(t.Unix(), 10)
}
