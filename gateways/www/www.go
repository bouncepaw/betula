// SPDX-FileCopyrightText: 2024 arne
// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2026 Danila Gorelko
//
// SPDX-License-Identifier: AGPL-3.0-only

package wwwgw

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"

	wwwports "git.sr.ht/~bouncepaw/betula/ports/www"
)

type WWW struct {
	readLimit   int64
	userAgentFn func() string
}

var _ wwwports.WorldWideWeb = &WWW{}

func New(userAgentFn func() string) *WWW {
	return &WWW{
		readLimit:   20000,
		userAgentFn: userAgentFn,
	}
}

func NewWithLimit(userAgentFn func() string, limit int64) *WWW {
	return &WWW{
		readLimit:   limit,
		userAgentFn: userAgentFn,
	}
}

var client = http.Client{
	Timeout: 2 * time.Second,
}

func (www *WWW) fetch(addr string) (r io.Reader, closeBody func(), err error) {
	req, err := http.NewRequest(http.MethodGet, addr, nil)
	if err != nil {
		return nil, nil, err
	}

	req.Header.Set("User-Agent", www.userAgentFn())
	resp, err := client.Do(req) //nolint:bodyclose // resp.Body is closed with closeBody func.
	if err != nil {
		if err.(*url.Error).Timeout() {
			return nil, nil, wwwports.ErrTimeout
		}
		return nil, nil, err
	}

	r, err = charset.NewReader(resp.Body, resp.Header.Get("Content-Type"))
	if err != nil {
		r = resp.Body
	}

	return r, func() { resp.Body.Close() }, nil
}

func (www *WWW) TitleOfPage(addr string) (string, error) {
	r, closeBody, err := www.fetch(addr)
	if err != nil {
		return "", err
	}
	defer closeBody()

	var buffer bytes.Buffer
	doc, err := html.Parse(
		io.LimitReader(
			io.TeeReader(r, &buffer),
			www.readLimit,
		),
	)
	if err != nil {
		return "", err
	}

	title := www.findTitle(doc)
	if title != "" {
		return title, nil
	}

	doc, err = html.Parse(
		io.MultiReader(&buffer, r),
	)
	if err != nil {
		return "", err
	}

	title = www.findTitle(doc)
	if title == "" {
		return "", wwwports.ErrNoTitleFound
	}
	return title, nil
}

func (www *WWW) RelAlternates(addr string) ([]wwwports.RelAlternate, error) {
	r, closeBody, err := www.fetch(addr)
	if err != nil {
		return nil, err
	}
	defer closeBody()

	var buffer bytes.Buffer
	doc, err := html.Parse(
		io.LimitReader(
			io.TeeReader(r, &buffer),
			www.readLimit,
		),
	)
	if err != nil {
		return nil, err
	}

	alternates := []wwwports.RelAlternate{}
	www.findRelAlternates(doc, &alternates)
	if len(alternates) > 0 {
		return alternates, nil
	}

	doc, err = html.Parse(io.MultiReader(&buffer, r))
	if err != nil {
		return nil, err
	}
	www.findRelAlternates(doc, &alternates)
	return alternates, nil
}

func (www *WWW) findRelAlternates(n *html.Node, alternates *[]wwwports.RelAlternate) {
	if n.Type == html.ElementNode && n.Data == "link" {
		var rel, typ, href, title string
		for _, attr := range n.Attr {
			switch strings.ToLower(attr.Key) {
			case "rel":
				rel = attr.Val
			case "type":
				typ = attr.Val
			case "href":
				href = attr.Val
			case "title":
				title = attr.Val
			}
		}
		if relContainsAlternate(rel) {
			*alternates = append(*alternates, wwwports.RelAlternate{
				Type:  typ,
				Href:  href,
				Title: title,
			})
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		www.findRelAlternates(c, alternates)
	}
}

func relContainsAlternate(rel string) bool {
	for token := range strings.FieldsSeq(rel) {
		if strings.EqualFold(token, "alternate") {
			return true
		}
	}
	return false
}

func (www *WWW) findTitle(n *html.Node) string {
	if n.Type == html.ElementNode && n.Data == "title" {
		if n.FirstChild != nil {
			return strings.TrimSpace(n.FirstChild.Data)
		}
	}
	if n.Type == html.ElementNode && n.Data == "body" {
		return ""
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if title := www.findTitle(c); title != "" {
			return title
		}
	}
	return ""
}
