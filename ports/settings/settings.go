// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package settingsports

import "context"

type (
	Repository interface {
		GetLoggingSettings(context.Context) (LoggingSettings, error)
		SetLoggingSettings(context.Context, LoggingSettings) error
	}
	Service interface {
		GetLoggingSettings(context.Context) (LoggingSettings, error)
		SaveLoggingSettings(context.Context, LoggingSettings) error
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
