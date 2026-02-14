// SPDX-FileCopyrightText: 2023 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2025 Danila Gorelko
// SPDX-FileCopyrightText: 2025 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package helpingsvc

import (
	"embed"
	"log"

	"git.sr.ht/~bouncepaw/betula/pkg/myco"
	helpingports "git.sr.ht/~bouncepaw/betula/ports/helping"
)

//go:embed docs/en/*
var english embed.FS

type Service struct {
	topics []helpingports.Topic
}

var _ helpingports.Service = &Service{}

func New() *Service {
	topicDefs := []struct {
		name         string
		sidebarTitle string
	}{
		{"index", "Betula introduction"},
		{"meta", "Metainformation"},
		{"mycomarkup", "Mycomarkup formatting"},
		{"search", "Advanced search"},
		{"errors", "Error codes"},
		{"miniflux", "Miniflux integration"},
	}

	topics := make([]helpingports.Topic, len(topicDefs))
	for i, def := range topicDefs {
		raw, err := english.ReadFile("docs/en/" + def.name + ".myco")
		if err != nil {
			log.Fatalln(err)
		}
		topics[i] = helpingports.Topic{
			Name:         def.name,
			SidebarTitle: def.sidebarTitle,
			Rendered:     myco.MarkupToHTML(string(raw)),
		}
	}
	return &Service{topics: topics}
}

func (svc *Service) AllTopics() []helpingports.Topic {
	return svc.topics
}

func (svc *Service) GetEnglishHelp(topicName string) (helpingports.Topic, bool) {
	for _, topic := range svc.topics {
		if topic.Name == topicName {
			return topic, true
		}
	}
	return helpingports.Topic{}, false
}
