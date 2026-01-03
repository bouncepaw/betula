// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package likingsvc

import (
	"errors"
	"fmt"
	likingports "git.sr.ht/~bouncepaw/betula/ports/liking"
	"git.sr.ht/~bouncepaw/betula/stricks"
	"git.sr.ht/~bouncepaw/betula/types"
	"log/slog"
	"strconv"
)

type Service struct {
	logger             *slog.Logger
	likeRepo           likingports.LikeRepository
	localBookmarkRepo  likingports.LocalBookmarkRepository
	remoteBookmarkRepo likingports.RemoteBookmarkRepository
}

var _ likingports.Service = &Service{}

func New(
	likeRepo likingports.LikeRepository,
	localBookmarkRepo likingports.LocalBookmarkRepository,
	remoteBookmarkRepo likingports.RemoteBookmarkRepository,
) *Service {
	return &Service{
		logger:             slog.Default(),
		likeRepo:           likeRepo,
		localBookmarkRepo:  localBookmarkRepo,
		remoteBookmarkRepo: remoteBookmarkRepo,
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

func (svc *Service) LikeAnyBookmark(bookmarkID string) error {
	if err := svc.validBookmarkID(bookmarkID); err != nil {
		return err
	}

	return svc.likeRepo.InsertLike(likingports.LikeModel{
		ObjectID: bookmarkID,
	})
}

func (svc *Service) validBookmarkID(bookmarkID string) error {
	errRemote := svc.validRemoteBookmarkID(bookmarkID)
	if errRemote == nil { // sic!
		return nil
	}
	errLocal := svc.validLocalBookmarkID(bookmarkID)
	if errLocal == nil { // sic!
		return nil
	}
	return errors.Join(errRemote, errLocal)
}

func (svc *Service) validRemoteBookmarkID(bookmarkID string) error {
	if !stricks.ValidURL(bookmarkID) {
		return fmt.Errorf("not url, thus not remote bookmark id: %s", bookmarkID)
	}
	exists, err := svc.remoteBookmarkRepo.Exists(bookmarkID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("remote bookmark id does not exist: %s", bookmarkID)
	}
	return nil
}

func (svc *Service) validLocalBookmarkID(bookmarkID string) error {
	localBookmarkID, err := strconv.Atoi(bookmarkID)
	if err != nil {
		return fmt.Errorf("not number, thus not local bookmark id: %s; %w",
			bookmarkID, err)
	}
	exists, err := svc.localBookmarkRepo.Exists(localBookmarkID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("local bookmark id does not exist: %s", bookmarkID)
	}
	return nil
}

func (svc *Service) UnlikeAnyBookmark(bookmarkID string) error {
	return svc.likeRepo.DeleteOurLikeOf(bookmarkID)
}

func (svc *Service) FillLikes(
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

	statusMap, err := svc.likeRepo.StatiFor(ids)
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
	}
	return nil
}
