package readpage

import (
	"context"
	"golang.org/x/net/html"
	"io"
	"log"
	"net/url"
)

func listenForTitle(ctx context.Context, nodes chan *html.Node, data *FoundData) {
	state := 0
	// 0 looking
	// 1 found
	for node := range nodes {
		log.Printf("Recv %v\n", node)
		if state == 0 {
			if node.Type == html.ElementNode && node.Data == "title" {
				data.title = node.FirstChild.Data
				state = 1
			}
		}
	}
	println("read all nodes")
}

func listenForBookmarkOf(ctx context.Context, nodes chan *html.Node, data *FoundData) {
	for {
		select {
		case n := <-nodes:
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
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func listenForPostName(ctx context.Context, nodes chan *html.Node, data *FoundData) {
	state := 0
	// 0 nothing found yet
	// 1 found a p-name
	// When 1, look for a text node. After finding it, return.
	for {
		select {
		case n := <-nodes:
			switch {
			case state == 0 && nodeHasClass(n, "p-name"):
				state = 1
			case state == 1 && n.Type == html.TextNode:
				data.PostName = n.Data
			}
		case <-ctx.Done():
			return
		}
	}
}

func listenForTags(ctx context.Context, nodes chan *html.Node, data *FoundData) {
	for {
		select {
		case <-ctx.Done():
			return
		case n := <-nodes:
			if n.Type == html.ElementNode && nodeHasClass(n, "p-category") {
				tag := n.FirstChild.Data
				data.Tags = append(data.Tags, tag)
			}
			// Not returning, there might be more.
		}
	}
}

func listenForMycomarkup(ctx context.Context, nodes chan *html.Node, data *FoundData) {
	for {
		select {
		case <-ctx.Done():
			return
		case n := <-nodes:
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
				return
			}
		}
	}
}

func listenForHFeed(ctx context.Context, nodes chan *html.Node, data *FoundData) {
	for {
		select {
		case <-ctx.Done():
			return
		case n := <-nodes:
			if nodeHasClass(n, "h-feed") {
				data.IsHFeed = true
				return
			}

			// If we've found an h-entry, then it's highly-highly unlikely that the
			// document is an h-feed. At least in Betula.
			if nodeHasClass(n, "h-entry") {
				return
			}
		}
	}
}
