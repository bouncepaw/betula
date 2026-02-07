// SPDX-FileCopyrightText: 2023 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2024 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2024 arne
// SPDX-FileCopyrightText: 2025 Danila Gorelko
// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

// Package readpage fetches information from a web page.
package readpage

import (
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"

	"git.sr.ht/~bouncepaw/betula/settings"
)

// SPDX-SnippetBegin
// SPDX-SnippetCopyrightText: Bourne Co. Music Publishers
/*
	When you wish upon a star
	Makes no difference who you are
	Anything your heart desires
	Will come to you
	— Leigh Harline, Ned Washington
*/
// SPDX-SnippetEnd

var (
	ErrNoTitleFound = errors.New("no title found in the document")
	ErrTimeout      = errors.New("request timed out")

	titleWorkers = []worker{listenForTitle}
)

// FindTitle finds a <title> in the document.
//
// If there is no title, or the title is empty string, ErrNoTitleFound is returned.
// If any other error occurred, it is returned.
//
// Deprecated: Use the service wrapper.
// TODO: delete this code, reimplement in the service.
func FindTitle(link string) (string, error) {
	data, err := findDataByLink(link, titleWorkers)
	if data.title == "" && err == nil {
		err = ErrNoTitleFound
	}
	return data.title, err
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
}

func findData(link string, workers []worker, doc *html.Node) (data FoundData, err error) {
	addr, err := url.ParseRequestURI(link)
	if err != nil {
		panic("Invalid URL passed.")
	}

	data.docurl = addr

	// The workers have 1 second to fulfill their fate. That's a lot of time!
	// I'm being generous here.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
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
	req, err := http.NewRequest(http.MethodGet, link, nil)
	if err != nil {
		log.Printf("Failed to construct request from ‘%s’\n", link)
		return data, err
	}

	req.Header.Set("User-Agent", settings.UserAgent())
	resp, err := client.Do(req)
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

	r, err := charset.NewReader(resp.Body, resp.Header.Get("Content-Type"))
	if err != nil {
		r = resp.Body
	}

	doc, err := html.Parse(r)
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
