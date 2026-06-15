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
		RefetchAllActors(ctx context.Context) error
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

	FollowService interface {
		Follow(ctx context.Context, nickname string) error
		Unfollow(ctx context.Context, nickname string) error
	}

	//nolint:interfacebloat // Forgive me... It's not forever...
	ActorRepository interface {
		GetActorByID(ctx context.Context, id string, opts GetActorsOpts) (types.Actor, error)
		StoreActor(ctx context.Context, actor types.Actor) error
		GetFollowers(context.Context) ([]types.Actor, error)
		GetFollowing(context.Context) ([]types.Actor, error)
		GetMutuals(context.Context) ([]types.Actor, error)
		AllActorIDs(context.Context) ([]string, error)

		// ActorByAcct returns the cached actor with the given handle, or
		// sql.ErrNoRows if there is none.
		ActorByAcct(ctx context.Context, user, host string) (types.Actor, error)
		// KeyPemByID returns the public key PEM for the given key ID, or an
		// empty string if there is none.
		KeyPemByID(ctx context.Context, keyID string) (string, error)

		AddFollower(ctx context.Context, id string) error
		RemoveFollower(ctx context.Context, id string) error
		AddPendingFollowing(ctx context.Context, id string) error
		MarkAsSurelyFollowing(ctx context.Context, id string) error
		StopFollowing(ctx context.Context, id string) error

		CountFollowing(context.Context) (uint, error)
		CountFollowers(context.Context) (uint, error)

		SubscriptionStatus(ctx context.Context, id string) (types.SubscriptionRelation, error)
	}
	GetActorsOpts struct {
		GetPublicKey bool
		// FIXME: remove this thing, always get the key.
	}
)
