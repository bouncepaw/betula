package notif

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

func TestGroupNotificationsByDay_sameDay(t *testing.T) {
	t.Parallel()
	now := time.Now()
	prevHour := now.Add(-time.Hour)

	notifications := []notiftypes.Notification{
		{CreatedAt: now},
		{CreatedAt: prevHour},
		{CreatedAt: prevHour},
	}

	expectedGroups := []notiftypes.NotificationGroup{
		{
			Title: now.Format(time.DateOnly),
			Notifications: []notiftypes.Notification{
				{CreatedAt: now},
				{CreatedAt: prevHour},
				{CreatedAt: prevHour},
			},
		},
	}

	actualGroups := GroupNotificationsByDay(notifications)

	if len(actualGroups) != len(expectedGroups) {
		t.Fatalf("Expected %d groups, but got %d", len(expectedGroups), len(actualGroups))
	}

	if actualGroups[0].Title != expectedGroups[0].Title {
		t.Errorf("Expected title %s, but got %s", expectedGroups[0].Title, actualGroups[0].Title)
	}

	if !reflect.DeepEqual(actualGroups[0].Notifications, expectedGroups[0].Notifications) {
		t.Errorf("Expected notifications %v, but got %v", expectedGroups[0].Notifications, actualGroups[0].Notifications)
	}
}
