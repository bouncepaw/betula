// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

// Package raindrop implements the Raindrop.io CSV format.
package raindrop

import (
	"encoding/csv"
	"io"
	"strconv"
	"strings"
	"time"
)

type Bookmark struct {
	ID         string
	Title      string
	Note       string
	Excerpt    string
	URL        string
	Folder     string
	Tags       []string
	Created    time.Time
	Cover      string
	Highlights string
	Favorite   bool
}

// Probe reports whether r contains a Raindrop.io CSV export by checking for its
// header row. It seeks r back to the start before returning, so the caller can
// pass the same reader to Read.
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
	return strings.HasPrefix(s, "id,title,note,excerpt,url"), nil
}

// Read parses a Raindrop.io CSV export from r and returns the bookmarks.
func Read(r io.Reader) ([]Bookmark, error) {
	cr := csv.NewReader(r)
	records, err := cr.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, nil
	}

	idx := make(map[string]int, len(records[0]))
	for i, name := range records[0] {
		idx[strings.TrimSpace(name)] = i
	}
	get := func(rec []string, name string) string {
		i, ok := idx[name]
		if !ok || i >= len(rec) {
			return ""
		}
		return rec[i]
	}

	bookmarks := make([]Bookmark, 0, len(records)-1)
	for _, rec := range records[1:] {
		b := Bookmark{
			ID:         get(rec, "id"),
			Title:      get(rec, "title"),
			Note:       get(rec, "note"),
			Excerpt:    get(rec, "excerpt"),
			URL:        get(rec, "url"),
			Folder:     get(rec, "folder"),
			Tags:       parseTags(get(rec, "tags")),
			Cover:      get(rec, "cover"),
			Highlights: get(rec, "highlights"),
			Favorite:   get(rec, "favorite") == "true",
		}
		if created := get(rec, "created"); created != "" {
			if t, err := time.Parse(time.RFC3339, created); err == nil {
				b.Created = t
			}
		}
		bookmarks = append(bookmarks, b)
	}
	return bookmarks, nil
}

// Column order. Used by Write. Read accepts any order.
var header = []string{"id", "title", "note", "excerpt", "url", "folder", "tags", "created", "cover", "highlights", "favorite"}

// Write encodes bookmarks as a Raindrop.io CSV export and writes it to w.
func Write(bookmarks []Bookmark, w io.Writer) error {
	cw := csv.NewWriter(w)
	if err := cw.Write(header); err != nil {
		return err
	}
	for _, b := range bookmarks {
		created := ""
		if !b.Created.IsZero() {
			created = b.Created.UTC().Format(time.RFC3339)
		}
		rec := []string{
			b.ID,
			b.Title,
			b.Note,
			b.Excerpt,
			b.URL,
			b.Folder,
			strings.Join(b.Tags, ", "),
			created,
			b.Cover,
			b.Highlights,
			strconv.FormatBool(b.Favorite),
		}
		if err := cw.Write(rec); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}

// parseTags splits a comma-separated tags field into individual tags, trimming
// surrounding whitespace and dropping empty entries.
func parseTags(s string) []string {
	var tags []string
	for t := range strings.SplitSeq(s, ",") {
		if t = strings.TrimSpace(t); t != "" {
			tags = append(tags, t)
		}
	}
	return tags
}
