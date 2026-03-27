// SPDX-FileCopyrightText: 2023 Goldstein
// SPDX-FileCopyrightText: 2024 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2026 Danila Gorelko
// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package types

import (
	"strconv"
	"testing"

	"github.com/nalgeon/be"
)

func TestCleanerLinkParts(t *testing.T) {
	check := func(url string, expectedLeft string, expectedRight string) {
		left, right := CleanerLinkParts(url)
		be.Equal(t, left, expectedLeft)
		be.Equal(t, right, expectedRight)
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
				{
					CreationTime: "2026-01-01 10:00",
					Title:        "It never did. Key 2 malfunctions to this day.",
				},
			},
			[]LocalBookmarkGroup{
				{"2024-01-10", []RenderedLocalBookmark{
					{
						Bookmark: Bookmark{
							CreationTime: "2024-01-10 15:35",
							Title:        "I spilled energy drink on my MacBook keyboard.",
						},
					},
					{
						Bookmark: Bookmark{
							CreationTime: "2024-01-10 15:37",
							Title:        "Why did I even buy it? I don't drink energy drinks!",
						},
					},
				}},
				{"2024-01-11", []RenderedLocalBookmark{
					{
						Bookmark: Bookmark{
							CreationTime: "2024-01-11 10:00",
							Title:        "I ordered some compressed air.",
						},
					},
				}},
				{"2024-01-12", []RenderedLocalBookmark{
					{
						Bookmark: Bookmark{
							CreationTime: "2024-01-12 12:45",
							Title:        "I hope it will help me.",
						},
					},
				}},
				{"2026-01-01", []RenderedLocalBookmark{
					{
						Bookmark: Bookmark{
							CreationTime: "2026-01-01 10:00",
							Title:        "It never did. Key 2 malfunctions to this day.",
						},
					},
				}},
			},
		},
		{
			nil, nil,
		},
	}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i+1), func(t *testing.T) {
			be.Equal(t, GroupLocalBookmarksByDate(RenderLocalBookmarks(tt.args)), tt.wantGroupedPosts)
		})
	}
}
