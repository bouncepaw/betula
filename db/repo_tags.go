// SPDX-FileCopyrightText: 2024 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package db

import (
	"context"
	"database/sql"
	"errors"

	taggingports "git.sr.ht/~bouncepaw/betula/ports/tagging"
	"git.sr.ht/~bouncepaw/betula/types"
)

type TagsRepo struct {
}

var _ taggingports.Repository = (*TagsRepo)(nil)

func NewTagsRepo() *TagsRepo {
	return &TagsRepo{}
}

func (repo *TagsRepo) SetTagDescription(ctx context.Context, tagName, description string) error {
	if description == "" {
		_, err := db.ExecContext(ctx, `delete from TagDescriptions where TagName = ?`, tagName)
		return err
	}
	_, err := db.ExecContext(ctx, `
replace into TagDescriptions (TagName, Description)
values (?, ?);
`, tagName, description)
	return err
}

func (repo *TagsRepo) DeleteTag(ctx context.Context, tagName string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `delete from TagDescriptions where TagName = ?`, tagName); err != nil {
		return errors.Join(err, tx.Rollback())
	}
	if _, err := tx.ExecContext(ctx, `delete from TagsToPosts where TagName = ?`, tagName); err != nil {
		return errors.Join(err, tx.Rollback())
	}
	return tx.Commit()
}

func (repo *TagsRepo) DescriptionForTag(ctx context.Context, tagName string) (string, error) {
	var description string
	err := db.QueryRowContext(ctx, `select Description from TagDescriptions where TagName = ?`, tagName).Scan(&description)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return description, err
}

func (repo *TagsRepo) TagCount(ctx context.Context, authorized bool) (uint, error) {
	const q = `
select
	count(distinct TagName)
from
	TagsToPosts
inner join
	(select ID from Bookmarks where DeletionTime is null and (Visibility = 1 or ?))
as
	Filtered
on
	TagsToPosts.PostID = Filtered.ID
`
	var count uint
	err := db.QueryRowContext(ctx, q, authorized).Scan(&count)
	return count, err
}

func (repo *TagsRepo) Tags(ctx context.Context, authorized bool) ([]types.Tag, error) {
	const q = `
select
   TagName,
   count(PostID)
from
   TagsToPosts
inner join
    (select ID from Bookmarks where DeletionTime is null and (Visibility = 1 or ?))
as
	Filtered
on
    TagsToPosts.PostID = Filtered.ID
group by
	TagName;
`
	rows, err := db.QueryContext(ctx, q, authorized)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []types.Tag
	for rows.Next() {
		var tag types.Tag
		if err := rows.Scan(&tag.Name, &tag.BookmarkCount); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	return tags, rows.Err()
}

func (repo *TagsRepo) TagExists(ctx context.Context, tagName string) (bool, error) {
	var has bool
	err := db.QueryRowContext(ctx, `select exists(select 1 from TagsToPosts where TagName = ?);`, tagName).Scan(&has)
	return has, err
}

func (repo *TagsRepo) RenameTag(ctx context.Context, oldTagName, newTagName string) error {
	_, err := db.ExecContext(ctx, `
update TagsToPosts
set TagName = ?
where TagName = ?;
`, newTagName, oldTagName)
	return err
}

func (repo *TagsRepo) SetTagsFor(ctx context.Context, bookmarkID int, tags []types.Tag) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `delete from TagsToPosts where PostID = ?;`, bookmarkID); err != nil {
		return errors.Join(err, tx.Rollback())
	}

	for _, tag := range tags {
		if tag.Name == "" {
			continue
		}
		if _, err := tx.ExecContext(ctx, `insert into TagsToPosts (TagName, PostID) values (?, ?);`, tag.Name, bookmarkID); err != nil {
			return errors.Join(err, tx.Rollback())
		}
	}
	return tx.Commit()
}

// TagsForBookmarkByID returns the tags for the given bookmark ID.
//
// Deprecated: Use the local bookmark repo.
func (repo *TagsRepo) TagsForBookmarkByID(ctx context.Context, id int) ([]types.Tag, error) {
	return tagsForBookmarkByID(ctx, db, id)
}

// tagsForBookmarkByID returns the tags set on the given bookmark, ordered by
// name. It is the single in-package implementation that the various repos
// delegate to. It takes a querier so callers can run it standalone (with the
// global *sql.DB) or inside a caller's transaction.
func tagsForBookmarkByID(ctx context.Context, q querier, id int) ([]types.Tag, error) {
	rows, err := q.QueryContext(ctx, `
select distinct TagName
from TagsToPosts
where PostID = ?
order by TagName;
`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []types.Tag
	for rows.Next() {
		var tag types.Tag
		if err := rows.Scan(&tag.Name); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	return tags, rows.Err()
}
