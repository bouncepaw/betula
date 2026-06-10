// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package sessionsports

import (
	"context"

	"git.sr.ht/~bouncepaw/betula/types"
)

type Repository interface {
	AddSession(ctx context.Context, token, userAgent string) error
	SessionExists(ctx context.Context, token string) (bool, error)
	StopSession(ctx context.Context, token string) error
	StopAllSessions(ctx context.Context, excludeToken string) error
	Sessions(ctx context.Context) ([]types.Session, error)
}
