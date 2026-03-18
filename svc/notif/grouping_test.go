// SPDX-FileCopyrightText: 2025 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2026 Danila Gorelko
//
// SPDX-License-Identifier: AGPL-3.0-only

package notifsvc

import (
	"testing"
	"time"

	notiftypes "git.sr.ht/~bouncepaw/betula/types/notif"
	"github.com/nalgeon/be"
)

func TestGroupNotificationsByDay(t *testing.T) {
	t.Parallel()

	t.Run("Happy path", func(t *testing.T) {
		now := time.Now()
		yesterday := now.Add(-24 * time.Hour)
		tomorrow := now.Add(24 * time.Hour)

		notifications := []notiftypes.Notification{
			{CreatedAt: now},
			{CreatedAt: yesterday},
			{CreatedAt: tomorrow},
			{CreatedAt: now},
		}

		expectedGroups := []notiftypes.NotificationGroup{
			{
				Title: tomorrow.Format(time.DateOnly),
				Notifications: []notiftypes.Notification{
					{CreatedAt: tomorrow},
				},
			},
			{
				Title: now.Format(time.DateOnly),
				Notifications: []notiftypes.Notification{
					{CreatedAt: now},
					{CreatedAt: now},
				},
			},
			{
				Title: yesterday.Format(time.DateOnly),
				Notifications: []notiftypes.Notification{
					{CreatedAt: yesterday},
				},
			},
		}

		actualGroups := GroupNotificationsByDay(notifications)

		be.Equal(t, len(actualGroups), len(expectedGroups))

		for i := range expectedGroups {
			be.Equal(t, actualGroups[i].Title, expectedGroups[i].Title)
			be.Equal(t, actualGroups[i].Notifications, expectedGroups[i].Notifications)
		}
	})

	t.Run("Empty", func(t *testing.T) {
		t.Parallel()

		actualGroups := GroupNotificationsByDay([]notiftypes.Notification{})

		be.Equal(t, len(actualGroups), 0)
	})
}
