// SPDX-FileCopyrightText: 2023 Danila Gorelko
// SPDX-FileCopyrightText: 2023 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2024 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2025 Danila Gorelko
// SPDX-FileCopyrightText: 2025 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	likingports "git.sr.ht/~bouncepaw/betula/ports/liking"
	"git.sr.ht/~bouncepaw/betula/types"
)

type RepoLocalBookmarks struct{}

var _ likingports.LocalBookmarkRepository = &RepoLocalBookmarks{}

func NewLocalBookmarksRepo() *RepoLocalBookmarks {
	return &RepoLocalBookmarks{}
}

// querier is satisfied by both *sql.DB and *sql.Tx, so helpers can run their
// queries either standalone or inside a caller's transaction.
type querier interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

// scanBookmarks reads all bookmarks from rows and closes them.
func scanBookmarks(rows *sql.Rows) ([]types.Bookmark, error) {
	defer rows.Close()

	var bookmarks []types.Bookmark
	for rows.Next() {
		var bm types.Bookmark
		if err := rows.Scan(&bm.ID, &bm.URL, &bm.Title, &bm.Description, &bm.Visibility, &bm.CreationTime, &bm.RepostOf, &bm.OriginalAuthor); err != nil {
			return nil, err
		}
		bookmarks = append(bookmarks, bm)
	}
	return bookmarks, rows.Err()
}

func (repo *RepoLocalBookmarks) Exists(
	ctx context.Context,
	bookmarkID int,
) (bool, error) {
	row := db.QueryRowContext(
		ctx,
		`select exists(select 1 from Bookmarks where ID = ?)`,
		bookmarkID,
	)
	var exists bool
	err := row.Scan(&exists)
	return exists, err
}

func (repo *RepoLocalBookmarks) GetBookmarkByID(
	ctx context.Context,
	id int,
) (types.Bookmark, error) {
	row := db.QueryRowContext(ctx, `
		select ID, URL, Title, Description, Visibility, CreationTime, RepostOf, OriginalAuthorID 
		from Bookmarks
		where ID = ? and DeletionTime is null
	`, id)

	var b types.Bookmark
	err := row.Scan(&b.ID, &b.URL, &b.Title, &b.Description, &b.Visibility, &b.CreationTime, &b.RepostOf, &b.OriginalAuthor)
	return b, err
}

func (repo *RepoLocalBookmarks) InsertBookmark(
	ctx context.Context,
	bm types.Bookmark,
) (int64, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}

	var res sql.Result
	if bm.CreationTime == "" {
		res, err = tx.ExecContext(ctx, `
insert into Bookmarks (URL, Title, Description, Visibility, RepostOf, OriginalAuthorID)
values (?, ?, ?, ?, ?, ?);
`, bm.URL, bm.Title, bm.Description, bm.Visibility, bm.RepostOf, bm.OriginalAuthor)
	} else {
		res, err = tx.ExecContext(ctx, `
insert into Bookmarks (URL, Title, Description, Visibility, RepostOf, OriginalAuthorID, CreationTime)
values (?, ?, ?, ?, ?, ?, ?);
`, bm.URL, bm.Title, bm.Description, bm.Visibility, bm.RepostOf, bm.OriginalAuthor, bm.CreationTime)
	}
	if err != nil {
		return 0, errors.Join(err, tx.Rollback())
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, errors.Join(err, tx.Rollback())
	}

	for _, tag := range bm.Tags {
		if tag.Name == "" {
			continue
		}
		_, err = tx.ExecContext(ctx, `insert into TagsToPosts (TagName, PostID) values (?, ?);`, tag.Name, id)
		if err != nil {
			return 0, errors.Join(err, tx.Rollback())
		}
	}
	return id, tx.Commit()
}

func (repo *RepoLocalBookmarks) GetBookmarkIDByURL(
	ctx context.Context,
	url string,
) (int, error) {
	row := db.QueryRowContext(ctx, `
select ID from Bookmarks where URL = ? and DeletionTime is null limit 1;
`, url)
	var id int
	err := row.Scan(&id)
	return id, err
}

func (repo *RepoLocalBookmarks) Bookmarks(
	ctx context.Context,
	authorized bool,
	page uint,
) (bookmarks []types.Bookmark, total uint, err error) {
	if page == 0 {
		return nil, 0, fmt.Errorf("page 0 makes 0 sense")
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, 0, err
	}

	if err = tx.QueryRowContext(ctx, `
select count(ID) from Bookmarks where DeletionTime is null and (Visibility = 1 or ?);
`, authorized).Scan(&total); err != nil {
		return nil, 0, errors.Join(err, tx.Rollback())
	}

	rows, err := tx.QueryContext(ctx, `
select ID, URL, Title, Description, Visibility, CreationTime, RepostOf, OriginalAuthorID
from Bookmarks
where DeletionTime is null and (Visibility = 1 or ?)
order by CreationTime desc
limit ?
offset (? * (? - 1));
`, authorized, types.BookmarksPerPage, types.BookmarksPerPage, page)
	if err != nil {
		return nil, 0, errors.Join(err, tx.Rollback())
	}
	bookmarks, err = scanBookmarks(rows)
	if err != nil {
		return nil, 0, errors.Join(err, tx.Rollback())
	}

	bookmarks, err = repo.tagsForManyBookmarks(ctx, tx, bookmarks)
	if err != nil {
		return nil, 0, errors.Join(err, tx.Rollback())
	}
	return bookmarks, total, tx.Commit()
}

// BookmarksForDay returns the bookmarks created on the given dayStamp, which
// looks like 2023-03-14. The result might as well be nil, that means
// there are no bookmarks for the day.
func (repo *RepoLocalBookmarks) BookmarksForDay(
	ctx context.Context,
	authorized bool,
	dayStamp string,
) ([]types.Bookmark, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	rows, err := tx.QueryContext(ctx, `
select
	ID, URL, Title, Description, Visibility, CreationTime, RepostOf, OriginalAuthorID
from
	Bookmarks
where
	DeletionTime is null and (Visibility = 1 or ?) and CreationTime like ?
order by
	CreationTime desc;
`, authorized, dayStamp+"%")
	if err != nil {
		return nil, errors.Join(err, tx.Rollback())
	}
	bookmarks, err := scanBookmarks(rows)
	if err != nil {
		return nil, errors.Join(err, tx.Rollback())
	}

	bookmarks, err = repo.tagsForManyBookmarks(ctx, tx, bookmarks)
	if err != nil {
		return nil, errors.Join(err, tx.Rollback())
	}
	return bookmarks, tx.Commit()
}

func (repo *RepoLocalBookmarks) BookmarksWithTag(
	ctx context.Context,
	authorized bool,
	tagName string,
	page uint,
) (bookmarks []types.Bookmark, total uint, err error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, 0, err
	}

	if err = tx.QueryRowContext(ctx, `
select
	count(ID)
from
	Bookmarks
inner join
	TagsToPosts
where
	ID = PostID and TagName = ? and DeletionTime is null and (Visibility = 1 or ?)
`, tagName, authorized).Scan(&total); err != nil {
		return nil, 0, errors.Join(err, tx.Rollback())
	}

	rows, err := tx.QueryContext(ctx, `
select
	ID, URL, Title, Description, Visibility, CreationTime, RepostOf, OriginalAuthorID
from
	Bookmarks
inner join
	TagsToPosts
where
	ID = PostID and TagName = ? and DeletionTime is null and (Visibility = 1 or ?)
order by
	CreationTime desc
limit ? offset ?;
`, tagName, authorized, types.BookmarksPerPage, types.BookmarksPerPage*(page-1))
	if err != nil {
		return nil, 0, errors.Join(err, tx.Rollback())
	}
	bookmarks, err = scanBookmarks(rows)
	if err != nil {
		return nil, 0, errors.Join(err, tx.Rollback())
	}

	bookmarks, err = repo.tagsForManyBookmarks(ctx, tx, bookmarks)
	if err != nil {
		return nil, 0, errors.Join(err, tx.Rollback())
	}
	return bookmarks, total, tx.Commit()
}

func (repo *RepoLocalBookmarks) EditBookmark(
	ctx context.Context,
	bm types.Bookmark,
) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
update Bookmarks
set
    URL = ?,
    Title = ?,
    Description = ?,
    Visibility = ?,
	RepostOf = ?,
    OriginalAuthorID = ?
where
    ID = ? and DeletionTime is null;
`, bm.URL, bm.Title, bm.Description, bm.Visibility, bm.RepostOf, bm.OriginalAuthor, bm.ID)
	if err != nil {
		return errors.Join(err, tx.Rollback())
	}

	if _, err = tx.ExecContext(ctx, `delete from TagsToPosts where PostID = ?;`, bm.ID); err != nil {
		return errors.Join(err, tx.Rollback())
	}
	for _, tag := range bm.Tags {
		if tag.Name == "" {
			continue
		}
		_, err = tx.ExecContext(ctx, `insert into TagsToPosts (TagName, PostID) values (?, ?);`, tag.Name, bm.ID)
		if err != nil {
			return errors.Join(err, tx.Rollback())
		}
	}
	return tx.Commit()
}

func (repo *RepoLocalBookmarks) RandomBookmarks(
	ctx context.Context,
	authorized bool,
	n uint,
) (bookmarks []types.Bookmark, total uint, err error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, 0, err
	}

	rows, err := tx.QueryContext(ctx, `
select * from
(
	select ID, URL, Title, Description, Visibility, CreationTime, RepostOf, OriginalAuthorID
	from Bookmarks
	where DeletionTime is null and (Visibility = 1 or ?)
	order by random() limit ?
)
order by CreationTime desc;`, authorized, n)
	if err != nil {
		return nil, 0, errors.Join(err, tx.Rollback())
	}
	bookmarks, err = scanBookmarks(rows)
	if err != nil {
		return nil, 0, errors.Join(err, tx.Rollback())
	}

	bookmarks, err = repo.tagsForManyBookmarks(ctx, tx, bookmarks)
	if err != nil {
		return nil, 0, errors.Join(err, tx.Rollback())
	}
	return bookmarks, uint(len(bookmarks)), tx.Commit()
}

func (repo *RepoLocalBookmarks) DeleteBookmark(
	ctx context.Context,
	id int,
) error {
	_, err := db.ExecContext(ctx, `update Bookmarks set DeletionTime = current_timestamp where ID = ?`, id)
	return err
}

func (repo *RepoLocalBookmarks) BookmarkCount(
	ctx context.Context,
	authorized bool,
) (uint, error) {
	row := db.QueryRowContext(ctx, `
with
	IgnoredBookmarks as (
		-- Ignore deleted bookmarks always
		select ID from Bookmarks where DeletionTime is not null
		union
		-- Ignore private bookmarks if so desired
		select ID from Bookmarks where Visibility = 0 and not ?
	)
select
	count(ID)
from
	Bookmarks
where
	ID not in IgnoredBookmarks;
`, authorized)
	var count uint
	err := row.Scan(&count)
	return count, err
}

func (repo *RepoLocalBookmarks) tagsForBookmarkByID(
	ctx context.Context,
	q querier,
	id int,
) ([]types.Tag, error) {
	rows, err := q.QueryContext(ctx, `
select distinct TagName from TagsToPosts where PostID = ? order by TagName;
`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []types.Tag
	for rows.Next() {
		var tag types.Tag
		if err = rows.Scan(&tag.Name); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	return tags, rows.Err()
}

func (repo *RepoLocalBookmarks) tagsForManyBookmarks(
	ctx context.Context,
	q querier,
	bookmarks []types.Bookmark,
) ([]types.Bookmark, error) {
	for i, bm := range bookmarks {
		tags, err := repo.tagsForBookmarkByID(ctx, q, bm.ID)
		if err != nil {
			return nil, err
		}
		bookmarks[i].Tags = tags
	}
	return bookmarks, nil
}
