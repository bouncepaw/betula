// SPDX-FileCopyrightText: 2022-2025 Betula contributors
//
// SPDX-License-Identifier: AGPL-3.0-only

package types

import (
	"fmt"
	"reflect"
	"testing"
)

func TestCleanerLinkParts(t *testing.T) {
	check := func(url string, expectedLeft string, expectedRight string) {
		left, right := CleanerLinkParts(url)
		if left != expectedLeft {
			t.Errorf("Wrong left part for `%s`: expected `%s`, got `%s`", url, expectedLeft, left)
		}
		if right != expectedRight {
			t.Errorf("Wrong right part for `%s`: expected `%s`, got `%s`", url, expectedRight, right)
		}
	}

	check("gopher://foo.bar/baz", "gopher://foo.bar", "/baz")
	check("https://example.com/", "example.com", "")
	check("http://xn--d1ahgkh6g.xn--90aczn5ei/%F0%9F%96%A4", "юникод.любовь", "/🖤")
	check("http://юникод.любовь/🖤", "юникод.любовь", "/🖤")
	check("http://example.com/?query=param#a/b", "example.com", "?query=param#a/b")
	check("mailto:user@example.com", "mailto:user@example.com", "")
	check("tel:+55551234567", "tel:+55551234567", "")
}

func TestGroupPostsByDate(t *testing.T) {
	tests := []struct {
		args             []Bookmark
		wantGroupedPosts []LocalBookmarkGroup
	}{
		{
			[]Bookmark{
				{
					CreationTime: "2024-01-10 15:35",
					Title:        "I spilled energy drink on my MacBook keyboard.",
				},
				{
					CreationTime: "2024-01-10 15:37",
					Title:        "Why did I even buy it? I don't drink energy drinks!",
				},
				{
					CreationTime: "2024-01-11 10:00",
					Title:        "I ordered some compressed air.",
				},
				{
					CreationTime: "2024-01-12 12:45",
					Title:        "I hope it will help me.",
				},
			},
			[]LocalBookmarkGroup{
				{"2024-01-10", []Bookmark{
					{
						CreationTime: "2024-01-10 15:35",
						Title:        "I spilled energy drink on my MacBook keyboard.",
					},
					{
						CreationTime: "2024-01-10 15:37",
						Title:        "Why did I even buy it? I don't drink energy drinks!",
					},
				}},
				{"2024-01-11", []Bookmark{
					{
						CreationTime: "2024-01-11 10:00",
						Title:        "I ordered some compressed air.",
					},
				}},
				{"2024-01-12", []Bookmark{
					{
						CreationTime: "2024-01-12 12:45",
						Title:        "I hope it will help me.",
					},
				}},
			},
		},
		{
			nil, nil,
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i+1), func(t *testing.T) {
			if gotGroupedPosts := GroupLocalBookmarksByDate(tt.args); !reflect.DeepEqual(gotGroupedPosts, tt.wantGroupedPosts) {
				t.Errorf("GroupPostsByDate() = %v, want %v", gotGroupedPosts, tt.wantGroupedPosts)
			}
		})
	}
}
