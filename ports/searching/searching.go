// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package searchingports

import "git.sr.ht/~bouncepaw/betula/types"

type Service interface {
	// For searches for the given query. authorized controls visibility;
	// page is 1-based.
	For(query string, authorized bool, page uint) (bookmarksInPage []types.Bookmark, totalBookmarks uint)

	// ForFederated runs a federated search with offset/limit pagination.
	ForFederated(query string, offset, limit uint) (bookmarks []types.Bookmark, totalBookmarks uint)
}
