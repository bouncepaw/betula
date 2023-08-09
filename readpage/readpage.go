// Package readpage fetches information from a web page.
package readpage

import (
	"context"
	"errors"
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

var (
	ErrNoTitleFound = errors.New("no title found in the document")

	titleWorkers  = []worker{listenForTitle}
	repostWorkers = []worker{
		listenForPostName, listenForBookmarkOf, listenForTags, listenForMycomarkup, listenForHFeed,
	}
)

// FindTitle finds a <title> in the document.
//
// If there is no title, or the title is empty string, ErrNoTitleFound is returned.
// If any other error occurred, it is returned.
func FindTitle(link string) (string, error) {
	data, err := findDataByLink(link, titleWorkers)
	if data.title == "" && err == nil {
		err = ErrNoTitleFound
	}
	return data.title, err
}

// FindRepostData finds data relevant to reposts in the document.
func FindRepostData(link string) (FoundData, error) {
	return findDataByLink(link, repostWorkers)
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

func findData(link string, workers []worker, doc *html.Node) (data FoundData, err error) {
	addr, err := url.ParseRequestURI(link)
	if err != nil {
		log.Fatalln("Invalid URL passed.")
	}

	data.docurl = addr

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

// findDataByLink finds the data you wished for in the document, considering the timeouts.
func findDataByLink(link string, workers []worker) (data FoundData, err error) {
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

	doc, err := html.Parse(resp.Body)
	if err != nil {
		log.Printf("Can't parse HTML from %s\n", link)
	}

	return findData(link, workers, doc)
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
