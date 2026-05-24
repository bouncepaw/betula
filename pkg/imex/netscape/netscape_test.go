// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package netscape

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/nalgeon/be"
)

func TestRead(t *testing.T) {
	t.Parallel()

	t.Run("firefox1", func(t *testing.T) {
		t.Parallel()
		f, err := os.Open("testdata/firefox1.html")
		be.Equal(t, err, nil)
		defer f.Close()

		root, err := Read(f)
		be.Equal(t, err, nil)

		be.Equal(t, root.Title, "Меню закладок")
		be.Equal(t, len(root.Items), 3)

		mozFolder, ok := root.Items[0].(*Folder)
		be.True(t, ok)
		be.Equal(t, mozFolder.Title, "Mozilla Firefox")
		be.Equal(t, mozFolder.Added, time.Unix(1770036858, 0))
		be.Equal(t, len(mozFolder.Items), 5)

		betula, ok := mozFolder.Items[4].(Bookmark)
		be.True(t, ok)
		be.Equal(t, betula.URL, "https://joinbetula.org/")
		be.Equal(t, betula.Title, "Betula")
		be.Equal(t, betula.Tags, []string(nil))
		be.Equal(t, betula.Added, time.Unix(1775421369, 0))
		be.Equal(t, betula.Modified, time.Unix(1775421369, 0))

		toolbarFolder, ok := root.Items[1].(*Folder)
		be.True(t, ok)
		be.Equal(t, toolbarFolder.Title, "Панель закладок")
		be.Equal(t, len(toolbarFolder.Items), 2)

		bouncepaw, ok := toolbarFolder.Items[0].(Bookmark)
		be.True(t, ok)
		be.Equal(t, bouncepaw.URL, "https://bouncepaw.com/")
		be.Equal(t, bouncepaw.Title, "Bouncepaw")
		be.Equal(t, bouncepaw.Tags, []string{"tag1", "tag2"})
		be.Equal(t, bouncepaw.Added, time.Unix(1775421246, 0))

		folderA, ok := toolbarFolder.Items[1].(*Folder)
		be.True(t, ok)
		be.Equal(t, folderA.Title, "Folder A")
		be.Equal(t, len(folderA.Items), 2)

		folderAA, ok := folderA.Items[1].(*Folder)
		be.True(t, ok)
		be.Equal(t, folderAA.Title, "Folder A/A")
		be.Equal(t, len(folderAA.Items), 1)

		exampleNet, ok := folderAA.Items[0].(Bookmark)
		be.True(t, ok)
		be.Equal(t, exampleNet.URL, "https://example.net/")
		be.Equal(t, exampleNet.Tags, []string{"tag1"})

		otherFolder, ok := root.Items[2].(*Folder)
		be.True(t, ok)
		be.Equal(t, otherFolder.Title, "Другие закладки")
		be.Equal(t, len(otherFolder.Items), 1)
	})

	t.Run("safari1", func(t *testing.T) {
		t.Parallel()
		f, err := os.Open("testdata/safari1.html")
		be.Equal(t, err, nil)
		defer f.Close()

		root, err := Read(f)
		be.Equal(t, err, nil)

		be.Equal(t, root.Title, "Bookmarks")
		be.Equal(t, len(root.Items), 3)

		favourites, ok := root.Items[0].(*Folder)
		be.True(t, ok)
		be.Equal(t, favourites.Title, "Favourites")
		be.Equal(t, favourites.IsReadingList, false)
		be.Equal(t, len(favourites.Items), 4)

		saveLink, ok := favourites.Items[0].(Bookmark)
		be.True(t, ok)
		be.Equal(t, saveLink.Title, "Save link")
		be.True(t, strings.HasPrefix(saveLink.URL, "javascript:"))

		codeberg, ok := favourites.Items[2].(Bookmark)
		be.True(t, ok)
		be.Equal(t, codeberg.URL, "https://codeberg.org")
		be.Equal(t, codeberg.Title, "Codeberg")
		be.Equal(t, codeberg.Added, time.Time{})

		bookmarksMenu, ok := root.Items[1].(*Folder)
		be.True(t, ok)
		be.Equal(t, bookmarksMenu.Title, "Bookmarks Menu")
		be.Equal(t, bookmarksMenu.IsReadingList, false)
		be.Equal(t, len(bookmarksMenu.Items), 0)

		readingList, ok := root.Items[2].(*Folder)
		be.True(t, ok)
		be.Equal(t, readingList.Title, "Reading List")
		be.Equal(t, readingList.IsReadingList, true)
		be.Equal(t, len(readingList.Items), 1)

		bouncepaw, ok := readingList.Items[0].(Bookmark)
		be.True(t, ok)
		be.Equal(t, bouncepaw.URL, "https://bouncepaw.com")
		be.Equal(t, bouncepaw.Title, "Bouncepaw")
	})
}

func TestWrite(t *testing.T) {
	t.Parallel()

	t.Run("Simple", func(t *testing.T) {
		t.Parallel()

		root := &Folder{
			Title:    "My Bookmarks",
			Added:    time.Unix(1000, 0),
			Modified: time.Unix(2000, 0),
			Items: []Item{
				&Folder{
					Title:    "Work",
					Added:    time.Unix(1100, 0),
					Modified: time.Unix(1200, 0),
					Items: []Item{
						Bookmark{
							URL:      "https://example.com/?a=1&b=2",
							Title:    "Example <Site>",
							Tags:     []string{"work", "example"},
							Added:    time.Unix(1300, 0),
							Modified: time.Unix(1400, 0),
						},
					},
				},
				Bookmark{
					URL:      "https://bare.example/",
					Title:    "No Tags",
					Added:    time.Unix(1500, 0),
					Modified: time.Unix(1600, 0),
				},
			},
		}

		var buf bytes.Buffer
		err := root.Write(&buf)
		be.Equal(t, err, nil)

		out := buf.String()
		be.True(t, strings.Contains(out, "<!DOCTYPE NETSCAPE-Bookmark-file-1>"))
		be.True(t, strings.Contains(out, "<H1>My Bookmarks</H1>"))
		be.True(t, strings.Contains(out, ">Work</H3>"))
		be.True(t, strings.Contains(out, `HREF="https://example.com/?a=1&amp;b=2"`))
		be.True(t, strings.Contains(out, ">Example &lt;Site&gt;</A>"))
		be.True(t, strings.Contains(out, `TAGS="work,example"`))
		be.True(t, strings.Contains(out, `HREF="https://bare.example/"`))
		be.Equal(t, strings.Count(out, "TAGS="), 1)
	})
}

func TestRoundtrip(t *testing.T) {
	t.Parallel()
	original := &Folder{
		Title:    "Root",
		Added:    time.Unix(100, 0).UTC(),
		Modified: time.Unix(200, 0).UTC(),
		Items: []Item{
			&Folder{
				Title:    "Sub",
				Added:    time.Unix(300, 0).UTC(),
				Modified: time.Unix(400, 0).UTC(),
				Items: []Item{
					Bookmark{
						URL:      "https://go.dev/",
						Title:    "Go",
						Tags:     []string{"lang", "tools"},
						Added:    time.Unix(500, 0).UTC(),
						Modified: time.Unix(600, 0).UTC(),
					},
					Bookmark{
						URL:      "https://example.org/",
						Title:    "Example",
						Added:    time.Unix(700, 0).UTC(),
						Modified: time.Unix(800, 0).UTC(),
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	be.Equal(t, original.Write(&buf), nil)

	got, err := Read(&buf)
	be.Equal(t, err, nil)

	be.Equal(t, got.Title, original.Title)
	be.Equal(t, len(got.Items), 1)

	sub, ok := got.Items[0].(*Folder)
	be.True(t, ok)
	be.Equal(t, sub.Title, "Sub")
	be.Equal(t, sub.Added, original.Items[0].(*Folder).Added)
	be.Equal(t, sub.Modified, original.Items[0].(*Folder).Modified)
	be.Equal(t, len(sub.Items), 2)

	bm0, ok := sub.Items[0].(Bookmark)
	be.True(t, ok)
	be.Equal(t, bm0.URL, "https://go.dev/")
	be.Equal(t, bm0.Title, "Go")
	be.Equal(t, bm0.Tags, []string{"lang", "tools"})
	be.Equal(t, bm0.Added, time.Unix(500, 0).UTC())
	be.Equal(t, bm0.Modified, time.Unix(600, 0).UTC())

	bm1, ok := sub.Items[1].(Bookmark)
	be.True(t, ok)
	be.Equal(t, bm1.URL, "https://example.org/")
	be.Equal(t, bm1.Title, "Example")
	be.Equal(t, bm1.Tags, []string(nil))
}

func TestProbe(t *testing.T) {
	t.Parallel()

	t.Run("firefox1", func(t *testing.T) {
		t.Parallel()
		f, err := os.Open("testdata/firefox1.html")
		be.Equal(t, err, nil)
		defer f.Close()

		ok, err := Probe(f)
		be.Equal(t, err, nil)
		be.Equal(t, ok, true)

		root, err := Read(f)
		be.Equal(t, err, nil)
		be.Equal(t, root.Title, "Меню закладок")
	})

	t.Run("safari1", func(t *testing.T) {
		t.Parallel()
		f, err := os.Open("testdata/safari1.html")
		be.Equal(t, err, nil)
		defer f.Close()

		ok, err := Probe(f)
		be.Equal(t, err, nil)
		be.Equal(t, ok, true)

		root, err := Read(f)
		be.Equal(t, err, nil)
		be.Equal(t, root.Title, "Bookmarks")
	})

	t.Run("not netscape", func(t *testing.T) {
		t.Parallel()
		r := strings.NewReader("<!DOCTYPE html><html></html>")
		ok, err := Probe(r)
		be.Equal(t, err, nil)
		be.Equal(t, ok, false)
	})
}
