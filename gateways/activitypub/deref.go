// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package apgw

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"git.sr.ht/~bouncepaw/betula/fediverse/signing"
	"git.sr.ht/~bouncepaw/betula/pkg/stricks"
	"git.sr.ht/~bouncepaw/betula/types"
)

func (ap *ActivityPub) dereferenceActorID(actorID string) (types.Actor, error) {
	req, err := http.NewRequest("GET", actorID, nil)
	if err != nil {
		return types.Actor{}, fmt.Errorf("failed to request actor: %w", err)
	}
	req.Header.Set("Accept", types.ActivityType)
	signing.SignRequest(req, nil)

	resp, err := ap.httpClient.Do(req)
	if err != nil {
		return types.Actor{}, fmt.Errorf("failed to request actor: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return types.Actor{}, fmt.Errorf("failed to request actor: status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return types.Actor{}, fmt.Errorf("failed to request actor: %w", err)
	}

	var a types.Actor
	if err = json.Unmarshal(data, &a); err != nil {
		return types.Actor{}, fmt.Errorf("failed to request actor: %w", err)
	}

	a.Domain = stricks.ParseValidURL(actorID).Host
	if !a.Valid() {
		ap.logger.Error("Invalid actor dereferenced",
			"actorID", actorID, "actor", a)
		return types.Actor{}, fmt.Errorf("actor %s invalid", a.ID)
	}
	if a.DisplayedName == "" {
		a.DisplayedName = a.PreferredUsername
	}
	return a, nil
}
