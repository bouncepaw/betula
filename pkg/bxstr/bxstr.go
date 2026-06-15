// SPDX-FileCopyrightText: 2024 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2026 Danila Gorelko
// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

// Package bxstr provides common string operations that looked like they belong here.
package bxstr

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
	"unicode"
)

func StringifyAnything(o any) string {
	switch s := o.(type) {
	case string:
		return s
	default:
		return ""
	}
}

func RandomWhatever() string {
	b := make([]byte, 20)
	rand.Read(b)
	return hex.EncodeToString(b)[2:20]
}

func TrimRightSpace(s string) string {
	return strings.TrimRightFunc(s, unicode.IsSpace)
}

// CommaSeparated takes a comma-separated sequence of words,
// trims every word from whitespace, and returns them.
//
// The suggested use is parsing tag names in export formats.
func CommaSeparated(s string) []string {
	var words []string
	for w := range strings.SplitSeq(s, ",") {
		if w = strings.TrimSpace(w); w != "" {
			words = append(words, w)
		}
	}
	return words
}
