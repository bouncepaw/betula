// SPDX-FileCopyrightText: 2025 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package likingsvc

import (
	"database/sql"
	"git.sr.ht/~bouncepaw/betula/fediverse/activities"
	likingports "git.sr.ht/~bouncepaw/betula/ports/liking"
	"git.sr.ht/~bouncepaw/betula/settings"
	"git.sr.ht/~bouncepaw/betula/stricks"
	"log/slog"
	"strconv"
	"strings"
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

func (svc *Service) HandleIncomingLikeActivity(report activities.LikeReport) error {
	if !strings.HasPrefix(report.ObjectID, settings.SiteURL()) {
		svc.logger.Info("Received Like{} of object that's not ours; ignoring",
			"activity", report.Activity)
		return nil
	}

	stringBookmarkID := strings.TrimPrefix(report.ObjectID, settings.SiteURL()+"/")
	bookmarkID, err := strconv.Atoi(stringBookmarkID)
	if err != nil {
		return err
	}

	exists, err := svc.localBookmarkRepo.Exists(bookmarkID)
	if err != nil {
		return err
	}
	if !exists {
		svc.logger.Info("Received Like{Note} of bookmark that does not exist; ignoring",
			"activity", report.Activity, "bookmarkID", bookmarkID)
		return nil
	}

	svc.logger.Info("Received Like{Note} of our bookmark",
		"actorID", report.ActorID, "bookmarkID", bookmarkID)

	return svc.likeRepo.InsertLike(likingports.LikeModel{
		ID: sql.NullString{
			String: report.ID,
			Valid:  true,
		},
		ActorID: sql.NullString{
			String: report.ActorID,
			Valid:  true,
		},
		ObjectID:   stringBookmarkID,
		SourceJSON: report.Activity,
	})
}

func (svc *Service) HandleIncomingUpdateNoteActivity(report activities.UpdateNoteReport) error {
	if report.LikesCollection == nil {
		return nil
	}

	svc.logger.Info("Update{Note} contained a likes collection; updating records",
		"collection", report.LikesCollection,
		"likedObjectID", report.Bookmark.ID,
		"totalItems", report.LikesCollection.TotalItems)
	return svc.likeRepo.UpsertLikeCollection(likingports.LikeCollectionModel{
		ID:            stricks.PointerToSQLNullString(report.LikesCollection.ID),
		LikedObjectID: report.Bookmark.ID,
		TotalItems:    report.LikesCollection.TotalItems,
		SourceJSON:    report.Bookmark.Activity,
	})
}

func (svc *Service) LikeLocalBookmark(bookmarkID int) error {
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
	err := svc.likeRepo.InsertLike(likingports.LikeModel{
		ObjectID: strconv.Itoa(bookmarkID),
	})
	if err != nil {
		return err
	}

	// TODO: broadcast activities

	return nil
}

func (svc *Service) LikeRemoteBookmark(bookmarkID string) error {
	err := svc.likeRepo.InsertLike(likingports.LikeModel{
		ObjectID: bookmarkID,
	})
	if err != nil {
		return err
	}

	// TODO: broadcast activities

	return nil
}

func (svc *Service) CountLikesForLocalBookmarks(bookmarkIDs []int) ([]int, error) {
	//TODO implement me
	panic("implement me")
}

func (svc *Service) CountLikesForRemoteBookmarks(bookmarkIDs []string) ([]int, error) {
	//TODO implement me
	panic("implement me")
}
