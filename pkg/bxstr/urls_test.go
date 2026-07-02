// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package bxstr

import (
	"testing"

	"github.com/nalgeon/be"
)

func TestMustParseURL(t *testing.T) {
	t.Parallel()
	var (
		a = `/followers#@bob@bob.bouncepaw.com`
		u = MustParseURL(a)
	)
	be.Equal(t, u.Fragment, "@bob@bob.bouncepaw.com")
}

func TestValidURLWithQuery(t *testing.T) {
	t.Parallel()
	t.Run("Unfollow method", func(t *testing.T) {
		t.Parallel()

		var (
			s   = `/followers#@bob@bob.bouncepaw.com`
			got = ValidURLWithQuery(s, map[string]string{
				"unfollow-ok": "true",
			})
			want = `/followers?unfollow-ok=true#@bob@bob.bouncepaw.com`
		)
		be.Equal(t, got, want)
	})
}
