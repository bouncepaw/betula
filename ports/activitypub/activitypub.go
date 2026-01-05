// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package apports

type (
	ActivityPub interface {
		KnowsRemoteBookmark(remoteBookmarkID string) (bool, error)
		AuthorOfRemoteBookmark(remoteBookmarkID string) (Actor, error)
		LocalBookmarkIDFromActivityPubID(id string) (int, error)
	}

	Actor interface {
		ID() string
		SendSerializedActivity(activity []byte) error
	}

	RemoteBookmarkRepository interface {
		Exists(id string) (bool, error)
		GetActorIDFor(bookmarkID string) (string, error)
	}
)
