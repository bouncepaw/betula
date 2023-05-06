package db

import (
	"git.sr.ht/~bouncepaw/betula/types"
	"testing"
)

func TestPostCount(t *testing.T) {
	InitInMemoryDB()
	resAuthed := PostCount(true)
	if resAuthed != 2 {
		t.Errorf("Wrong authorized LinkCount, got %d", resAuthed)
	}
	resAnon := PostCount(false)
	if resAnon != 1 {
		t.Errorf("Wrong unauthorized LinkCount, got %d", resAnon)
	}
}

func TestAddPost(t *testing.T) {
	InitInMemoryDB()
	post := types.Post{
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
	if PostCount(true) != 3 {
		t.Errorf("Faulty AddPost")
	}
}
