// SPDX-FileCopyrightText: 2023 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2024 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2025 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2026 Danila Gorelko
// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package searchingsvc

import (
	"context"
	"log/slog"
	"regexp"

	searchingports "git.sr.ht/~bouncepaw/betula/ports/searching"
	"git.sr.ht/~bouncepaw/betula/types"
)

var (
	// TODO: Exclude more characters
	excludeTagRe = regexp.MustCompile(`-#([^?!:#@<>*|'"&%{}\\\s]+)\s*`)
	includeTagRe = regexp.MustCompile(`#([^?!:#@<>*|'"&%{}\\\s]+)\s*`)
	// TODO: argument will be added in a future version
	includeRemarkRe = regexp.MustCompile(`\bremark:()\s*`)
)

type Service struct {
	repo searchingports.Repository
}

var _ searchingports.Service = &Service{}

func New(repo searchingports.Repository) *Service {
	return &Service{repo: repo}
}

func (svc *Service) ForFederated(query string, offset, limit uint) (bookmarks []types.Bookmark, totalBookmarks uint) {
	if limit > types.BookmarksPerPage {
		limit = types.BookmarksPerPage
	}

	query, excludedTags := svc.extractWithRegex(query, excludeTagRe)
	query, includedTags := svc.extractWithRegex(query, includeTagRe)

	bookmarks, totalBookmarks, err := svc.repo.SearchOffset(context.Background(), searchingports.OffsetQuery{
		Text:         query,
		IncludedTags: includedTags,
		ExcludedTags: excludedTags,
		Offset:       offset,
		Limit:        limit,
	})
	if err != nil {
		slog.Error("Failed to run federated search", "query", query, "err", err)
		return nil, 0
	}
	return bookmarks, totalBookmarks
}

func (svc *Service) For(query string, authorized bool, page uint) (bookmarksInPage []types.Bookmark, totalBookmarks uint) {
	query, excludedTags := svc.extractWithRegex(query, excludeTagRe)
	query, includedTags := svc.extractWithRegex(query, includeTagRe)
	query, includedRemarkMarkers := svc.extractWithRegex(query, includeRemarkRe)

	bookmarksInPage, totalBookmarks, err := svc.repo.Search(context.Background(), searchingports.Query{
		Text:         query,
		IncludedTags: includedTags,
		ExcludedTags: excludedTags,
		RemarksOnly:  len(includedRemarkMarkers) != 0,
		Authorized:   authorized,
		Page:         page,
	})
	if err != nil {
		slog.Error("Failed to run search", "query", query, "err", err)
		return nil, 0
	}
	return bookmarksInPage, totalBookmarks
}

func (svc *Service) extractWithRegex(query string, regex *regexp.Regexp) (string, []string) {
	var extracted []string
	for _, result := range regex.FindAllStringSubmatch(query, -1) {
		extracted = append(extracted, result[1])
	}
	query = regex.ReplaceAllString(query, "")
	return query, extracted
}
