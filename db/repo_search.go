// SPDX-FileCopyrightText: 2023 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2024 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2025 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2026 Danila Gorelko
// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package db

import (
	"context"
	"slices"
	"sort"
	"strings"

	searchingports "git.sr.ht/~bouncepaw/betula/ports/searching"
	"git.sr.ht/~bouncepaw/betula/types"
)

type SearchRepo struct {
}

var _ searchingports.Repository = (*SearchRepo)(nil)

func NewSearchRepo() *SearchRepo {
	return &SearchRepo{}
}

func (repo *SearchRepo) SearchOffset(ctx context.Context, query searchingports.OffsetQuery) (results []types.Bookmark, totalResults uint, err error) {
	text := strings.ToLower(query.Text)
	sort.Strings(query.IncludedTags)
	sort.Strings(query.ExcludedTags)

	rows, err := db.QueryContext(ctx, `
select ID, URL, Title, Description, Visibility, CreationTime, RepostOf, OriginalAuthorID
from Bookmarks
where DeletionTime is null and Visibility = 1
order by CreationTime desc
`)
	if err != nil {
		return nil, 0, err
	}

	unfilteredBookmarks, err := scanBookmarks(rows)
	if err != nil {
		return nil, 0, err
	}

	var i uint = 0
	var ignoredBookmarks uint = 0
	bookmarksToIgnore := query.Offset

	for _, post := range unfilteredBookmarks {
		if !textOK(post, text) {
			continue
		}

		post.Tags, err = tagsForBookmarkByID(ctx, db, post.ID)
		if err != nil {
			return nil, 0, err
		}
		if !tagsOK(post.Tags, query.IncludedTags, query.ExcludedTags) {
			continue
		}

		isRepost := post.RepostOf != nil
		if isRepost {
			continue // for now
		}

		totalResults++
		if ignoredBookmarks >= bookmarksToIgnore && i < query.Limit {
			results = append(results, post)
			i++
		} else {
			ignoredBookmarks++
		}
	}
	return results, totalResults, nil
}

func (repo *SearchRepo) Search(ctx context.Context, query searchingports.Query) (results []types.Bookmark, totalResults uint, err error) {
	text := strings.ToLower(query.Text)
	sort.Strings(query.IncludedTags)
	sort.Strings(query.ExcludedTags)

	rows, err := db.QueryContext(ctx, `
select ID, URL, Title, Description, Visibility, CreationTime, RepostOf, OriginalAuthorID
from Bookmarks
where DeletionTime is null and (Visibility = 1 or ?)
order by CreationTime desc
`, query.Authorized)
	if err != nil {
		return nil, 0, err
	}

	unfilteredBookmarks, err := scanBookmarks(rows)
	if err != nil {
		return nil, 0, err
	}

	var i uint = 0
	var ignoredBookmarks uint = 0
	bookmarksToIgnore := (query.Page - 1) * types.BookmarksPerPage

	// ‘Say, Bouncepaw, why did not you implement tag inclusion/exclusion
	//  part in SQL directly?’, some may ask.
	// ‘I did, and it was not worth it’, so I would respond.
	//
	// Addendum: I tried to make case-insensitive search in SQL too, and
	// failed loudly. Now all the search is done in Go. Per aspera ad
	// astra.
	//
	// We can't even parallelize it.
	for _, post := range unfilteredBookmarks {
		if !textOK(post, text) {
			continue
		}

		post.Tags, err = tagsForBookmarkByID(ctx, db, post.ID)
		if err != nil {
			return nil, 0, err
		}
		if !tagsOK(post.Tags, query.IncludedTags, query.ExcludedTags) {
			continue
		}

		isRepost := post.RepostOf != nil
		if !isRepost && query.RepostsOnly {
			continue
		}

		totalResults++
		if ignoredBookmarks >= bookmarksToIgnore && i < types.BookmarksPerPage {
			results = append(results, post)
			i++
		} else {
			ignoredBookmarks++
		}
	}
	return results, totalResults, nil
}

// true if keep, false if discard.
func textOK(post types.Bookmark, text string) bool {
	return strings.Contains(strings.ToLower(post.Title), text) ||
		strings.Contains(strings.ToLower(post.Description), text) ||
		strings.Contains(strings.ToLower(post.URL), text)
}

// true if keep, false if discard. All slices are sorted.
func tagsOK(postTags []types.Tag, includedTags, excludedTags []string) bool {
	J, K := len(includedTags), len(excludedTags)
	j, k := 0, 0
	includeMask := make([]bool, J)
	for _, postTag := range postTags {
		name := postTag.Name
		switch {
		case k < K && excludedTags[k] == name:
			return false
		case j < J && includedTags[j] == name:
			includeMask[j] = true
			j++
			continue
		}

		for j < J && includedTags[j] < name {
			j++
		}

		for k < K && excludedTags[k] < name {
			k++
		}
	}

	return !slices.Contains(includeMask, false)
}
