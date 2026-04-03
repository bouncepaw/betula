// SPDX-FileCopyrightText: 2024 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2026 Danila Gorelko
//
// SPDX-License-Identifier: AGPL-3.0-only

package db

import (
	"testing"

	"git.sr.ht/~bouncepaw/betula/types"
	"github.com/nalgeon/be"
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

	be.Equal(t, len(Tags(true)), 2)
	be.Equal(t, len(Tags(false)), 1)
}

func TestDescriptions(t *testing.T) {
	initInMemoryTags()

	desc := "Octopi have 8 legs."
	SetTagDescription("octopus", desc)

	be.Equal(t, DescriptionForTag("octopus"), desc)
	be.Equal(t, DescriptionForTag("flounder"), "")
}

func TestDeleteTagDescription(t *testing.T) {
	initInMemoryTags()

	desc := "Octopi have 8 legs."
	SetTagDescription("octopus", desc)
	deleteTagDescription("octopus")

	be.Equal(t, DescriptionForTag("octopus"), "")
}

func TestDeleteTag(t *testing.T) {
	initInMemoryTags()

	desc := "Flounder has no legs."
	SetTagDescription("flounder", desc)
	DeleteTag("flounder")

	be.True(t, !TagExists("flounder"))
	be.Equal(t, DescriptionForTag("flounder"), "")
}

func TestTagExists(t *testing.T) {
	initInMemoryTags()

	be.True(t, TagExists("flounder"))
	be.True(t, !TagExists("orca"))
}

func TestRenameTag(t *testing.T) {
	initInMemoryTags()

	RenameTag("flounder", "orca")
	cats := Tags(true)
	be.Equal(t, len(cats), 2)

	RenameTag("orca", "octopus")
	cats = Tags(true)
	be.Equal(t, len(cats), 1)
}

// tests SetTagsFor and TagsForBookmarkByID.
func TestPostTags(t *testing.T) {
	initInMemoryTags()
	tags := []types.Tag{
		{Name: "salmon"},
		{Name: "carp"},
	}
	SetTagsFor(2, tags)

	tags = TagsForBookmarkByID(2)
	be.Equal(t, len(tags), 2)
}

func TestTagCount(t *testing.T) {
	initInMemoryTags()
	be.Equal(t, TagCount(true), 2)
	be.Equal(t, TagCount(false), 1)
}
