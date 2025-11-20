// SPDX-FileCopyrightText: 2022-2025 Betula contributors
//
// SPDX-License-Identifier: AGPL-3.0-only

package db

import (
	"git.sr.ht/~bouncepaw/betula/types"
	"testing"
)

func initInMemoryTags() {
	InitInMemoryDB()
	q := `
insert into TagsToPosts (TagName, PostID) values
('octopus', 1),
('flounder', 2),
('flounder', 3);
`
	mustExec(q)
}

func TestTags(t *testing.T) {
	initInMemoryTags()
	tagsWithRights := Tags(true)
	if len(tagsWithRights) != 2 {
		t.Errorf("Wrong authorized categories count")
	}
	tagsWithoutRights := Tags(false)
	if len(tagsWithoutRights) != 1 {
		t.Errorf("Wrong unauthorized categories count")
	}
}

func TestDescriptions(t *testing.T) {
	initInMemoryTags()

	desc := "Octopi have 8 legs."
	SetTagDescription("octopus", desc)
	if DescriptionForTag("octopus") != desc {
		t.Errorf("Octopus has wrong description: %s", DescriptionForTag("octopus"))
	}

	if DescriptionForTag("flounder") != "" {
		t.Errorf("Flound has a description: %s", DescriptionForTag("flounder"))
	}
}

func TestDeleteTagDescription(t *testing.T) {
	initInMemoryTags()

	desc := "Octopi have 8 legs."
	SetTagDescription("octopus", desc)
	deleteTagDescription("octopus")

	if DescriptionForTag("octopus") != "" {
		t.Errorf("Octopus has wrong description: %s", DescriptionForTag("octopus"))
	}
}

func TestDeleteTag(t *testing.T) {
	initInMemoryTags()

	desc := "Flounder has no legs."
	SetTagDescription("flounder", desc)
	DeleteTag("flounder")

	if TagExists("flounder") {
		t.Errorf("Faulty deletion flounder")
	}
	if DescriptionForTag("flounder") != "" {
		t.Errorf("Flounder has wrong description: %s", DescriptionForTag("flounder"))
	}
}

func TestTagExists(t *testing.T) {
	initInMemoryTags()

	if !TagExists("flounder") {
		t.Errorf("Flounder does not exist")
	}
	if TagExists("orca") {
		t.Errorf("Orca exists")
	}
}

func TestRenameTag(t *testing.T) {
	initInMemoryTags()

	RenameTag("flounder", "orca")
	cats := Tags(true)
	if len(cats) != 2 {
		t.Errorf("Faulty renaming from Flounder to Orca")
	}

	RenameTag("orca", "octopus")
	cats = Tags(true)
	if len(cats) != 1 {
		t.Errorf("Faulty merging orca into octopus")
	}
}

// tests SetTagsFor and TagsForBookmarkByID
func TestPostTags(t *testing.T) {
	initInMemoryTags()
	tags := []types.Tag{
		{Name: "salmon"},
		{Name: "carp"},
	}
	SetTagsFor(2, tags)

	tags = TagsForBookmarkByID(2)
	if len(tags) != 2 {
		t.Errorf("Faulty tag saving")
	}
}

func TestTagCount(t *testing.T) {
	initInMemoryTags()
	if TagCount(true) != 2 {
		t.Errorf("Wrong authorized categories count")
	}
	if TagCount(false) != 1 {
		t.Errorf("Wrong unauthorized categories count")
	}
}
