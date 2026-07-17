// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package bxerr

import "errors"

func IgnoreAkin(err error, akin ...error) error {
	for _, e := range akin {
		if errors.Is(err, e) {
			return nil
		}
	}
	return err
}
