package db

import (
	"context"
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

func (repo *RepoNotif) Store(ctx context.Context, notifications ...notiftypes.Notification) error {
	tx, err := db.BeginTx(ctx, nil)
	defer tx.Rollback()
	if err != nil {
		return err
	}

	for _, notification := range notifications {
		_, err = tx.Exec(
			"insert into Notifications (ID, CreatedAt, Kind, Payload) values (?, ?, ?, ?)",
			notification.ID, notification.CreatedAt, notification.Kind, notification.Payload)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (repo *RepoNotif) GetAll(ctx context.Context) ([]notiftypes.Notification, error) {
	rows, err := db.QueryContext(ctx, "select ID, CreatedAt, Kind, Payload from Notifications")
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

var _ notiftypes.Repository = &RepoNotif{}

func New() *RepoNotif {
	return &RepoNotif{}
}
