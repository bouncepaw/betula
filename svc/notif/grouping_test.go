// SPDX-FileCopyrightText: 2025 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package notifsvc

import (
	notiftypes "git.sr.ht/~bouncepaw/betula/types/notif"
	"reflect"
	"testing"
	"time"
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

		if len(actualGroups) != len(expectedGroups) {
			t.Fatalf("Expected %d groups, but got %d", len(expectedGroups), len(actualGroups))
		}

		for i := range expectedGroups {
			if actualGroups[i].Title != expectedGroups[i].Title {
				t.Errorf("Group %d: Expected title %s, but got %s", i, expectedGroups[i].Title, actualGroups[i].Title)
			}

			if !reflect.DeepEqual(actualGroups[i].Notifications, expectedGroups[i].Notifications) {
				t.Errorf("Group %d: Expected notifications %v, but got %v", i, expectedGroups[i].Notifications, actualGroups[i].Notifications)
			}
		}
	})

	t.Run("Empty", func(t *testing.T) {
		t.Parallel()

		actualGroups := GroupNotificationsByDay([]notiftypes.Notification{})

		if len(actualGroups) != 0 {
			t.Errorf("Expected empty slice, but got %v", actualGroups)
		}
	})
}
