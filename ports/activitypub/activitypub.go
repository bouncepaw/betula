// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package apports

import (
	"context"
	"git.sr.ht/~bouncepaw/betula/types"
)

type (
	ActivityPub interface {
		KnowsRemoteBookmark(remoteBookmarkID string) (bool, error)
		AuthorOfRemoteBookmark(remoteBookmarkID string) (Actor, error)
		LocalBookmarkIDFromActivityPubID(id string) (int, error)
		ActorByID(ctx context.Context, actorID string, opts GetActorsOpts) (Actor, error)
		BroadcastToFollowers(ctx context.Context, activity []byte) error
	}

	Actor interface {
		ID() string
		Acct() string
		DisplayedName() string

		SendSerializedActivity(activity []byte) error
	}

	RemoteBookmarkRepository interface {
		Exists(id string) (bool, error)
		GetActorIDFor(bookmarkID string) (string, error)
	}

	ActorRepository interface {
		GetActorByID(ctx context.Context, id string, opts GetActorsOpts) (types.Actor, error)
		StoreActor(ctx context.Context, actor types.Actor) error
		GetFollowers(context.Context) ([]types.Actor, error)
	}
	GetActorsOpts struct {
		GetPublicKey bool
	}
)
