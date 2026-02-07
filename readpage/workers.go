// SPDX-FileCopyrightText: 2023 Danila Gorelko
// SPDX-FileCopyrightText: 2023 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package readpage

import (
	"golang.org/x/net/html"
)

const (
	stateLooking = iota
	stateFound
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
