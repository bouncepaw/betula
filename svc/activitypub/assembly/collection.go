// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package assembly

import "errors"

type Collection struct {
	ID         *string `json:"id"`
	Type       string  `json:"type"`
	TotalItems int     `json:"totalItems"`
	// No Items.
}

func (c Collection) Valid() error {
	// Empty ID allowed.
	switch {
	case c.Type != "Collection" && c.Type != "OrderedCollection":
		return errors.New("invalid collection type")
	case c.TotalItems < 0:
		return errors.New("sub-zero total items")
	default:
		return nil
	}
}
