// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package importers

import (
	"errors"
	"iter"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/nalgeon/be"

	wwwports "git.sr.ht/~bouncepaw/betula/ports/www"
	"git.sr.ht/~bouncepaw/betula/types"
)

type fakeWWW struct {
	titles map[string]string
}

func (f fakeWWW) TitleOfPage(addr string) (string, error) {
	return f.titles[addr], nil
}

func (f fakeWWW) RelAlternates(addr string) ([]wwwports.RelAlternate, error) {
	return nil, nil
}

type fakeErringReader struct {
	err error
}

func (r fakeErringReader) Read([]byte) (int, error) {
	return 0, r.err
}

func collectSortURLs(seq iter.Seq[string]) []string {
	urls := slices.Collect(seq)
	slices.Sort(urls)
	return urls
}

func collectSortBookmarks(seq iter.Seq2[types.Bookmark, error]) ([]types.Bookmark, error) {
	var bookmarks []types.Bookmark
	for bm, err := range seq {
		if err != nil {
			return nil, err
		}
		bookmarks = append(bookmarks, bm)
	}
	slices.SortFunc(bookmarks, func(a, b types.Bookmark) int {
		return strings.Compare(a.URL, b.URL)
	})
	return bookmarks, nil
}

func TestExtractURLs(t *testing.T) {
	t.Parallel()

	t.Run("Multiple schemes", func(t *testing.T) {
		t.Parallel()
		input := strings.NewReader(`
Check out https://example.com/ and http://example.org/path.
Also gemini://gemini.example/ and gopher://gopher.example/1
`)
		seq, err := extractURLs(input)
		be.Equal(t, err, nil)
		be.Equal(t, collectSortURLs(seq), []string{
			"gemini://gemini.example/",
			"gopher://gopher.example/1",
			"http://example.org/path",
			"https://example.com/",
		})
	})

	t.Run("Trimmed trailing punctuation", func(t *testing.T) {
		t.Parallel()
		input := strings.NewReader(`See https://example.com/page.,;:!?(<{[]}>).`)
		seq, err := extractURLs(input)
		be.Equal(t, err, nil)
		be.Equal(t, collectSortURLs(seq), []string{"https://example.com/page"})
	})

	t.Run("Deduplicated", func(t *testing.T) {
		t.Parallel()
		input := strings.NewReader(`https://example.com/ and again https://example.com/.`)
		seq, err := extractURLs(input)
		be.Equal(t, err, nil)
		be.Equal(t, collectSortURLs(seq), []string{"https://example.com/"})
	})

	t.Run("Ignore invalid URLs", func(t *testing.T) {
		t.Parallel()
		input := strings.NewReader(`https://example.org/ not-a-url http://`)
		seq, err := extractURLs(input)
		be.Equal(t, err, nil)
		be.Equal(t, collectSortURLs(seq), []string{"https://example.org/"})
	})

	t.Run("Empty input", func(t *testing.T) {
		t.Parallel()
		seq, err := extractURLs(strings.NewReader(""))
		be.Equal(t, err, nil)
		be.Equal(t, collectSortURLs(seq), []string(nil))
	})

	t.Run("Read error", func(t *testing.T) {
		t.Parallel()
		seq, err := extractURLs(fakeErringReader{err: errors.New("read failed")})
		be.Equal(t, seq, nil)
		be.True(t, err != nil)
	})
}

func TestPlainImporterProbe(t *testing.T) {
	t.Parallel()

	imp := NewPlainImporter(1, fakeWWW{})
	ok, err := imp.Probe(strings.NewReader("anything"))
	be.Equal(t, err, nil)
	be.Equal(t, ok, true)
}

func TestPlainImporterImport(t *testing.T) {
	t.Parallel()

	t.Run("Use fetched titles", func(t *testing.T) {
		t.Parallel()
		imp := NewPlainImporter(2, fakeWWW{
			titles: map[string]string{
				"https://example.com/": "Example",
				"https://go.dev/":      "Go",
			},
		})

		input := strings.NewReader(`https://example.com/ https://go.dev/`)
		seq, err := imp.Import(input)
		be.Equal(t, err, nil)

		bookmarks, err := collectSortBookmarks(seq)
		be.Equal(t, err, nil)
		be.Equal(t, len(bookmarks), 2)

		be.Equal(t, bookmarks[0].URL, "https://example.com/")
		be.Equal(t, bookmarks[0].Title, "Example")
		be.Equal(t, bookmarks[0].Visibility, types.Private)

		be.Equal(t, bookmarks[1].URL, "https://go.dev/")
		be.Equal(t, bookmarks[1].Title, "Go")
		be.Equal(t, bookmarks[1].Visibility, types.Private)

		for _, bm := range bookmarks {
			_, parseErr := time.Parse(types.TimeLayout, bm.CreationTime)
			be.Equal(t, parseErr, nil)
		}
	})

	t.Run("URL fallback", func(t *testing.T) {
		t.Parallel()
		imp := NewPlainImporter(1, fakeWWW{})

		input := strings.NewReader(`https://example.com/path/?q=1`)
		seq, err := imp.Import(input)
		be.Equal(t, err, nil)

		bookmarks, err := collectSortBookmarks(seq)
		be.Equal(t, err, nil)
		be.Equal(t, len(bookmarks), 1)
		be.Equal(t, bookmarks[0].Title, types.CleanerLink("https://example.com/path/?q=1"))
	})

	t.Run("Extract error", func(t *testing.T) {
		t.Parallel()
		imp := NewPlainImporter(1, fakeWWW{})

		seq, err := imp.Import(fakeErringReader{err: errors.New("read failed")})
		be.Equal(t, seq, nil)
		be.True(t, err != nil)
	})
}
