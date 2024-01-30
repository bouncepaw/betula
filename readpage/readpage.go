// Package readpage fetches information from a web page.
package readpage

import (
	"context"
	"errors"
	"git.sr.ht/~bouncepaw/betula/fediverse/activities"
	"golang.org/x/net/html"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
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
	ErrTimeout      = errors.New("request timed out")

	titleWorkers      = []worker{listenForTitle}
	makeRepostWorkers = []worker{
		listenForPostName, listenForBookmarkOf, listenForTags, listenForMycomarkup, listenForHFeed,
	}
	checkRepostWorkers = []worker{listenForRepostOf}
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

// FindDataForMyRepost finds data relevant for us to make a repost.
func FindDataForMyRepost(link string) (FoundData, error) {
	return findDataByLink(link, makeRepostWorkers)
}

func IsThisValidRepost(report activities.AnnounceReport) (validRepost bool, err error) {
	data, err := findDataByLink(report.RepostPage, checkRepostWorkers)
	valid := data.RepostOf != "" && data.RepostOf == report.OriginalPage
	return valid, err
}

// The rest of the package is private.

type worker func(chan *html.Node, *FoundData)

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

	// BookmarkOf is the first u-bookmark-of found. Must be a valid URL.
	BookmarkOf string

	// RepostOf is the first u-repost-of found. Must be a valid URL.
	RepostOf string

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
		panic("Invalid URL passed.")
	}

	data.docurl = addr

	// The workers have 1 second to fulfill their fate. That's a lot of time!
	// I'm being generous here.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var (
		wg            sync.WaitGroup
		nodeReceivers = make([]chan *html.Node, len(workers))
		nodes         = make(chan *html.Node)
	)

	// Enter workers. They make your wishes come true!
	for i, w := range workers {
		nodeReceivers[i] = make(chan *html.Node)
		wg.Add(1)

		go func(i int, w worker) {
			w(nodeReceivers[i], &data)
			wg.Done()
		}(i, w)
	}

	// Enter traverser. It goes through all the elements and yields them.
	wg.Add(1)
	go func() {
		traverse(ctx, doc, nodes)
		close(nodes)
		wg.Done()
	}()

	// Fan out nodes.
	for node := range nodes {
		for _, nodeReceiver := range nodeReceivers {
			// They will listen.
			nodeReceiver <- node
		}
	}
	for _, nodeReceiver := range nodeReceivers {
		close(nodeReceiver)
	}

	wg.Wait()

	return
}

// findDataByLink finds the data you wished for in the document, considering the timeouts.
func findDataByLink(link string, workers []worker) (data FoundData, err error) {
	resp, err := client.Get(link)
	if err != nil {
		if err.(*url.Error).Timeout() {
			log.Printf("Request to %s timed out\n", link)
			return data, ErrTimeout
		}

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
		return data, err
	}

	return findData(link, workers, doc)
}

// Depth-first traversal.
func traverse(ctx context.Context, n *html.Node, nodes chan *html.Node) {
	select {
	case <-ctx.Done():
		return
	default:
	}
	// We don't care about other types of nodes. So let's just drop them!
	if n.Type == html.ElementNode || n.Type == html.TextNode {
		nodes <- n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		traverse(ctx, c, nodes)
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
