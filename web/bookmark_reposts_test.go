package web

import (
	"net/http/httptest"
	"testing"
	"time"

	"git.sr.ht/~bouncepaw/betula/db"
	"git.sr.ht/~bouncepaw/betula/types"
)

func TestRenderBookmarkIncludesReposts(t *testing.T) {
	db.InitInMemoryDB()

	bm := types.Bookmark{
		URL:        "https://example.com",
		Title:      "Example",
		Visibility: types.Public,
	}
	id := db.InsertBookmark(bm)
	bookmark, found := db.GetBookmarkByID(int(id))
	if !found {
		t.Fatalf("bookmark not found after insert")
	}

	db.SaveRepost(bookmark.ID, types.RepostInfo{URL: "https://links.alice/1", Name: "Alice", Timestamp: time.Now()})
	db.SaveRepost(bookmark.ID, types.RepostInfo{URL: "https://links.bob/2", Name: "Bob", Timestamp: time.Now()})

	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	data := renderBookmark(bookmark, w, r)
	if len(data.Reposts) != 2 {
		t.Errorf("expected 2 reposts in render data, got %d", len(data.Reposts))
	}
}
