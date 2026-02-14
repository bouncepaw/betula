// SPDX-FileCopyrightText: 2023 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2024 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2025 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package searchingsvc

import (
	"regexp"

	"git.sr.ht/~bouncepaw/betula/db"
	searchingports "git.sr.ht/~bouncepaw/betula/ports/searching"
	"git.sr.ht/~bouncepaw/betula/types"
)

var (
	// TODO: Exclude more characters
	excludeTagRe = regexp.MustCompile(`-#([^?!:#@<>*|'"&%{}\\\s]+)\s*`)
	includeTagRe = regexp.MustCompile(`#([^?!:#@<>*|'"&%{}\\\s]+)\s*`)
	// TODO: argument will be added in a future version
	includeRepostRe = regexp.MustCompile(`\brepost:()\s*`)
)

type Service struct{}

var _ searchingports.Service = &Service{}

func New() *Service {
	return &Service{}
}

func (svc *Service) ForFederated(query string, offset, limit uint) (bookmarks []types.Bookmark, totalBookmarks uint) {
	if limit > types.BookmarksPerPage {
		limit = types.BookmarksPerPage
	}

	query, excludedTags := svc.extractWithRegex(query, excludeTagRe)
	query, includedTags := svc.extractWithRegex(query, includeTagRe)

	return db.SearchOffset(query, includedTags, excludedTags, offset, limit)
}

func (svc *Service) For(query string, authorized bool, page uint) (bookmarksInPage []types.Bookmark, totalBookmarks uint) {
	query, excludedTags := svc.extractWithRegex(query, excludeTagRe)
	query, includedTags := svc.extractWithRegex(query, includeTagRe)
	query, includedRepostMarkers := svc.extractWithRegex(query, includeRepostRe)

	return db.Search(query, includedTags, excludedTags, len(includedRepostMarkers) != 0, authorized, page)
}

func (svc *Service) extractWithRegex(query string, regex *regexp.Regexp) (string, []string) {
	var extracted []string
	for _, result := range regex.FindAllStringSubmatch(query, -1) {
		result := result
		extracted = append(extracted, result[1])
	}
	query = regex.ReplaceAllString(query, "")
	return query, extracted
}
