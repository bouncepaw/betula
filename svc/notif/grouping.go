// SPDX-FileCopyrightText: 2025 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package notifsvc

import (
	notiftypes "git.sr.ht/~bouncepaw/betula/types/notif"
	"slices"
	"strings"
	"time"
)

// GroupNotificationsByDay sorts notifications into groups by day.
func GroupNotificationsByDay(notifications []notiftypes.Notification) []notiftypes.NotificationGroup {
	groups := make(map[string][]notiftypes.Notification)
	for _, notification := range notifications {
		date := notification.CreatedAt.Format(time.DateOnly)
		groups[date] = append(groups[date], notification)
	}

	var result []notiftypes.NotificationGroup
	for title, notifications := range groups {
		result = append(result, notiftypes.NotificationGroup{
			Title:         title,
			Notifications: notifications,
		})
	}

	slices.SortFunc(result, func(a, b notiftypes.NotificationGroup) int {
		return strings.Compare(b.Title, a.Title)
	})

	return result
}
