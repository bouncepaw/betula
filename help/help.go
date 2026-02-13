// SPDX-FileCopyrightText: 2023 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2025 Danila Gorelko
// SPDX-FileCopyrightText: 2025 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

// Package help manages the built-in documentation.
package help

import (
	"embed"
	"html/template"
	"log"

	"git.sr.ht/~bouncepaw/betula/pkg/myco"
)

type Topic struct {
	Name         string
	SidebarTitle string
	Rendered     template.HTML
}

var (
	//go:embed en/*
	english embed.FS
	Topics  = []Topic{
		{"index", "Betula introduction", ""},
		{"meta", "Metainformation", ""},
		{"mycomarkup", "Mycomarkup formatting", ""},
		{"search", "Advanced search", ""},
		{"errors", "Error codes", ""},
		{"miniflux", "Miniflux integration", ""},
	}
)

func init() {
	for i, topic := range Topics {
		raw, err := english.ReadFile("en/" + topic.Name + ".myco")
		if err != nil {
			log.Fatalln(err)
		}

		topic.Rendered = myco.MarkupToHTML(string(raw))
		Topics[i] = topic
	}
}

// GetEnglishHelp returns English-language help for the given topic.
func GetEnglishHelp(topicName string) (topic Topic, found bool) {
	for _, topic := range Topics {
		if topic.Name == topicName {
			return topic, true
		}
	}
	return
}
