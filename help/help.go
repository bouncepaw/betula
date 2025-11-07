// Package help manages the built-in documentation.
package help

import (
	"embed"
	"git.sr.ht/~bouncepaw/betula/myco"
	"html/template"
	"log"
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
