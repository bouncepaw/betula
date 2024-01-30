package db

import (
	"git.sr.ht/~bouncepaw/betula/types"
	"testing"
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
			types.Tag{Name: "cat"},
			types.Tag{Name: "dog"},
		},
		URL:         "https://betula.mycorrhiza.wiki",
		Title:       "Betula",
		Description: "",
		Visibility:  types.Public,
	}
	AddPost(post)
	if BookmarkCount(true) != 3 {
		t.Errorf("Faulty AddPost")
	}
}
