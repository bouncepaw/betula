// SPDX-FileCopyrightText: 2023 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2024 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2025 Danila Gorelko
// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package db

import (
	"context"
	"database/sql"
	"html/template"
	"log/slog"
	"time"

	remarkingports "git.sr.ht/~bouncepaw/betula/ports/remarking"
	"git.sr.ht/~bouncepaw/betula/types"
)

type RemarksRepo struct {
}

var _ remarkingports.Repository = (*RemarksRepo)(nil)

func NewRemarksRepo() *RemarksRepo {
	return &RemarksRepo{}
}

func (repo *RemarksRepo) RemarksOf(ctx context.Context, bookmarkID int) ([]types.RemarkInfo, error) {
	const q = `
select KR.RepostURL, KR.ReposterName, KR.RepostedAt, T.Source, T.SourceType, T.HTML
from KnownReposts KR
left join Timeline T on T.ID = KR.RepostURL
where KR.PostID = ?`
	rows, err := db.QueryContext(ctx, q, bookmarkID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var remarks []types.RemarkInfo
	for rows.Next() {
		var (
			remark          types.RemarkInfo
			timestamp       string
			sourceType      sql.NullString
			descriptionHTML sql.NullString
		)
		if err := rows.Scan(&remark.URL, &remark.Name, &timestamp, &remark.Source, &sourceType, &descriptionHTML); err != nil {
			return nil, err
		}
		remark.SourceType = types.SourceTypeFromDB(sourceType)
		remark.DescriptionHTML = template.HTML(descriptionHTML.String)
		remark.Timestamp, err = time.Parse(types.TimeLayout, timestamp)
		if err != nil {
			slog.Error("Failed to parse remark timestamp", "bookmarkID", bookmarkID, "err", err)
		}
		remarks = append(remarks, remark)
	}
	return remarks, rows.Err()
}

func (repo *RemarksRepo) SaveRemark(ctx context.Context, bookmarkID int, remark types.RemarkInfo) error {
	const q = `
insert into KnownReposts (RepostURL, PostID, ReposterName)
values (?, ?, ?)
on conflict do nothing`
	_, err := db.ExecContext(ctx, q, remark.URL, bookmarkID, remark.Name)
	return err
}

func (repo *RemarksRepo) DeleteRemark(ctx context.Context, bookmarkID int, remarkURL string) error {
	_, err := db.ExecContext(ctx, `delete from KnownReposts where RepostURL = ? and PostID = ?`, remarkURL, bookmarkID)
	return err
}
