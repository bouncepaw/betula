// SPDX-FileCopyrightText: 2023 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2024 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2026 Danila Gorelko
//
// SPDX-License-Identifier: AGPL-3.0-only

package feedssvc

import (
	"testing"
	"time"

	"git.sr.ht/~bouncepaw/betula/db"
	"github.com/nalgeon/be"
)

func TestFiveLastDays(t *testing.T) {
	db.InitInMemoryDB()
	db.MoreTestingBookmarks()
	days, dayStamps, dayPosts := fiveLastDays(
		time.Date(2023, 3, 21, 0, 0, 0, 0, time.UTC))

	_ = days

	correctDayStamps := []string{"2023-03-20", "2023-03-19", "2023-03-18", "2023-03-17", "2023-03-16"}
	for i, stamp := range dayStamps {
		be.Equal(t, correctDayStamps[i], stamp)
	}

	correctBookmarkCounts := []int{2, 1, 0, 1, 0}
	for i, posts := range dayPosts {
		be.Equal(t, correctBookmarkCounts[i], len(posts))
	}
}
