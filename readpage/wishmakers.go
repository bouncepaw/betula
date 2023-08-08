package readpage

import (
	"context"
	"golang.org/x/net/html"
	"net/url"
)

func listenForTitle(ctx context.Context, incoming chan *html.Node, data *Data) {
	for {
		select {
		case node := <-incoming:
			if node.Type == html.ElementNode && node.Data == "title" {
				data.Title = node.FirstChild.Data
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func listenForBookmarkOf(ctx context.Context, incoming chan *html.Node, data *Data) {
	for {
		select {
		case n := <-incoming:
			if n.Type == html.ElementNode && nodeHasClass(n, "u-bookmark-of") {
				href, found := attrValue(n, "href")
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

func listenForPostName(ctx context.Context, incoming chan *html.Node, data *Data) {
	state := 0
	// 0 nothing found yet
	// 1 found a p-name
	// When 1, look for a text node. After finding it, return.
	for {
		select {
		case n := <-incoming:
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
