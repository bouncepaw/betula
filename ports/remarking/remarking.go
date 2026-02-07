// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package remarkingports

import "context"

type (
	Service interface {
		ReceiveLegacyRemark(context.Context, EventLegacyRemark) error
		ReceiveLegacyUnremark(context.Context, EventLegacyUnremark) error
	}

	EventLegacyRemark struct {
		ActorID    string
		AnnounceID string // id of the repost
		ObjectID   string // object that was reposted
	}
	EventLegacyUnremark struct {
		ActorID    string
		AnnounceID string // id of the repost
		ObjectID   string // object that was reposted
	}
)
