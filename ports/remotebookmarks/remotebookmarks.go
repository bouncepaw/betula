// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package remotebookmarksports

import (
	"context"
	"database/sql"
	"encoding/json"
)

type (
	RemoteBookmarkRepository interface {
		Exists(id string) (bool, error)
		GetActorIDFor(bookmarkID string) (string, error)
		Delete(ctx context.Context, bookmarkID string) error
	}

	// TODO: Finish
	RemoteBookmarkModel struct {
		ID       string
		RepostOf sql.NullString
		ActorID  string

		Title string
		URL   sql.NullString

		Tags []string

		HTML       sql.NullString
		Mycomarkup sql.NullString

		PublishedAt string
		UpdatedAt   sql.NullString
		Activity    json.RawMessage
	}
)
