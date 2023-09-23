package readpage

import (
	"golang.org/x/net/html"
	"io"
	"log"
	"net/url"
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

			uri, err := url.ParseRequestURI(href)
			if err != nil {
				// Huh? Can't you produce a worthy document once in a while? OK.
				//
				// Maybe we could overcome it sometimes later. However, Betula
				// provides valid absolute URL:s here, so whatever. Other
				// implementations strive for better!
				return
			}

			data.BookmarkOf = uri
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

			uri, err := url.ParseRequestURI(href)
			if err != nil {
				// Huh? Can't you produce a worthy document once in a while? OK.
				//
				// Maybe we could overcome it sometimes later. However, Betula
				// provides valid absolute URL:s here, so whatever. Other
				// implementations strive for better!
				return
			}

			data.RepostOf = uri
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
		case state == stateFound && n.Type == html.TextNode:
			data.PostName = n.Data
			state = stateFoundText
		case state == stateLooking && nodeHasClass(n, "p-name"):
			state = stateFound
		}
	}
}

func listenForTags(nodes chan *html.Node, data *FoundData) {
	for n := range nodes {
		if n.Type == html.ElementNode && nodeHasClass(n, "p-category") {
			tag := n.FirstChild.Data
			data.Tags = append(data.Tags, tag)
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

			resp, err := client.Get(addr.String())
			if err != nil {
				log.Printf("Failed to fetch Mycomarkup document from ‘%s’\n", addr.String())
			}

			raw, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Printf("Failed to read Mycomarkup document from ‘%s’\n", addr.String())
			}

			data.Mycomarkup = string(raw)
			state = 1
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
