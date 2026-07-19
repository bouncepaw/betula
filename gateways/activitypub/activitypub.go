// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

// Package apgw provides the ActivityPub gateway.
package apgw

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"git.sr.ht/~bouncepaw/betula/fediverse/signing"
	"git.sr.ht/~bouncepaw/betula/pkg/bxstr"
	apports "git.sr.ht/~bouncepaw/betula/ports/activitypub"
	"git.sr.ht/~bouncepaw/betula/settings"
	"git.sr.ht/~bouncepaw/betula/svc/activitypub/parsing"
	"git.sr.ht/~bouncepaw/betula/types"
)

type ActivityPub struct {
	actorRepo          apports.ActorRepository
	remoteBookmarkRepo apports.RemoteBookmarkRepository

	logger     *slog.Logger
	httpClient *http.Client
}

var _ apports.ActivityPub = &ActivityPub{}

var noteParser = parsing.NewNoteParser()

func NewActivityPub(
	actorRepo apports.ActorRepository,
	remoteBookmarkRepo apports.RemoteBookmarkRepository,
) *ActivityPub {
	return &ActivityPub{
		actorRepo:          actorRepo,
		remoteBookmarkRepo: remoteBookmarkRepo,

		logger: slog.Default(),
		httpClient: &http.Client{
			Timeout: time.Second * 5,
		},
	}
}

func (ap *ActivityPub) BroadcastToFollowers(
	ctx context.Context,
	activity []byte,
) error {
	followers, err := ap.actorRepo.GetFollowers(ctx)
	if err != nil {
		return err
	}
	ap.logger.Info("Broadcasting to followers",
		"count", len(followers), "activity", string(activity))

	var (
		wg        sync.WaitGroup
		succSends atomic.Int32
		sema      = make(chan struct{}, 10)
	)
	for i, follower := range followers {
		wg.Add(1)
		go func() {
			sema <- struct{}{}
			defer wg.Done()
			defer func() { <-sema }()

			actor := newActor(follower)
			err := actor.sendActivityQuiet(activity)
			if err != nil {
				slog.Error("Failed to send activity to follower",
					"err", err, "i", i, "follower", follower)
				return
			}
			succSends.Add(1)
		}()
	}
	wg.Wait()

	ap.logger.Info("Broadcasted to followers",
		"success", succSends.Load(), "total", len(followers))
	return nil
}

func (ap *ActivityPub) AuthorOfRemoteBookmark(remoteBookmarkID string) (apports.Actor, error) {
	actorID, err := ap.remoteBookmarkRepo.GetActorIDFor(remoteBookmarkID)
	if err != nil {
		return nil, err
	}

	return ap.ActorByID(context.Background(), actorID, apports.GetActorsOpts{
		GetPublicKey: true,
	})
}

func (ap *ActivityPub) KnowsRemoteBookmark(remoteBookmarkID string) (bool, error) {
	if !bxstr.IsValidURL(remoteBookmarkID) {
		return false, fmt.Errorf("not url, thus not remote bookmark id: %s", remoteBookmarkID)
	}

	return ap.remoteBookmarkRepo.Exists(remoteBookmarkID)
}

func (ap *ActivityPub) LocalBookmarkIDFromActivityPubID(id string) (int, error) {
	if !strings.HasPrefix(id, settings.SiteURL()) {
		return 0, fmt.Errorf("url %s does not start with our address %s: %w",
			id, settings.SiteURL(), apports.ErrNotLocal)
	}

	parts := strings.Split(id, "/")
	lastPart := parts[len(parts)-1]
	return strconv.Atoi(lastPart)
}

func (ap *ActivityPub) ActorByID(
	ctx context.Context,
	actorID string,
	opts apports.GetActorsOpts,
) (apports.Actor, error) {
	actorDTO, err := ap.actorRepo.GetActorByID(ctx, actorID, opts)
	if err == nil {
		return newActor(actorDTO), nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	actorDTO, err = ap.dereferenceActorID(actorID)
	if err != nil {
		return nil, err
	}

	err = ap.actorRepo.StoreActor(ctx, actorDTO)
	if err != nil {
		return nil, err
	}

	return newActor(actorDTO), nil
}

func (ap *ActivityPub) RefetchAllActors(ctx context.Context) error {
	actorIDs, err := ap.actorRepo.AllActorIDs(ctx)
	if err != nil {
		return err
	}

	ap.logger.Info("Refetching all actors...", "count", len(actorIDs))

	for _, actorID := range actorIDs {
		actorDTO, derefErr := ap.dereferenceActorID(actorID)
		if derefErr != nil {
			ap.logger.Warn("Failed to dereference actor", "id", actorID, "err", derefErr)
			err = errors.Join(err, derefErr)
			continue
		}

		repoErr := ap.actorRepo.StoreActor(ctx, actorDTO)
		if repoErr != nil {
			ap.logger.Warn("Failed to store actor after dereferencing", "id", actorID, "err", repoErr)
			err = errors.Join(err, repoErr)
			continue
		}

		ap.logger.Info("Refetched actor", "actor", actorDTO)
	}

	ap.logger.Info("Done refetching", "count", len(actorIDs), "err", err)
	return err
}

func (ap *ActivityPub) DerefRemoteBookmark(ctx context.Context, id string) (types.RemoteBookmark, error) {
	req, err := http.NewRequest(http.MethodGet, id, nil)
	if err != nil {
		return types.RemoteBookmark{}, err
	}
	req.Header.Set("User-Agent", settings.UserAgent())
	req.Header.Set("Accept", types.OtherActivityType)
	req.Header.Add("Accept", types.ActivityType)

	signing.SignRequest(req, nil)
	resp, err := ap.httpClient.Do(req)
	if err != nil {
		return types.RemoteBookmark{}, err
	}
	defer resp.Body.Close()

	var object apports.Dict
	if err := json.NewDecoder(io.LimitReader(resp.Body, 128_000)).Decode(&object); err != nil {
		return types.RemoteBookmark{}, err
	}

	bookmark, err := noteParser.BookmarkFromNote(object)
	if err != nil {
		return types.RemoteBookmark{}, err
	}
	return *bookmark, nil
}
