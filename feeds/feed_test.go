package feeds

import (
	"git.sr.ht/~bouncepaw/betula/db"
	"testing"
	"time"
)

func TestFiveLastDays(t *testing.T) {
	db.InitInMemoryDB()
	db.MoreTestingBookmarks()
	days, dayStamps, dayPosts := fiveLastDays(
		time.Date(2023, 3, 21, 0, 0, 0, 0, time.UTC))

	_ = days

	correctDayStamps := []string{"2023-03-20", "2023-03-19", "2023-03-18", "2023-03-17", "2023-03-16"}
	for i, stamp := range dayStamps {
		if correctDayStamps[i] != stamp {
			t.Errorf("Incorrect day stamp generated. Got %s, expected %s.", stamp, correctDayStamps[i])
		}
	}

	correctBookmarkCounts := []int{2, 1, 0, 1, 0}
	for i, posts := range dayPosts {
		if correctBookmarkCounts[i] != len(posts) {
			t.Errorf("Incorrect post count for %s. Got %d, expected %d. Data: %v.", dayStamps[i], len(posts), correctBookmarkCounts[i], posts)
		}
	}
}
