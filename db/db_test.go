package db

import (
	"git.sr.ht/~bouncepaw/betula/types"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func InitInMemoryDB() {
	Initialize(":memory:")
	const q = `
insert into Posts
   (URL, Title, Description, Visibility, CreationTime, DeletionTime)
values
	(
	 'https://bouncepaw.com',
	 'Bouncepaw website',
	 'A cute website by Bouncepaw',
	 0, '2023-03-17', null
	),
   (
    'https://mycorrhiza.wiki',
    'Mycorrhiza Wiki',
    'A wiki engine',
    1, '2023-03-17', null
   ),
	(
	 'http://lesarbr.es',
	 'Les Arbres',
	 'Legacy mirror of [[1]]',
	 1, '2023-03-17', '2023-03-18'
	)
`
	mustExec(q)
}

func TestLinkCount(t *testing.T) {
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
		Categories: []types.Category{
			types.Category{Name: "cat"},
			types.Category{Name: "dog"},
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
