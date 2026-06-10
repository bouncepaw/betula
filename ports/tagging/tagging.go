// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package taggingports

import (
	"context"

	"git.sr.ht/~bouncepaw/betula/types"
)

type Repository interface {
	// SetTagDescription sets the description for the tag. An empty description
	// removes it.
	SetTagDescription(ctx context.Context, tagName, description string) error
	// DeleteTag removes the tag along with its description.
	DeleteTag(ctx context.Context, tagName string) error
	// DescriptionForTag returns the tag's description, or an empty string if it
	// has none.
	DescriptionForTag(ctx context.Context, tagName string) (string, error)
	TagCount(ctx context.Context, authorized bool) (uint, error)
	Tags(ctx context.Context, authorized bool) ([]types.Tag, error)
	TagExists(ctx context.Context, tagName string) (bool, error)
	RenameTag(ctx context.Context, oldTagName, newTagName string) error
	SetTagsFor(ctx context.Context, bookmarkID int, tags []types.Tag) error
	// TagsForBookmarkByID returns the tags for the given bookmark ID.
	//
	// Deprecated: Use the local bookmark repo.
	TagsForBookmarkByID(ctx context.Context, id int) ([]types.Tag, error)
}
