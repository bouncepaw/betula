// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package apports

import (
	"context"
	"encoding/json"
	"fmt"

	"git.sr.ht/~bouncepaw/betula/types"
)

var (
	ErrNotLocal = fmt.Errorf("not local")
)

type (
	ActivityPub interface {
		KnowsRemoteBookmark(remoteBookmarkID string) (bool, error)
		AuthorOfRemoteBookmark(remoteBookmarkID string) (Actor, error)
		LocalBookmarkIDFromActivityPubID(id string) (int, error)
		ActorByID(ctx context.Context, actorID string, opts GetActorsOpts) (Actor, error)
		BroadcastToFollowers(ctx context.Context, activity []byte) error
		RefetchAllActors(ctx context.Context) error
		DerefRemoteBookmark(ctx context.Context, id string) (types.RemoteBookmark, error)
		Deref(ctx context.Context, id string) (Dict, error)
	}

	Actor interface {
		ID() string
		Acct() string
		PreferredUsername() string
		DisplayedName() string

		SendSerializedActivity(activity []byte) error
	}

	RemoteBookmarkRepository interface {
		Exists(id string) (bool, error)
		GetActorIDFor(bookmarkID string) (string, error)
	}

	FollowService interface {
		// Follow follows the actor with this nickname and returns its resolved
		// acct. The nickname might also be a URL.
		Follow(ctx context.Context, nickname string) (acct string, err error)
		// Unfollow unfollows the actor wtih this nickname. Must not be a URL.
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

	Dict = map[string]any
	//nolint:interfacebloat // This is probably forever.
	Assembly interface {
		NewLike(likedObjectID, recipientID string) (json.RawMessage, error)
		NewUndoLike(likedObjectID, recipientID string) (json.RawMessage, error)
		NewAccept(acceptedActivity Dict) (json.RawMessage, error)
		NewReject(rejectedActivity Dict) (json.RawMessage, error)
		NewAnnounce(originalURL, repostURL string) (json.RawMessage, error)
		NewFollowFromUs(objectID string) (json.RawMessage, error)
		NewUndoFollowFromUs(objectID string) (json.RawMessage, error)
		DeleteNote(postID int) (json.RawMessage, error)
		CreateNote(post types.Bookmark) (json.RawMessage, error)
		UpdateNote(post types.Bookmark) (json.RawMessage, error)
		UpdateNoteWithLikes(post types.Bookmark, likeCounter int) (json.RawMessage, error)
		NoteFromBookmark(bookmark types.Bookmark) (Dict, error)
	}
)
