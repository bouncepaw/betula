// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package raindrop

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/nalgeon/be"
)

// testdata/raindrop1.csv produced by importing pinboard1.json into Raindrop and exporting it.
// Hence, the shared bookmarks and the "pinboard1" folder on every row.
func TestRead(t *testing.T) {
	t.Parallel()

	f, err := os.Open("testdata/raindrop1.csv")
	be.Equal(t, err, nil)
	defer f.Close()

	bookmarks, err := Read(f)
	be.Equal(t, err, nil)
	be.Equal(t, len(bookmarks), 5)

	betula := bookmarks[0]
	be.Equal(t, betula.ID, "1739704207")
	be.Equal(t, betula.Title, "Betula")
	be.Equal(t, betula.Note, "")
	be.Equal(t, betula.Excerpt, "")
	be.Equal(t, betula.URL, "https://joinbetula.org/")
	be.Equal(t, betula.Folder, "pinboard1")
	be.Equal(t, betula.Tags, []string{"bookmarks", "software"})
	be.Equal(t, betula.Created, time.Date(2024, 5, 1, 14, 0, 0, 0, time.UTC))
	be.Equal(t, betula.Cover, "")
	be.Equal(t, betula.Highlights, "")
	be.Equal(t, betula.Favorite, false)

	sourcehut := bookmarks[3]
	be.Equal(t, sourcehut.Title, "Sourcehut")
	be.Equal(t, sourcehut.Note, "This suite of open source tools is the software development platform you've been waiting for.")
	be.Equal(t, sourcehut.URL, "https://sourcehut.org/")
	be.Equal(t, sourcehut.Tags, []string{"forge", "git"})

	codeberg := bookmarks[4]
	be.Equal(t, codeberg.Title, "Codeberg")
	be.Equal(t, codeberg.URL, "https://codeberg.org/")
	be.Equal(t, codeberg.Tags, []string(nil))
}

func TestWrite(t *testing.T) {
	t.Parallel()

	bookmarks := []Bookmark{
		{
			ID:       "1",
			Title:    "Example",
			Note:     "Some notes.",
			Excerpt:  "An example domain.",
			URL:      "https://example.com/",
			Folder:   "work",
			Tags:     []string{"foo", "bar"},
			Created:  time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC),
			Cover:    "https://example.com/cover.png",
			Favorite: true,
		},
		{
			ID:    "2",
			Title: "Example Org",
			URL:   "https://example.org/",
		},
	}

	var buf bytes.Buffer
	err := Write(bookmarks, &buf)
	be.Equal(t, err, nil)

	out := buf.String()
	be.True(t, strings.Contains(out, "id,title,note,excerpt,url,folder,tags,created,cover,highlights,favorite"))
	be.True(t, strings.Contains(out, "https://example.com/"))
	be.True(t, strings.Contains(out, "work"))
	be.True(t, strings.Contains(out, `"foo, bar"`))
	be.True(t, strings.Contains(out, "2024-06-01T10:00:00Z"))
	be.True(t, strings.Contains(out, "true"))
	be.True(t, strings.Contains(out, "https://example.org/"))
}

func TestRoundtrip(t *testing.T) {
	t.Parallel()

	original := []Bookmark{
		{
			ID:       "1",
			Title:    "Go",
			Note:     "A note.",
			Excerpt:  "The Go programming language.",
			URL:      "https://go.dev/",
			Folder:   "langs",
			Tags:     []string{"lang", "tools"},
			Created:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			Cover:    "https://go.dev/cover.png",
			Favorite: true,
		},
		{
			ID:      "2",
			Title:   "Go Packages",
			URL:     "https://pkg.go.dev/",
			Created: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	var buf bytes.Buffer
	be.Equal(t, Write(original, &buf), nil)

	got, err := Read(&buf)
	be.Equal(t, err, nil)
	be.Equal(t, len(got), 2)

	be.Equal(t, got[0].ID, original[0].ID)
	be.Equal(t, got[0].Title, original[0].Title)
	be.Equal(t, got[0].Note, original[0].Note)
	be.Equal(t, got[0].Excerpt, original[0].Excerpt)
	be.Equal(t, got[0].URL, original[0].URL)
	be.Equal(t, got[0].Folder, original[0].Folder)
	be.Equal(t, got[0].Tags, original[0].Tags)
	be.Equal(t, got[0].Created, original[0].Created)
	be.Equal(t, got[0].Cover, original[0].Cover)
	be.Equal(t, got[0].Favorite, original[0].Favorite)

	be.Equal(t, got[1].URL, original[1].URL)
	be.Equal(t, got[1].Tags, []string(nil))
	be.Equal(t, got[1].Favorite, false)
	be.Equal(t, got[1].Created, original[1].Created)
}

func TestProbe(t *testing.T) {
	t.Parallel()

	t.Run("raindrop1", func(t *testing.T) {
		t.Parallel()
		f, err := os.Open("testdata/raindrop1.csv")
		be.Equal(t, err, nil)
		defer f.Close()

		ok, err := Probe(f)
		be.Equal(t, err, nil)
		be.Equal(t, ok, true)

		bookmarks, err := Read(f)
		be.Equal(t, err, nil)
		be.Equal(t, len(bookmarks), 5)
	})

	t.Run("Not Raindrop", func(t *testing.T) {
		t.Parallel()
		r := strings.NewReader("name,age\nAlice,30\n")
		ok, err := Probe(r)
		be.Equal(t, err, nil)
		be.Equal(t, ok, false)
	})
}
