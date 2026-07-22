// SPDX-FileCopyrightText: 2024 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2026 Danila Gorelko
// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package db

import (
	"context"
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
	ctx := context.Background()
	repo := NewTagsRepo()

	authedTags, err := repo.Tags(ctx, true)
	be.Err(t, err, nil)
	be.Equal(t, len(authedTags), 2)

	publicTags, err := repo.Tags(ctx, false)
	be.Err(t, err, nil)
	be.Equal(t, len(publicTags), 1)
}

func TestDescriptions(t *testing.T) {
	initInMemoryTags()
	ctx := context.Background()
	repo := NewTagsRepo()

	desc := "Octopi have 8 legs."
	be.Err(t, repo.SetTagDescription(ctx, "octopus", desc), nil)

	got, err := repo.DescriptionForTag(ctx, "octopus")
	be.Err(t, err, nil)
	be.Equal(t, got, desc)

	got, err = repo.DescriptionForTag(ctx, "flounder")
	be.Err(t, err, nil)
	be.Equal(t, got, "")
}

func TestDeleteTagDescription(t *testing.T) {
	initInMemoryTags()
	ctx := context.Background()
	repo := NewTagsRepo()

	desc := "Octopi have 8 legs."
	be.Err(t, repo.SetTagDescription(ctx, "octopus", desc), nil)
	be.Err(t, repo.SetTagDescription(ctx, "octopus", ""), nil)

	got, err := repo.DescriptionForTag(ctx, "octopus")
	be.Err(t, err, nil)
	be.Equal(t, got, "")
}

func TestDeleteTag(t *testing.T) {
	initInMemoryTags()
	ctx := context.Background()
	repo := NewTagsRepo()

	desc := "Flounder has no legs."
	be.Err(t, repo.SetTagDescription(ctx, "flounder", desc), nil)
	be.Err(t, repo.DeleteTag(ctx, "flounder"), nil)

	exists, err := repo.TagExists(ctx, "flounder")
	be.Err(t, err, nil)
	be.True(t, !exists)

	got, err := repo.DescriptionForTag(ctx, "flounder")
	be.Err(t, err, nil)
	be.Equal(t, got, "")
}

func TestTagExists(t *testing.T) {
	initInMemoryTags()
	ctx := context.Background()
	repo := NewTagsRepo()

	exists, err := repo.TagExists(ctx, "flounder")
	be.Err(t, err, nil)
	be.True(t, exists)

	exists, err = repo.TagExists(ctx, "orca")
	be.Err(t, err, nil)
	be.True(t, !exists)
}

func TestRenameTag(t *testing.T) {
	initInMemoryTags()
	ctx := context.Background()
	repo := NewTagsRepo()

	be.Err(t, repo.RenameTag(ctx, "flounder", "orca"), nil)
	cats, err := repo.Tags(ctx, true)
	be.Err(t, err, nil)
	be.Equal(t, len(cats), 2)

	be.Err(t, repo.RenameTag(ctx, "orca", "octopus"), nil)
	cats, err = repo.Tags(ctx, true)
	be.Err(t, err, nil)
	be.Equal(t, len(cats), 1)
}

// tests SetTagsFor and TagsForBookmarkByID.
func TestBookmarkTags(t *testing.T) {
	initInMemoryTags()
	ctx := context.Background()
	repo := NewTagsRepo()

	tags := []types.Tag{
		{Name: "salmon"},
		{Name: "carp"},
	}
	be.Err(t, repo.SetTagsFor(ctx, 2, tags), nil)

	tags, err := repo.TagsForBookmarkByID(ctx, 2)
	be.Err(t, err, nil)
	be.Equal(t, len(tags), 2)
}

func TestTagCount(t *testing.T) {
	initInMemoryTags()
	ctx := context.Background()
	repo := NewTagsRepo()

	count, err := repo.TagCount(ctx, true)
	be.Err(t, err, nil)
	be.Equal(t, count, uint(2))

	count, err = repo.TagCount(ctx, false)
	be.Err(t, err, nil)
	be.Equal(t, count, uint(1))
}
