// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package helpingports

import "html/template"

// Topic is a single help documentation topic.
type Topic struct {
	Name         string
	SidebarTitle string
	Rendered     template.HTML
}

type Service interface {
	// AllTopics returns all help topics (e.g. for sidebar).
	AllTopics() []Topic

	// GetEnglishHelp returns English-language help for the given topic name.
	GetEnglishHelp(topicName string) (topic Topic, found bool)
}
