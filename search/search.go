// SPDX-FileCopyrightText: 2023 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2024 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2025 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package search

import (
	"git.sr.ht/~bouncepaw/betula/db"
	"git.sr.ht/~bouncepaw/betula/types"
	"regexp"
)

var (
	// TODO: Exclude more characters
	excludeTagRe = regexp.MustCompile(`-#([^?!:#@<>*|'"&%{}\\\s]+)\s*`)
	includeTagRe = regexp.MustCompile(`#([^?!:#@<>*|'"&%{}\\\s]+)\s*`)

	// TODO: argument will be added in a future version
	includeRepostRe = regexp.MustCompile(`\brepost:()\s*`)
)

func ForFederated(query string, offset, limit uint) (bookmarks []types.Bookmark, totalResults uint) {
	if limit > types.BookmarksPerPage {
		limit = types.BookmarksPerPage
	}

	query, excludedTags := extractWithRegex(query, excludeTagRe)
	query, includedTags := extractWithRegex(query, includeTagRe)

	return db.SearchOffset(query, includedTags, excludedTags, offset, limit)
}

// For searches for the given query.
func For(query string, authorized bool, page uint) (postsInPage []types.Bookmark, totalPosts uint) {
	// We extract excluded tags first.
	query, excludedTags := extractWithRegex(query, excludeTagRe)
	query, includedTags := extractWithRegex(query, includeTagRe)
	query, includedRepostMarkers := extractWithRegex(query, includeRepostRe)

	return db.Search(query, includedTags, excludedTags, len(includedRepostMarkers) != 0, authorized, page)
}

func extractWithRegex(query string, regex *regexp.Regexp) (string, []string) {
	var extracted []string
	for _, result := range regex.FindAllStringSubmatch(query, -1) {
		result := result
		extracted = append(extracted, result[1])
	}
	query = regex.ReplaceAllString(query, "")
	return query, extracted
}
