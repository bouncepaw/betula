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

	row := db.QueryRowContext(ctx, `
select count(ID) from Bookmarks where DeletionTime is null and (Visibility = 1 or ?);
`, authorized)
	if err = row.Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := db.QueryContext(ctx, `
select ID, URL, Title, Description, Visibility, CreationTime, RepostOf, OriginalAuthorID
from Bookmarks
where DeletionTime is null and (Visibility = 1 or ?)
order by CreationTime desc
limit ?
offset (? * (? - 1));
`, authorized, types.BookmarksPerPage, types.BookmarksPerPage, page)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var bm types.Bookmark
		if err = rows.Scan(&bm.ID, &bm.URL, &bm.Title, &bm.Description, &bm.Visibility, &bm.CreationTime, &bm.RepostOf, &bm.OriginalAuthor); err != nil {
			return nil, 0, err
		}
		bookmarks = append(bookmarks, bm)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}
	bookmarks, err = repo.tagsForManyBookmarks(ctx, bookmarks)
	return bookmarks, total, err
}

func (repo *RepoLocalBookmarks) tagsForBookmarkByID(
	ctx context.Context,
	id int,
) ([]types.Tag, error) {
	rows, err := db.QueryContext(ctx, `
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
	bookmarks []types.Bookmark,
) ([]types.Bookmark, error) {
	for i, bm := range bookmarks {
		tags, err := repo.tagsForBookmarkByID(ctx, bm.ID)
		if err != nil {
			return nil, err
		}
		bookmarks[i].Tags = tags
	}
	return bookmarks, nil
}

// TODO: port old queries to this repo
// https://codeberg.org/bouncepaw/betula/issues/138
