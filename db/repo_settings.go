// SPDX-FileCopyrightText: 2023 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2024 Danila Gorelko
// SPDX-FileCopyrightText: 2024 Timur Ismagilov <https://bouncepaw.com>
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

func (repo *SettingsRepo) SetCredentials(ctx context.Context, name, hash string) error {
	_, err := db.ExecContext(ctx, `
insert or replace into BetulaMeta (Key, Value) values
	(?, ?),
	(?, ?);
`,
		settingsports.BetulaMetaAdminUsername, name,
		settingsports.BetulaMetaAdminPasswordHash, hash,
	)
	return err
}

// metaEntry reads a single BetulaMeta value into a T. A missing key yields the
// zero value and a nil error, matching the behaviour the typed reader methods
// expose through the repository interface.
func metaEntry[T any](ctx context.Context, key settingsports.BetulaMetaKey) (T, error) {
	var val T
	err := db.QueryRowContext(ctx, `select Value from BetulaMeta where Key = ? limit 1;`, key).Scan(&val)
	if errors.Is(err, sql.ErrNoRows) {
		return val, nil
	}
	return val, err
}

func setMetaEntry(ctx context.Context, key settingsports.BetulaMetaKey, val any) error {
	_, err := db.ExecContext(ctx, `insert or replace into BetulaMeta values (?, ?);`, key, val)
	return err
}

func (repo *SettingsRepo) MetaEntryNullString(ctx context.Context, key settingsports.BetulaMetaKey) (sql.NullString, error) {
	return metaEntry[sql.NullString](ctx, key)
}

func (repo *SettingsRepo) MetaEntryNullInt64(ctx context.Context, key settingsports.BetulaMetaKey) (sql.NullInt64, error) {
	return metaEntry[sql.NullInt64](ctx, key)
}

func (repo *SettingsRepo) MetaEntryString(ctx context.Context, key settingsports.BetulaMetaKey) (string, error) {
	return metaEntry[string](ctx, key)
}

func (repo *SettingsRepo) MetaEntryBytes(ctx context.Context, key settingsports.BetulaMetaKey) ([]byte, error) {
	return metaEntry[[]byte](ctx, key)
}

func (repo *SettingsRepo) SetMetaEntryString(ctx context.Context, key settingsports.BetulaMetaKey, val string) error {
	return setMetaEntry(ctx, key, val)
}

func (repo *SettingsRepo) SetMetaEntryUint(ctx context.Context, key settingsports.BetulaMetaKey, val uint) error {
	return setMetaEntry(ctx, key, val)
}

func (repo *SettingsRepo) SetMetaEntryBool(ctx context.Context, key settingsports.BetulaMetaKey, val bool) error {
	return setMetaEntry(ctx, key, val)
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
