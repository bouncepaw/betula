// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package likingsvc

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	apports "git.sr.ht/~bouncepaw/betula/ports/activitypub"
	likingports "git.sr.ht/~bouncepaw/betula/ports/liking"
	notifports "git.sr.ht/~bouncepaw/betula/ports/notif"
	"git.sr.ht/~bouncepaw/betula/stricks"
	"git.sr.ht/~bouncepaw/betula/types"
	"log/slog"
	"strconv"
)

type Service struct {
	logger             *slog.Logger
	likeRepo           likingports.LikeRepository
	likeCollectionRepo likingports.LikeCollectionRepository
	localBookmarkRepo  likingports.LocalBookmarkRepository
	notifRepo          notifports.Repository

	activityPub apports.ActivityPub
}

var _ likingports.Service = &Service{}

func New(
	likeRepo likingports.LikeRepository,
	likeCollectionRepo likingports.LikeCollectionRepository,
	localBookmarkRepo likingports.LocalBookmarkRepository,
	notifRepo notifports.Repository,

	activityPub apports.ActivityPub,
) *Service {
	return &Service{
		logger:             slog.Default(),
		likeRepo:           likeRepo,
		likeCollectionRepo: likeCollectionRepo,
		localBookmarkRepo:  localBookmarkRepo,
		notifRepo:          notifRepo,

		activityPub: activityPub,
	}
}

// John Keats had died a very long time ago, so he doesn't exactly
// hold any copyright for this poem anymore, I guess? There's no
// “Public Domain” license in SPDX. See the Legal Team's statement:
//
// => https://github.com/spdx/old-wiki/blob/main/Pages/Legal%20Team/Decisions/Dealing%20with%20Public%20Domain%20within%20SPDX%20Files.md
//
// Marking it up as a snippet to highlight that it's “different”.
// Didn't find any comment header to mark up the author.
// SPDX-SnippetBegin
/*
	A THING OF BEAUTY
		from Endymion

	A thing of beauty is a joy for ever
	Its loveliness increases; it will never
	Pass into nothingness; but still will keep
	A bower quiet for us, and a sleep
	Full of sweet dreams, and health, and quiet breathing
	— John Keats
*/
// SPDX-SnippetEnd

func (svc *Service) FillLikes(
	ctx context.Context,
	localBookmarks []types.RenderedLocalBookmark,
	remoteBookmarks []types.RenderedRemoteBookmark,
) error {
	var ids = make([]string, len(localBookmarks)+len(remoteBookmarks))
	for i, bookmark := range localBookmarks {
		ids[i] = strconv.Itoa(bookmark.ID)
	}
	for i, bookmark := range remoteBookmarks {
		ids[len(localBookmarks)+i] = bookmark.ID
	}

	statusMap, err := svc.likeRepo.StatiFor(ctx, ids)
	if err != nil {
		return err
	}

	for i, bookmark := range localBookmarks {
		status := statusMap[strconv.Itoa(bookmark.ID)]
		localBookmarks[i].LikeCounter = status.Count
		localBookmarks[i].LikedByUs = status.LikedByUs
	}
	for i, bookmark := range remoteBookmarks {
		status := statusMap[bookmark.ID]
		remoteBookmarks[i].LikeCounter = status.Count
		remoteBookmarks[i].LikedByUs = status.LikedByUs

		totalItems, err := svc.likeCollectionRepo.GetTotalItemsFor(ctx, bookmark.ID)
		if errors.Is(err, sql.ErrNoRows) {
			continue
		}
		if err != nil {
			return err
		}
		remoteBookmarks[i].LikeCounter = totalItems
	}
	return nil
}

func (svc *Service) validBookmarkID(ctx context.Context, bookmarkID string) (bool, error) {
	errRemote := svc.validRemoteBookmarkID(bookmarkID)
	if errRemote == nil { // sic!
		return true, nil
	}
	errLocal := svc.validLocalBookmarkID(ctx, bookmarkID)
	if errLocal == nil { // sic!
		return false, nil
	}
	return false, errors.Join(errRemote, errLocal)
}

func (svc *Service) validRemoteBookmarkID(bookmarkID string) error {
	knows, err := svc.activityPub.KnowsRemoteBookmark(bookmarkID)
	if err != nil {
		return err
	}
	if !knows {
		return fmt.Errorf("remote bookmark id does not exist: %s", bookmarkID)
	}
	return nil
}

func (svc *Service) validLocalBookmarkID(ctx context.Context, bookmarkID string) error {
	localBookmarkID, err := strconv.Atoi(bookmarkID)
	if err != nil {
		return fmt.Errorf("not number, thus not local bookmark id: %s; %w",
			bookmarkID, err)
	}
	exists, err := svc.localBookmarkRepo.Exists(ctx, localBookmarkID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("local bookmark id does not exist: %s", bookmarkID)
	}
	return nil
}

func (svc *Service) ReceiveLikeCollection(
	ctx context.Context,
	event likingports.EventLikeCollectionSeen,
) error {
	model := likingports.LikeCollectionModel{
		ID:            stricks.NullStringFromPtr(event.ID),
		LikedObjectID: event.LikedObjectID,
		TotalItems:    event.TotalItems,
		SourceJSON:    event.SourceJSON,
	}

	return svc.likeCollectionRepo.UpsertLikeCollection(ctx, model)
}
