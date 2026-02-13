// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package feedsports

import "git.sr.ht/~bouncepaw/betula/pkg/rss"

type Service interface {
	DigestFeed() (*rss.Feed, error)
	BookmarksFeed() (*rss.Feed, error)
}
