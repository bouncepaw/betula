// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package remotebookmarksports

import (
	"context"

	"git.sr.ht/~bouncepaw/betula/types"
)

type (
	RemoteBookmarkRepository interface {
		Exists(id string) (bool, error)
		GetActorIDFor(bookmarkID string) (string, error)
		Delete(ctx context.Context, bookmarkID string) error

		// TODO: Add error to the signatures of the following methods.

		GetRemoteBookmarksBy(authorID string, page uint) (bookmarks []types.RemoteBookmark, total uint)
		GetRemoteBookmarks(page uint) (bookmarks []types.RemoteBookmark, total uint)
		GetRemoteBookmarkByID(bookmarkID string) (types.RemoteBookmark, bool)
		InsertRemoteBookmark(b types.RemoteBookmark)
		UpdateRemoteBookmark(b types.RemoteBookmark)
	}

	Service interface {
		Render(context.Context, []types.RemoteBookmark) ([]types.RenderedRemoteBookmark, error)
		GetRemoteBookmarkByID(ctx context.Context, id string) (types.RemoteBookmark, error)
	}
)
