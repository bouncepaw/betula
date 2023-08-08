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

const (
	WishTitle = 1 << iota
	WishPostName
	WishBookmarkOf
	WishTags
	WishMycomarkup
)

type wishmakerFunc func(context.Context, chan *html.Node, *Data)

var wishesToWishmakers = map[int]wishmakerFunc{
	WishTitle:      listenForTitle,
	WishPostName:   listenForPostName,
	WishBookmarkOf: listenForBookmarkOf,
}

// Data is all data returned by GetData. Specific fields are set iff you wish for them.
type Data struct {
	Title      string
	PostName   string
	BookmarkOf *url.URL
	Tags       []string
	Mycomarkup string
}

// GetTitle is a shorthand for wishing for page title only.
func GetTitle(link string) (string, error) {
	data, err := GetData(link, WishTitle)
	return data.Title, err
}

// GetData finds the data you wished for in the linked document, considering the timeouts.
func GetData(link string, wishes int) (data Data, err error) {
	// TODO: Set some timeout
	resp, err := http.Get(link)
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
		log.Printf("Can't parse html from %s\n", link)
	}

	// The wishmakers have 2 seconds to fulfill their fate. That's a lot of time!
	// I'm being generous here. The showstopper will tell the wishmakers when
	// the time is up.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Enter traverser. It goes through all the elements and yields them.
	nodes := make(chan *html.Node)
	go traverse(doc, nodes)

	// Enter wishmakers. They make your wishes come true!
	var (
		completeWishes = make(chan int)
		wishmakers     []wishmakerFunc
		nodeReceivers  []chan *html.Node
		i              = 0
	)
	for wish, wishmaker := range wishesToWishmakers { // For all possible wishes
		if wishes&wish == 0 {
			// If this wish is not wished for, we don't care.
			continue
		}
		wishmakers = append(wishmakers, wishmaker)
		nodeReceivers = append(nodeReceivers, make(chan *html.Node))
		i++

		go func(i int, wishmaker wishmakerFunc) {
			wishmaker(ctx, nodeReceivers[i], &data)
			close(nodeReceivers[i])
			// Wishmaker tells fanouter it's done.
			completeWishes <- i
		}(i, wishmaker)
	}

	// Enter fanouter. It broadcasts all found nodes to the wishmakers who are
	// still online.
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case i := <-completeWishes:
				// Wishmaker told it's done. OK, don't take it into account later.
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

func traverse(n *html.Node, outcoming chan *html.Node) {
	outcoming <- n
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		traverse(c, outcoming)
	}
}

func attrValue(node *html.Node, attrName string) (value string, found bool) {
	for _, attr := range node.Attr {
		if attr.Key == attrName {
			return attr.Val, true
		}
	}
	return "", false
}

func nodeHasClass(node *html.Node, class string) bool {
	classList, found := attrValue(node, "class")
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
