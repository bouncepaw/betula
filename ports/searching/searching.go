// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package searchingports

import (
	"context"

	"git.sr.ht/~bouncepaw/betula/types"
)

type (
	// Query describes a page-based search over local bookmarks.
	Query struct {
		// Text is the free-text part of the query.
		Text string
		// IncludedTags and ExcludedTags filter results by tag.
		IncludedTags []string
		ExcludedTags []string
		// RepostsOnly keeps only reposts when set.
		RepostsOnly bool
		// Authorized controls visibility: private bookmarks are included only
		// when set.
		Authorized bool
		// Page is 1-based.
		Page uint
	}

	// OffsetQuery describes an offset/limit search over public bookmarks, used
	// for federated search.
	OffsetQuery struct {
		// Text is the free-text part of the query.
		Text string
		// IncludedTags and ExcludedTags filter results by tag.
		IncludedTags []string
		ExcludedTags []string
		// Offset and Limit paginate the results.
		Offset uint
		Limit  uint
	}
)

type Repository interface {
	// Search runs a page-based search over local bookmarks.
	Search(ctx context.Context, query Query) (bookmarksInPage []types.Bookmark, totalBookmarks uint, err error)
	// SearchOffset runs an offset/limit search over public bookmarks.
	SearchOffset(ctx context.Context, query OffsetQuery) (bookmarks []types.Bookmark, totalBookmarks uint, err error)
}

type Service interface {
	// For searches for the given query. authorized controls visibility;
	// page is 1-based.
	For(query string, authorized bool, page uint) (bookmarksInPage []types.Bookmark, totalBookmarks uint)

	// ForFederated runs a federated search with offset/limit pagination.
	ForFederated(query string, offset, limit uint) (bookmarks []types.Bookmark, totalBookmarks uint)
}
