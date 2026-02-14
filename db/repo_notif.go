// SPDX-FileCopyrightText: 2025 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package db

import (
	"context"
	"encoding/json"
	"git.sr.ht/~bouncepaw/betula/ports/notif"
	notiftypes "git.sr.ht/~bouncepaw/betula/types/notif"
	"time"
)

type RepoNotif struct {
}

func (repo *RepoNotif) Count(ctx context.Context) (int64, error) {
	var count int64
	err := db.QueryRowContext(ctx, "select count(*) from Notifications").Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (repo *RepoNotif) Store(ctx context.Context, kind notiftypes.Kind, payload any) error {
	j, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx,
		"insert into Notifications (Kind, Payload) values (?, ?)",
		kind, j)

	return err
}

func (repo *RepoNotif) GetAll(ctx context.Context) ([]notiftypes.Notification, error) {
	rows, err := db.QueryContext(ctx,
		"select ID, CreatedAt, Kind, Payload from Notifications order by CreatedAt desc")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []notiftypes.Notification
	for rows.Next() {
		var (
			notification notiftypes.Notification
			timestamp    string
		)

		err = rows.Scan(
			&notification.ID, &timestamp,
			&notification.Kind, &notification.Payload)
		if err != nil {
			return nil, err
		}

		notification.CreatedAt, err = time.Parse(time.DateTime, timestamp)
		if err != nil {
			return nil, err
		}

		notifications = append(notifications, notification)
	}
	return notifications, nil
}

func (repo *RepoNotif) DeleteAll(ctx context.Context) error {
	_, err := db.ExecContext(ctx, "delete from Notifications")
	return err
}

func (repo *RepoNotif) DeleteDate(ctx context.Context, date string) error {
	_, err := db.ExecContext(ctx, "delete from Notifications where CreatedAt like ?", date+"%")
	return err
}

var _ notifports.Repository = &RepoNotif{}

func New() *RepoNotif {
	return &RepoNotif{}
}
