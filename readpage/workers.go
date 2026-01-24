// SPDX-FileCopyrightText: 2022-2025 Betula contributors
//
// SPDX-License-Identifier: AGPL-3.0-only

package readpage

import (
	"io"
	"log"
	"net/http"
	"strings"

	"git.sr.ht/~bouncepaw/betula/pkg/stricks"
	"git.sr.ht/~bouncepaw/betula/settings"
	"golang.org/x/net/html"
)

const (
	stateLooking = iota
	stateFound
	stateFoundText
)

const (
	stateNotSure = iota
	stateSure
)

func nodeIsNotEmptyText(node *html.Node) bool {
	if node.Type != html.TextNode {
		return false
	}
	for _, c := range node.Data {
		if !strings.ContainsRune(" \t\n\r", c) {
			return true
		}
	}
	return false
}

func listenForTitle(nodes chan *html.Node, data *FoundData) {
	state := stateLooking
	for n := range nodes {
		if state == stateLooking {
			if n.Type == html.ElementNode && n.Data == "title" {
				data.title = n.FirstChild.Data
				state = stateFound
			}
		}
	}
}

func listenForBookmarkOf(nodes chan *html.Node, data *FoundData) {
	state := stateLooking
	for n := range nodes {
		if state == stateFound {
			continue
		}

		if n.Type == html.ElementNode && nodeHasClass(n, "u-bookmark-of") {
			href, found := nodeAttribute(n, "href")
			if !found {
				// Huh? OK, a faulty document, stuff happens.
				return
			}

			if !stricks.ValidURL(href) {
				// Huh? Can't you produce a worthy document once in a while? OK.
				//
				// Maybe we could overcome it sometimes later. However, Betula
				// provides valid absolute URL:s here, so whatever. Other
				// implementations strive for better!
				state = stateFound
				continue
			}

			data.BookmarkOf = href
			state = stateFound
		}
	}
}

func listenForRepostOf(nodes chan *html.Node, data *FoundData) {
	state := stateLooking
	for n := range nodes {
		if state == stateFound {
			continue
		}

		if n.Type == html.ElementNode && nodeHasClass(n, "u-repost-of") {
			href, found := nodeAttribute(n, "href")
			if !found {
				// Huh? OK, a faulty document, stuff happens.
				return
			}

			if !stricks.ValidURL(href) {
				state = stateFound
				continue
			}

			data.RepostOf = href
			state = stateFound
		}
	}
}

func listenForPostName(nodes chan *html.Node, data *FoundData) {
	state := stateLooking
	for n := range nodes {
		switch {
		case state == stateFoundText:
			continue
		case state == stateFound && nodeIsNotEmptyText(n):
			data.PostName = n.Data
			state = stateFoundText
		case state == stateLooking && nodeHasClass(n, "p-name"):
			state = stateFound
		}
	}
}

func listenForTags(nodes chan *html.Node, data *FoundData) {
	state := stateLooking // or stateFound
	for n := range nodes {
		switch {
		case state == stateFound && nodeIsNotEmptyText(n):
			tagName := strings.TrimSpace(n.Data)
			data.Tags = append(data.Tags, tagName)
			state = stateLooking
		case state == stateLooking && nodeHasClass(n, "p-category"):
			state = stateFound
		}
	}
}

func listenForMycomarkup(nodes chan *html.Node, data *FoundData) {
	state := stateLooking
	for n := range nodes {
		if state == stateFound {
			continue
		}

		// Looking for <link rel="alternate" type="text/mycomarkup" href="...">
		if n.Type == html.ElementNode && n.Data == "link" {
			rel, foundRel := nodeAttribute(n, "rel")
			kind, foundKind := nodeAttribute(n, "type")
			href, foundHref := nodeAttribute(n, "href")

			if !foundRel || !foundKind || !foundHref ||
				rel != "alternate" || kind != "text/mycomarkup" {
				continue
			}

			addr, err := data.docurl.Parse(href)
			if err != nil {
				log.Printf("URL ‘%s’ is a bad URL.\n", href)
				// Link issue.
				continue
			}

			// We've found a valid <link> to a Mycomarkup document! Let's fetch it.

			req, err := http.NewRequest(http.MethodGet, addr.String(), nil)
			if err != nil {
				log.Printf("Failed to construct request from ‘%s’\n", addr.String())
				continue
			}

			req.Header.Set("User-Agent", settings.UserAgent())
			resp, err := client.Do(req)
			if err != nil {
				log.Printf("Failed to fetch Mycomarkup document from ‘%s’\n", addr.String())
				continue
			}

			raw, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Printf("Failed to read Mycomarkup document from ‘%s’\n", addr.String())
				resp.Body.Close()
				continue
			}

			data.Mycomarkup = string(raw)
			state = 1
			resp.Body.Close()
		}
	}
}

func listenForHFeed(nodes chan *html.Node, data *FoundData) {
	state := stateNotSure
	for n := range nodes {
		if state == stateSure {
			continue
		}

		if nodeHasClass(n, "h-feed") {
			data.IsHFeed = true
			state = stateSure
			continue
		}

		// If we've found an h-entry, then it's highly-highly unlikely that the
		// document is an h-feed. At least in Betula.
		if nodeHasClass(n, "h-entry") {
			state = stateSure
		}
	}
}
