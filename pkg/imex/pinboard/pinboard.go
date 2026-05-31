// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

// Package pinboard implements the Pinboard JSON format.
package pinboard

import (
	"encoding/json"
	"io"
	"strings"

	"git.sr.ht/~bouncepaw/betula/pkg/bxtime"
)

// Tags is a slice of tag strings that marshals to/from a single
// space-separated string in JSON, matching Pinboard's format.
type Tags []string

func (t Tags) MarshalJSON() ([]byte, error) {
	return json.Marshal(strings.Join(t, " "))
}

func (t *Tags) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s == "" {
		*t = nil
		return nil
	}
	*t = strings.Fields(s)
	return nil
}

// Bookmark represents a single Pinboard bookmark as it appears in a JSON export.
type Bookmark struct {
	Href        string             `json:"href"`
	Description string             `json:"description"`
	Extended    string             `json:"extended"`
	Meta        string             `json:"meta"`
	Hash        string             `json:"hash"`
	Time        bxtime.TimeRFC3339 `json:"time"`
	Shared      string             `json:"shared"` // "yes" or "no"
	ToRead      string             `json:"toread"` // "yes" or "no"
	Tags        Tags               `json:"tags"`
}

// Probe reports whether r contains a Pinboard JSON export by checking that it
// starts with [ and contains "href". It seeks r back to the start before
// returning, so the caller can pass the same reader to Read.
func Probe(r io.ReadSeeker) (bool, error) {
	buf := make([]byte, 512)
	n, err := r.Read(buf)
	if err != nil && err != io.EOF {
		return false, err
	}
	if _, err = r.Seek(0, io.SeekStart); err != nil {
		return false, err
	}
	s := strings.TrimSpace(string(buf[:n]))
	return strings.HasPrefix(s, "[") && strings.Contains(s, `"href"`), nil
}

// Read parses a Pinboard JSON export from r and returns the bookmarks.
func Read(r io.Reader) ([]Bookmark, error) {
	var bookmarks []Bookmark
	return bookmarks, json.NewDecoder(r).Decode(&bookmarks)
}

// Write encodes bookmarks as a Pinboard JSON export and writes it to w.
func Write(bookmarks []Bookmark, w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "\t")
	return enc.Encode(bookmarks)
}
