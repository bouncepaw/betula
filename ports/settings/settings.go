// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2026 Iaroslav Angliuster <https://mysh.dev>
//
// SPDX-License-Identifier: AGPL-3.0-only

package settingsports

import (
	"context"
	"database/sql"
)

type (
	Repository interface {
		GetLoggingSettings(context.Context) (LoggingSettings, error)
		SetLoggingSettings(context.Context, LoggingSettings) error
		SetCredentials(ctx context.Context, name, hash string) error

		// MetaEntry* read a single BetulaMeta value. They return the zero value
		// (an invalid sql.Null*, an empty string, a nil slice) when the key is
		// absent. These are differently typed variants of one query because Go
		// interfaces cannot carry generic methods.
		MetaEntryNullString(ctx context.Context, key BetulaMetaKey) (sql.NullString, error)
		MetaEntryNullInt64(ctx context.Context, key BetulaMetaKey) (sql.NullInt64, error)
		MetaEntryString(ctx context.Context, key BetulaMetaKey) (string, error)
		MetaEntryBytes(ctx context.Context, key BetulaMetaKey) ([]byte, error)

		// SetMetaEntry* write a single BetulaMeta value. Differently typed
		// variants for the same reason as the readers above.
		SetMetaEntryString(ctx context.Context, key BetulaMetaKey, val string) error
		SetMetaEntryUint(ctx context.Context, key BetulaMetaKey, val uint) error
		SetMetaEntryBool(ctx context.Context, key BetulaMetaKey, val bool) error
	}
	Service interface {
		GetLoggingSettings(context.Context) (LoggingSettings, error)
		SaveLoggingSettings(context.Context, LoggingSettings) error
		ApplyLoggingSettings(context.Context) error
	}
)

type (
	LoggingSettings struct {
		Method   *LoggingMethod `json:"method"`
		URL      *string        `json:"url"`
		Username *string        `json:"username"`
		// Token is also reused for passwords.
		Token *string `json:"token"`
	}

	LoggingMethod string
)

const (
	LoggingMethodDefault      LoggingMethod = ""
	LoggingMethodECSNoAuth    LoggingMethod = "ECS + No Auth"
	LoggingMethodECSBasicAuth LoggingMethod = "ECS + Basic Auth"
	LoggingMethodECSBearer    LoggingMethod = "ECS + Bearer"
)
