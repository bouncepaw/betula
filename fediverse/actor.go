// SPDX-FileCopyrightText: 2022-2025 Betula contributors
//
// SPDX-License-Identifier: AGPL-3.0-only

package fediverse

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"git.sr.ht/~bouncepaw/betula/db"
	"git.sr.ht/~bouncepaw/betula/fediverse/signing"
	"git.sr.ht/~bouncepaw/betula/pkg/stricks"
	"git.sr.ht/~bouncepaw/betula/types"
)

// RequestActorByNickname returns actor by string like @bouncepaw@links.bouncepaw.com or bouncepaw@links.bouncepaw.com. The returned value might be from the cache and perhaps stale.
func RequestActorByNickname(nickname string) (*types.Actor, error) {
	user, host, ok := strings.Cut(strings.TrimPrefix(nickname, "@"), "@")
	if !ok {
		return nil, fmt.Errorf("bad username: %s", nickname)
	}

	// get cached if possible
	a, found := db.ActorByAcct(user, host)
	if found {
		return a, nil
	}

	// find id
	id, err := requestIdByWebFingerAcct(user, host)
	if err == nil && id == "" {
		return nil, fmt.Errorf("user not found 404: %s", nickname)
	}
	if err != nil {
		return nil, err
	}

	// make network request
	actor, err := dereferenceActorID(id)
	if err != nil {
		return nil, fmt.Errorf("while fetching actor %s: %w", id, err)
	}

	return actor, nil
}

// RequestActorByID fetches the actor activity on the specified address. The returned value might be from the cache and perhaps stale.
//
// Deprecated: use apports.ActivityPub
func RequestActorByID(actorID string) (*types.Actor, error) {
	// get cached if possible
	a, found := db.ActorByID(actorID)
	if found {
		return a, nil
	}

	// make network request
	actor, err := dereferenceActorID(actorID)
	if err != nil {
		return nil, fmt.Errorf("while fetching actor %s: %w", actorID, err)
	}

	return actor, nil
}

func dereferenceActorID(actorID string) (*types.Actor, error) {
	req, err := http.NewRequest("GET", actorID, nil)
	if err != nil {
		return nil, fmt.Errorf("requesting actor: %w", err)
	}
	req.Header.Set("Accept", types.ActivityType)
	signing.SignRequest(req, nil)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("requesting actor: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("requesting actor: status not 200, id est %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("requesting actor: %w", err)
	}

	var a types.Actor
	if err = json.Unmarshal(data, &a); err != nil {
		return nil, fmt.Errorf("requesting actor: %w", err)
	}

	a.Domain = stricks.ParseValidURL(actorID).Host
	if !a.Valid() {
		fmt.Println(a)
		return nil, errors.New("actor invalid")
	}
	if a.DisplayedName == "" {
		a.DisplayedName = a.PreferredUsername
	}
	db.StoreValidActor(a)
	return &a, nil
}

func RequestActorInboxByID(actorID string) string {
	actor, err := RequestActorByID(actorID)
	if err != nil {
		log.Printf("When requesting actor %s inbox: %s\n", actorID, err)
		return ""
	}
	return actor.Inbox
}
