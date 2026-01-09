// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package activities

import (
	"encoding/json"
	"errors"
)

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

func collectionFromDict(dict Dict) (*Collection, error) {
	// A bit ineffective innit.
	j, err := json.Marshal(dict)
	if err != nil {
		return nil, err
	}

	var collection Collection
	err = json.Unmarshal(j, &collection)
	if err != nil {
		return nil, err
	}

	if err = collection.Valid(); err != nil {
		return nil, err
	}

	return &collection, nil
}
