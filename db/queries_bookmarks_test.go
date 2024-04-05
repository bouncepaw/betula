package db

import (
	"testing"

	"git.sr.ht/~bouncepaw/betula/types"
)

func TestBookmarkCount(t *testing.T) {
	InitInMemoryDB()
	resAuthed := BookmarkCount(true)
	if resAuthed != 2 {
		t.Errorf("Wrong authorized LinkCount, got %d", resAuthed)
	}
	resAnon := BookmarkCount(false)
	if resAnon != 1 {
		t.Errorf("Wrong unauthorized LinkCount, got %d", resAnon)
	}
}

func TestAddPost(t *testing.T) {
	InitInMemoryDB()
	post := types.Bookmark{
		CreationTime: "2023-03-18",
		Tags: []types.Tag{
			{Name: "cat"},
			{Name: "dog"},
		},
		URL:         "https://betula.mycorrhiza.wiki",
		Title:       "Betula",
		Description: "",
		Visibility:  types.Public,
	}
	InsertBookmark(post)
	if BookmarkCount(true) != 3 {
		t.Errorf("Faulty AddPost")
	}
}

func TestRandomBookmarks(t *testing.T) {
	InitInMemoryDB()
	MoreTestingBookmarks()

	cases := []struct {
		authorized bool
		n          uint
	}{
		{true, 20},
		{false, 20},
	}

	for _, tc := range cases {
		bookmarks, total := RandomBookmarks(tc.authorized, tc.n)
		if len(bookmarks) != int(total) {
			t.Errorf("Length of bookmarks does not match the total count")
		}
		creationTime := bookmarks[0].CreationTime
		for _, bookmark := range bookmarks[1:] {
			if bookmark.CreationTime > creationTime {
				t.Errorf("Bookmarks not in correct order")
			}
			creationTime = bookmark.CreationTime
		}
	}
}
