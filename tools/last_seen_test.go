// SPDX-FileCopyrightText: 2022-2025 Betula contributors
//
// SPDX-License-Identifier: AGPL-3.0-only

package tools

import (
	"testing"
	"time"
)

func TestLastSeen(t *testing.T) {
	toTime := time.Date(2023, 3, 21, 0, 0, 0, 0, time.UTC)

	testCases := []struct {
		name     string
		fromTime time.Time
		expected string
	}{
		{
			name:     "just now",
			fromTime: toTime.Add(-time.Millisecond * 10),
			expected: "just now",
		},
		{
			name:     "a second ago",
			fromTime: toTime.Add(-1 * time.Second),
			expected: "a second ago",
		},
		{
			name:     "seconds ago",
			fromTime: toTime.Add(-10 * time.Second),
			expected: "10 seconds ago",
		},
		{
			name:     "a minute ago",
			fromTime: toTime.Add(-1 * time.Minute),
			expected: "a minute ago",
		},
		{
			name:     "minutes ago",
			fromTime: toTime.Add(-3 * time.Minute),
			expected: "3 minutes ago",
		},
		{
			name:     "an hour ago",
			fromTime: toTime.Add(-1 * time.Hour),
			expected: "an hour ago",
		},
		{
			name:     "hours ago",
			fromTime: toTime.Add(-2 * time.Hour),
			expected: "2 hours ago",
		},
		{
			name:     "a day ago",
			fromTime: toTime.Add(-24 * time.Hour),
			expected: "a day ago",
		},
		{
			name:     "days ago",
			fromTime: toTime.Add(-3 * 24 * time.Hour),
			expected: "3 days ago",
		},
		{
			name:     "a week ago",
			fromTime: toTime.Add(-7 * 24 * time.Hour),
			expected: "a week ago",
		},
		{
			name:     "weeks ago",
			fromTime: toTime.Add(-14 * 24 * time.Hour),
			expected: "on Tuesday, March 7, 2023",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := LastSeen(tc.fromTime, toTime)
			if result != tc.expected {
				t.Errorf("Expected %s, but got %s", tc.expected, result)
			}
		})
	}
}
