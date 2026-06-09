// SPDX-FileCopyrightText: 2026 Danila Gorelko
// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

// Package bxstr provides common string operations that looked like they belong here.
package bxstr

import (
	"crypto/rand"
	"encoding/hex"
	"net/url"
	"strings"
	"unicode"
)

func ValidURL(s string) bool {
	_, err := url.ParseRequestURI(s)
	return err == nil
}

func ParseValidURL(s string) *url.URL {
	u, err := url.ParseRequestURI(s)
	if err != nil {
		panic(err)
	}
	return u
}

func SameHost(s1, s2 string) bool {
	u1, err1 := url.ParseRequestURI(s1)
	u2, err2 := url.ParseRequestURI(s2)
	return err1 == nil && err2 == nil && u1.Host == u2.Host
}

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
