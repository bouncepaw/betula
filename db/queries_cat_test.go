package db

import (
	"git.sr.ht/~bouncepaw/betula/types"
	"testing"
)

func initInMemoryCategories() {
	InitInMemoryDB()
	q := `
insert into CategoriesToPosts (CatName, PostID) values
('octopus', 1),
('flounder', 2),
('flounder', 3);
`
	mustExec(q)
}

func TestCategories(t *testing.T) {
	initInMemoryCategories()
	catsWithRights := Categories(true)
	if len(catsWithRights) != 2 {
		t.Errorf("Wrong authorized categories count")
	}
	catsWithoutRights := Categories(false)
	if len(catsWithoutRights) != 1 {
		t.Errorf("Wrong unauthorized categories count")
	}
}

func TestDescriptions(t *testing.T) {
	initInMemoryCategories()

	desc := "Octopi have 8 legs."
	SetCategoryDescription("octopus", desc)
	if DescriptionForCategory("octopus") != desc {
		t.Errorf("Octopus has wrong description: %s", DescriptionForCategory("octopus"))
	}

	if DescriptionForCategory("flounder") != "" {
		t.Errorf("Flound has a description: %s", DescriptionForCategory("flounder"))
	}
}

func TestCategoryExists(t *testing.T) {
	initInMemoryCategories()

	if !CategoryExists("flounder") {
		t.Errorf("Flounder does not exist")
	}
	if CategoryExists("orca") {
		t.Errorf("Orca exists")
	}
}

func TestRenameCategory(t *testing.T) {
	initInMemoryCategories()

	RenameCategory("flounder", "orca")
	cats := Categories(true)
	if len(cats) != 2 {
		t.Errorf("Faulty renaming from Flounder to Orca")
	}

	RenameCategory("orca", "octopus")
	cats = Categories(true)
	if len(cats) != 1 {
		t.Errorf("Faulty merging orca into octopus")
	}
}

// tests SetCategoriesFor and CategoriesForPost
func TestPostCategories(t *testing.T) {
	initInMemoryCategories()
	cats := []types.Category{
		{Name: "salmon"},
		{Name: "carp"},
	}
	SetCategoriesFor(2, cats)

	cats = CategoriesForPost(2)
	if len(cats) != 2 {
		t.Errorf("Faulty category saving")
	}
}
