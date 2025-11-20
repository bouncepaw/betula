// SPDX-FileCopyrightText: 2022-2025 Betula contributors
//
// SPDX-License-Identifier: AGPL-3.0-only

package readpage

import (
	"embed"
	"golang.org/x/net/html"
	"io"
	"reflect"
	"strings"
	"testing"
)

//go:embed testdata/*
var testdata embed.FS

func TestTrickyURL(t *testing.T) {
	f, err := testdata.Open("testdata/h-entry with substituted url.html")
	if err != nil {
		panic(err)
	}
	docb, err := io.ReadAll(f)
	if err != nil {
		panic(err)
	}
	doc := string(docb)
	tricks := []string{
		`https://willcrichton.net/notes/portable-epubs#epub-content%2FEPUB%2Findex.xhtml$`,
		`https://garden.bouncepaw.com/#Fresh_свежак_freŝa`,
	}
	for _, trick := range tricks {
		txt := strings.ReplaceAll(doc, "BETULA", trick)
		ht, err := html.Parse(strings.NewReader(txt))
		if err != nil {
			panic(err)
		}

		data, err := findData("https://bouncepaw.com", []worker{listenForBookmarkOf}, ht)
		if err != nil {
			panic(err)
		}
		if data.BookmarkOf != trick {
			t.Errorf("Got %q want %q", data.BookmarkOf, trick)
		}
	}
}

func TestTitles(t *testing.T) {
	table := map[string]string{
		"title outside head": "A title!",
		"title none":         "",
	}
	for name, expectedTitle := range table {
		file, _ := testdata.Open("testdata/" + name + ".html")
		doc, _ := html.Parse(file)
		data, err := findData("https://bouncepaw.com", titleWorkers, doc)
		if data.title != expectedTitle {
			t.Errorf("In ‘%s’, expected title ‘%s’, got ‘%s’. Error value is ‘%v’.",
				name, expectedTitle, data.title, err)
		}
	}
}

func TestHEntries(t *testing.T) {
	gutenberg := "https://www.gutenberg.org/files/2701/2701-h/2701-h.htm#link2HCH0001"
	mushatlas := "https://mushroomcoloratlas.com/"

	table := map[string]FoundData{
		"h-entry with p-name": {
			PostName:   "CHAPTER 1. Loomings.",
			BookmarkOf: "",
			Tags:       nil,
			Mycomarkup: "",
			IsHFeed:    false,
		},

		"h-entry with p-name u-bookmark-of": {
			PostName:   "CHAPTER 1. Loomings.",
			BookmarkOf: gutenberg,
			Tags:       nil,
			Mycomarkup: "",
			IsHFeed:    false,
		},

		"h-feed with h-entries": {
			PostName:   "CHAPTER 1. Loomings.",
			BookmarkOf: "",
			Tags:       nil,
			Mycomarkup: "",
			IsHFeed:    true,
		},

		"mycomarkup linked": {
			PostName:   "Mushroom color atlas",
			BookmarkOf: mushatlas,
			Tags:       []string{"myco"},
			Mycomarkup: "Many cool colors",
			IsHFeed:    false,
		},
	}

	for name, expectedData := range table {
		file, _ := testdata.Open("testdata/" + name + ".html")
		doc, _ := html.Parse(file)
		data, err := findData("https://bouncepaw.com", makeRepostWorkers, doc)
		data.docurl = nil
		if !reflect.DeepEqual(data, expectedData) {
			t.Errorf("In ‘%s’,\nwant %v,\ngot %v. Error value is ‘%v’.",
				name, expectedData, data, err)
		}
	}
}
