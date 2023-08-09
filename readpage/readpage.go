// Package readpage fetches information from a web page.
package readpage

import (
	"context"
	"golang.org/x/net/html"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

/*
	When you wish upon a star
	Makes no difference who you are
	Anything your heart desires
	Will come to you
	— Leigh Harline, Ned Washington
*/

// FindTitle finds a <title> in the document.
func FindTitle(link string) (string, error) {
	data, err := findData(link, []worker{listenForTitle})
	return data.title, err
}

// FindRepostData finds data relevant to reposts in the document.
func FindRepostData(link string) (FoundData, error) {
	return findData(link, []worker{
		listenForPostName, listenForBookmarkOf, listenForTags, listenForMycomarkup, listenForHFeed,
	})
}

// The rest of the package is private.

type worker func(context.Context, chan *html.Node, *FoundData)

var client = http.Client{
	Timeout: 2 * time.Second,
}

// FoundData is all data found in a document. Specific fields are set iff you wish for them.
type FoundData struct {
	// title is the first <title> found.
	title string

	// docurl is the URL of the very document we're working with now. Needed to resolve relative links.
	docurl *url.URL

	// PostName is the first p-name found.
	PostName string

	// BookmarkOf is the first u-bookmark-of found.
	BookmarkOf *url.URL

	// Tags are all p-category found.
	Tags []string

	// Mycomarkup is the Mycomarkup text. It is fetched with a second request.
	Mycomarkup string

	// IsHFeed is true if the document has an h-feed somewhere in the beginning. You don't repost h-feed:s.
	IsHFeed bool
}

// findData finds the data you wished for in the linked document, considering the timeouts.
func findData(link string, workers []worker) (data FoundData, err error) {
	// TODO: Handle timeout
	resp, err := client.Get(link)
	if err != nil {
		log.Printf("Can't get response from %s\n", link)
		return data, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("Can't close the body of the response from %s\n", link)
		}
	}(resp.Body)

	addr, err := url.ParseRequestURI(link)
	if err != nil {
		log.Fatalln("Invalid URL passed.")
	}

	data.docurl = addr

	doc, err := html.Parse(resp.Body)
	if err != nil {
		log.Printf("Can't parse HTML from %s\n", link)
	}

	// The workers have 1 second to fulfill their fate. That's a lot of time!
	// I'm being generous here.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Enter traverser. It goes through all the elements and yields them.
	nodes := make(chan *html.Node)
	go traverse(doc, nodes)

	// Enter workers. They make your wishes come true!
	var (
		workDone      = make(chan int)
		nodeReceivers = make([]chan *html.Node, len(workers))
	)
	for i, wishmaker := range workers {
		nodeReceivers[i] = make(chan *html.Node)

		go func(i int, wishmaker worker) {
			wishmaker(ctx, nodeReceivers[i], &data)
			// Worker reports it's done.
			// — It's done.
			// — It's done.
			workDone <- i
			close(nodeReceivers[i])
		}(i, wishmaker)
	}

	// Fan out nodes to workers.
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case i := <-workDone:
				// A worker reported it's done. OK, taking notes.
				nodeReceivers[i] = nil
			case node := <-nodes:
				for _, nodeReceiver := range nodeReceivers {
					if nodeReceiver == nil {
						continue
					}
					// They will listen.
					nodeReceiver <- node
				}
			}
		}
	}()

	return
}

// Depth-first traversal.
func traverse(n *html.Node, outcoming chan *html.Node) {
	if n.Type == html.ElementNode || n.Type == html.TextNode {
		outcoming <- n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		traverse(c, outcoming)
	}
}

func nodeAttribute(node *html.Node, attrName string) (value string, found bool) {
	for _, attr := range node.Attr {
		if attr.Key == attrName {
			return attr.Val, true
		}
	}
	return "", false
}

func nodeHasClass(node *html.Node, class string) bool {
	classList, found := nodeAttribute(node, "class")
	if !found {
		return false
	}

	for _, c := range strings.Split(classList, " ") {
		if c == class {
			return true
		}
	}

	return false
}
