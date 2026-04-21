// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	settingsports "git.sr.ht/~bouncepaw/betula/ports/settings"
)

type SettingsRepo struct {
}

var _ settingsports.Repository = (*SettingsRepo)(nil)

func (repo *SettingsRepo) GetLoggingSettings(ctx context.Context) (settingsports.LoggingSettings, error) {
	var data []byte
	err := db.QueryRowContext(ctx, `
		select json_object(
			'method',   (select Value from BetulaMeta where Key = ? limit 1),
			'url',      (select Value from BetulaMeta where Key = ? limit 1),
			'username', (select Value from BetulaMeta where Key = ? limit 1),
			'token',    (select Value from BetulaMeta where Key = ? limit 1)
		)
	`, settingsports.BetulaMetaLoggingMethod,
		settingsports.BetulaMetaLoggingURL,
		settingsports.BetulaMetaLoggingUsername,
		settingsports.BetulaMetaLoggingToken,
	).Scan(&data)
	if err != nil {
		return settingsports.LoggingSettings{}, err
	}

	var ls settingsports.LoggingSettings
	return ls, json.Unmarshal(data, &ls)
}

func betulaMetaWriteOrDelete[Str ~string, T *Str](
	ctx context.Context,
	tx *sql.Tx,
	key settingsports.BetulaMetaKey,
	val T,
) error {
	if val == nil {
		_, err := tx.ExecContext(ctx, "delete from BetulaMeta where Key = ?", key)
		return err
	}
	_, err := tx.ExecContext(ctx, "insert or replace into BetulaMeta (Key, Value) values (?, ?)", key, *val)
	return err
}

func (repo *SettingsRepo) SetLoggingSettings(ctx context.Context, settings settingsports.LoggingSettings) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	err = errors.Join(
		betulaMetaWriteOrDelete(ctx, tx, settingsports.BetulaMetaLoggingMethod, settings.Method),
		betulaMetaWriteOrDelete(ctx, tx, settingsports.BetulaMetaLoggingURL, settings.URL),
		betulaMetaWriteOrDelete(ctx, tx, settingsports.BetulaMetaLoggingUsername, settings.Username),
		betulaMetaWriteOrDelete(ctx, tx, settingsports.BetulaMetaLoggingToken, settings.Token),
	)
	if err != nil {
		return errors.Join(err, tx.Rollback())
	}
	return tx.Commit()
}
