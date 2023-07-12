// Package help manages the built-in documentation.
package help

import (
	"embed"
	"git.sr.ht/~bouncepaw/betula/myco"
	"html/template"
	"log"
)

type Topic struct {
	Id             string
	NameForSidebar string
	HTML           template.HTML
}

var (
	//go:embed en/*
	english       embed.FS
	englishLookup = map[string]Topic{
		"index":      {"index", "Betula introduction", ""},
		"mycomarkup": {"mycomarkup", "Mycomarkup formatting", ""},
		"search":     {"search", "Advanced search", ""},
	}
)

func init() {
	for name, topic := range englishLookup {
		raw, err := english.ReadFile("en/" + name + ".myco")
		if err != nil {
			log.Fatalln(err)
		}

		topic.HTML = myco.MarkupToHTML(string(raw))
		englishLookup[name] = topic
	}
}

// GetEnglishHelp returns English-language help for the given topic.
func GetEnglishHelp(topicName string) (topic Topic, found bool) {
	// Oh hey, we precalculate everything! How cool is that! Isn't that a O(1)!?
	topic, found = englishLookup[topicName]
	return
}
