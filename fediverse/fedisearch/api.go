// SPDX-FileCopyrightText: 2025 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package fedisearch

import (
	"encoding/json"
	"errors"
	"git.sr.ht/~bouncepaw/betula/db"
	"git.sr.ht/~bouncepaw/betula/fediverse"
	"git.sr.ht/~bouncepaw/betula/types"
)

var (
	ErrUnsupportedVersion = errors.New("unsupported version")
	ErrWrongTo            = errors.New("field to does not match")
	ErrNotMutual          = errors.New("not mutual")
)

func ParseAPIRequest(bytes []byte) (*Request, error) {
	var req Request
	var err = json.Unmarshal(bytes, &req)
	if err != nil {
		return nil, err
	}

	switch {
	case req.Version != "v1":
		return nil, ErrUnsupportedVersion
	case req.To != fediverse.OurID():
		return nil, ErrWrongTo
	case db.SubscriptionStatus(req.From) != types.SubscriptionMutual:
		return nil, ErrNotMutual
	}

	return &req, nil
}
