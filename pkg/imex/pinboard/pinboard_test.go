// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package pinboard

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/nalgeon/be"

	"git.sr.ht/~bouncepaw/betula/pkg/bxtime"
)

func TestRead(t *testing.T) {
	t.Parallel()

	f, err := os.Open("testdata/pinboard1.json")
	be.Equal(t, err, nil)
	defer f.Close()

	bookmarks, err := Read(f)
	be.Equal(t, err, nil)
	be.Equal(t, len(bookmarks), 5)

	codeberg := bookmarks[0]
	be.Equal(t, codeberg.Href, "https://codeberg.org")
	be.Equal(t, codeberg.Description, "Codeberg")
	be.Equal(t, codeberg.Extended, "")
	be.Equal(t, codeberg.Tags, Tags(nil))
	be.Equal(t, codeberg.Time, bxtime.TimeRFC3339(time.Date(2024, 1, 10, 8, 0, 0, 0, time.UTC)))
	be.Equal(t, codeberg.Shared, "yes")
	be.Equal(t, codeberg.ToRead, "no")

	sourcehut := bookmarks[1]
	be.Equal(t, sourcehut.Href, "https://sourcehut.org/")
	be.Equal(t, sourcehut.Description, "Sourcehut")
	be.Equal(t, sourcehut.Extended, "This suite of open source tools is the software development platform you've been waiting for.")
	be.Equal(t, sourcehut.Tags, Tags{"forge", "git"})
	be.Equal(t, sourcehut.Shared, "yes")
	be.Equal(t, sourcehut.ToRead, "no")

	bouncepaw := bookmarks[2]
	be.Equal(t, bouncepaw.Href, "https://bouncepaw.com/")
	be.Equal(t, bouncepaw.Tags, Tags{"tag1", "tag2"})
	be.Equal(t, bouncepaw.Shared, "no")

	mycorrhiza := bookmarks[3]
	be.Equal(t, mycorrhiza.Href, "https://mycorrhiza.wiki/")
	be.Equal(t, mycorrhiza.Extended, "A wiki engine.")
	be.Equal(t, mycorrhiza.Shared, "no")
	be.Equal(t, mycorrhiza.ToRead, "yes")

	betula := bookmarks[4]
	be.Equal(t, betula.Href, "https://joinbetula.org/")
	be.Equal(t, betula.Description, "Betula")
	be.Equal(t, betula.Tags, Tags{"bookmarks", "software"})
	be.Equal(t, betula.Shared, "yes")
}

func TestWrite(t *testing.T) {
	t.Parallel()

	bookmarks := []Bookmark{
		{
			Href:        "https://example.com/",
			Description: "Example",
			Extended:    "Some notes.",
			Meta:        "aaaa",
			Hash:        "bbbb",
			Time:        bxtime.TimeRFC3339(time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC)),
			Shared:      "yes",
			ToRead:      "no",
			Tags:        Tags{"foo", "bar"},
		},
		{
			Href:        "https://example.org/",
			Description: "Example Org",
			Extended:    "",
			Meta:        "cccc",
			Hash:        "dddd",
			Time:        bxtime.TimeRFC3339(time.Date(2024, 7, 1, 12, 0, 0, 0, time.UTC)),
			Shared:      "no",
			ToRead:      "no",
			Tags:        nil,
		},
	}

	var buf bytes.Buffer
	err := Write(bookmarks, &buf)
	be.Equal(t, err, nil)

	out := buf.String()
	be.True(t, strings.Contains(out, `"href": "https://example.com/"`))
	be.True(t, strings.Contains(out, `"description": "Example"`))
	be.True(t, strings.Contains(out, `"extended": "Some notes."`))
	be.True(t, strings.Contains(out, `"tags": "foo bar"`))
	be.True(t, strings.Contains(out, `"shared": "yes"`))
	be.True(t, strings.Contains(out, `"time": "2024-06-01T10:00:00Z"`))
	be.True(t, strings.Contains(out, `"href": "https://example.org/"`))
	be.True(t, strings.Contains(out, `"shared": "no"`))
	be.True(t, strings.Contains(out, `"tags": ""`))
}

func TestRoundtrip(t *testing.T) {
	t.Parallel()

	original := []Bookmark{
		{
			Href:        "https://go.dev/",
			Description: "Go",
			Extended:    "The Go programming language.",
			Meta:        "meta1",
			Hash:        "hash1",
			Time:        bxtime.TimeRFC3339(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
			Shared:      "yes",
			ToRead:      "no",
			Tags:        Tags{"lang", "tools"},
		},
		{
			Href:        "https://pkg.go.dev/",
			Description: "Go Packages",
			Extended:    "",
			Meta:        "meta2",
			Hash:        "hash2",
			Time:        bxtime.TimeRFC3339(time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)),
			Shared:      "no",
			ToRead:      "yes",
			Tags:        nil,
		},
	}

	var buf bytes.Buffer
	be.Equal(t, Write(original, &buf), nil)

	got, err := Read(&buf)
	be.Equal(t, err, nil)
	be.Equal(t, len(got), 2)

	be.Equal(t, got[0].Href, original[0].Href)
	be.Equal(t, got[0].Description, original[0].Description)
	be.Equal(t, got[0].Extended, original[0].Extended)
	be.Equal(t, got[0].Tags, original[0].Tags)
	be.Equal(t, got[0].Time, original[0].Time)
	be.Equal(t, got[0].Shared, original[0].Shared)
	be.Equal(t, got[0].ToRead, original[0].ToRead)

	be.Equal(t, got[1].Href, original[1].Href)
	be.Equal(t, got[1].Tags, Tags(nil))
	be.Equal(t, got[1].Shared, original[1].Shared)
	be.Equal(t, got[1].ToRead, original[1].ToRead)
}

func TestProbe(t *testing.T) {
	t.Parallel()

	t.Run("pinboard1", func(t *testing.T) {
		t.Parallel()
		f, err := os.Open("testdata/pinboard1.json")
		be.Equal(t, err, nil)
		defer f.Close()

		ok, err := Probe(f)
		be.Equal(t, err, nil)
		be.Equal(t, ok, true)

		bookmarks, err := Read(f)
		be.Equal(t, err, nil)
		be.Equal(t, len(bookmarks), 5)
	})

	t.Run("Not Pinboard", func(t *testing.T) {
		t.Parallel()
		r := strings.NewReader(`{"key": "value"}`)
		ok, err := Probe(r)
		be.Equal(t, err, nil)
		be.Equal(t, ok, false)
	})
}
