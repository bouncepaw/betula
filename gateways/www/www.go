// SPDX-FileCopyrightText: 2024 arne
// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2026 Danila Gorelko
//
// SPDX-License-Identifier: AGPL-3.0-only

package wwwgw

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"

	wwwports "git.sr.ht/~bouncepaw/betula/ports/www"
	"git.sr.ht/~bouncepaw/betula/settings"
)

type WWW struct {
	readLimit int64
}

var _ wwwports.WorldWideWeb = &WWW{}

func New() *WWW {
	return &WWW{readLimit: 20000}
}

func NewWithLimit(limit int64) *WWW {
	return &WWW{readLimit: limit}
}

var client = http.Client{
	Timeout: 2 * time.Second,
}

func (www *WWW) TitleOfPage(addr string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, addr, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", settings.UserAgent())
	resp, err := client.Do(req)
	if err != nil {
		if err.(*url.Error).Timeout() {
			return "", wwwports.ErrTimeout
		}
		return "", err
	}
	defer func(Body io.ReadCloser) {
		if err := Body.Close(); err != nil {
			log.Printf("Can't close the body of the response from %s\n", addr)
		}
	}(resp.Body)

	r, err := charset.NewReader(resp.Body, resp.Header.Get("Content-Type"))
	if err != nil {
		r = resp.Body
	}

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
