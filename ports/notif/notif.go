// SPDX-FileCopyrightText: 2022-2025 Betula contributors
//
// SPDX-License-Identifier: AGPL-3.0-only

package notifports

import (
	"context"
	"git.sr.ht/~bouncepaw/betula/types/notif"
)

type Service interface {
	// Count returns number of unread notifications.
	// The result is usually cached.
	Count() (uint, error)

	// InvalidateCache invalidates the Count cache.
	InvalidateCache()

	// GetAll returns all notifications, grouped into
	// dates.
	GetAll() ([]notiftypes.NotificationGroup, error)

	// MarkAllAsRead marks all notifications as read.
	// That means they will be deleted.
	MarkAllAsRead() error
}

type Repository interface {
	Count(context.Context) (int64, error)
	Store(context.Context, notiftypes.Kind, any) error
	GetAll(context.Context) ([]notiftypes.Notification, error)
	DeleteAll(context.Context) error
}
