// SPDX-FileCopyrightText: 2026 Danila Gorelko
//
// SPDX-License-Identifier: AGPL-3.0-only

package wwwgw

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nalgeon/be"
	"golang.org/x/net/html"

	wwwports "git.sr.ht/~bouncepaw/betula/ports/www"
)

func testUserAgent() string { return "Betula-test" }

func TestFindTitle(t *testing.T) {
	t.Parallel()

	findTitle := func(t *testing.T, input, expected string) {
		t.Helper()
		doc, err := html.Parse(strings.NewReader(input))
		be.Err(t, err, nil)

		result := New(testUserAgent).findTitle(doc)
		be.Equal(t, result, expected)
	}

	t.Run("Title in head", func(t *testing.T) {
		t.Parallel()
		findTitle(t, `<html><head><title>Test Title</title></head><body></body></html>`, "Test Title")
	})

	t.Run("Title in root", func(t *testing.T) {
		t.Parallel()
		findTitle(t, `<html><title>Root Title</title><head></head><body></body></html>`, "Root Title")
	})

	t.Run("Title with whitespace", func(t *testing.T) {
		t.Parallel()
		findTitle(t, `<html><head><title>  Title with spaces  </title></head><body></body></html>`, "Title with spaces")
	})

	t.Run("No title element", func(t *testing.T) {
		t.Parallel()
		findTitle(t, `<html><head></head><body></body></html>`, "")
	})

	t.Run("Empty title", func(t *testing.T) {
		t.Parallel()
		findTitle(t, `<html><head><title></title></head><body></body></html>`, "")
	})

	t.Run("Title after body should not be found", func(t *testing.T) {
		t.Parallel()
		findTitle(t, `<html><head></head><body><title>Body Title</title></body></html>`, "")
	})

	t.Run("Complex html structure", func(t *testing.T) {
		t.Parallel()
		findTitle(t, `<!DOCTYPE html><html lang="en"><head><meta charset="UTF-8"><title>Complex Page</title></head><body><div>Content</div></body></html>`, "Complex Page")
	})
}

func TestRelAlternates(t *testing.T) {
	t.Parallel()

	relAlternates := func(t *testing.T, input string, expected []wwwports.RelAlternate) {
		t.Helper()
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write([]byte(input))
		}))
		defer server.Close()

		result, err := New(testUserAgent).RelAlternates(server.URL)
		be.Err(t, err, nil)
		be.Equal(t, result, expected)
	}

	t.Run("Single alternate", func(t *testing.T) {
		t.Parallel()
		relAlternates(t,
			`<html><head><link rel="alternate" type="application/rss+xml" href="/feed.xml" title="RSS"></head><body></body></html>`,
			[]wwwports.RelAlternate{{Type: "application/rss+xml", Href: "/feed.xml", Title: "RSS"}})
	})

	t.Run("No title attribute", func(t *testing.T) {
		t.Parallel()
		relAlternates(t,
			`<html><head><link rel="alternate" type="application/atom+xml" href="/atom.xml"></head></html>`,
			[]wwwports.RelAlternate{{Type: "application/atom+xml", Href: "/atom.xml", Title: ""}})
	})

	t.Run("Multiple rel tokens", func(t *testing.T) {
		t.Parallel()
		relAlternates(t,
			`<html><head><link rel="alternate stylesheet" type="text/css" href="/alt.css"></head></html>`,
			[]wwwports.RelAlternate{{Type: "text/css", Href: "/alt.css", Title: ""}})
	})

	t.Run("Case-insensitive attributes and rel", func(t *testing.T) {
		t.Parallel()
		relAlternates(t,
			`<html><head><LINK REL="ALTERNATE" TYPE="application/json" HREF="/feed.json"></head></html>`,
			[]wwwports.RelAlternate{{Type: "application/json", Href: "/feed.json", Title: ""}})
	})

	t.Run("ActivityPub links as rendered by GET /", func(t *testing.T) {
		t.Parallel()
		relAlternates(t,
			`<html><head><link rel="alternate" type='application/ld+json; profile="https://www.w3.org/ns/activitystreams"' href="/@user"><link rel="alternate" type='application/activity+json' href="/@user"></head></html>`,
			[]wwwports.RelAlternate{
				{Type: `application/ld+json; profile="https://www.w3.org/ns/activitystreams"`, Href: "/@user", Title: ""},
				{Type: "application/activity+json", Href: "/@user", Title: ""},
			})
	})

	t.Run("Non-alternate ignored", func(t *testing.T) {
		t.Parallel()
		relAlternates(t,
			`<html><head><link rel="stylesheet" href="/style.css"><link rel="icon" href="/favicon.ico"></head></html>`,
			[]wwwports.RelAlternate{})
	})

	t.Run("No links", func(t *testing.T) {
		t.Parallel()
		relAlternates(t,
			`<html><head></head><body></body></html>`,
			[]wwwports.RelAlternate{})
	})
}

func TestTitleOfPageWithSmallLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(`<html><head><meta charset="utf-8"><title>Late Title</title></head><body></body></html>`))
	}))
	defer server.Close()
	title, err := NewWithLimit(testUserAgent, 10).TitleOfPage(server.URL)
	be.Err(t, err, nil)
	be.Equal(t, title, "Late Title")
}

func TestTitleOfPageWithDefaultLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(`<html><head><meta charset="utf-8"><title>Late Title</title></head><body></body></html>`))
	}))
	defer server.Close()
	title, err := New(testUserAgent).TitleOfPage(server.URL)
	be.Err(t, err, nil)
	be.Equal(t, title, "Late Title")
}
