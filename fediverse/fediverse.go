// SPDX-FileCopyrightText: 2022-2025 Betula contributors
//
// SPDX-License-Identifier: AGPL-3.0-only

// Package fediverse has some of the Fediverse-related functions.
package fediverse

import (
	"bytes"
	"git.sr.ht/~bouncepaw/betula/fediverse/signing"
	"io"
	"log"
	"log/slog"
	"net/http"
	"time"

	"git.sr.ht/~bouncepaw/betula/db"
	"git.sr.ht/~bouncepaw/betula/myco"
	"git.sr.ht/~bouncepaw/betula/settings"
	"git.sr.ht/~bouncepaw/betula/types"
)

var client = http.Client{
	Timeout: 2 * time.Second,
}

func PostSignedDocumentToAddress(doc []byte, contentType string, accept string, addr string) ([]byte, int, error) {
	rq, err := http.NewRequest(http.MethodPost, addr, bytes.NewReader(doc))
	if err != nil {
		slog.Error("Failed to prepare signed document request",
			"err", err, "addr", addr)
		return nil, 0, err
	}

	rq.Header.Set("User-Agent", settings.UserAgent())
	rq.Header.Set("Content-Type", contentType)
	rq.Header.Set("Accept", accept)
	signing.SignRequest(rq, doc)

	resp, err := client.Do(rq)
	if err != nil {
		slog.Error("Failed to send signed document request",
			"err", err, "addr", addr)
		return nil, 0, err
	}

	var (
		bodyReader = io.LimitReader(resp.Body, 1024*1024*10)
		body       []byte
	)
	body, err = io.ReadAll(bodyReader)
	if err != nil {
		slog.Error("Failed to read body", "err", err)
		return nil, 0, err
	}

	if resp.StatusCode != http.StatusOK {
		slog.Warn("Non-OK status code returned",
			"err", err, "addr", addr, "status", resp.StatusCode)
	}

	return body, resp.StatusCode, nil
}

func OurID() string {
	return settings.SiteURL() + "/@" + settings.AdminUsername()
}

func RenderRemoteBookmarks(raws []types.RemoteBookmark) (renders []types.RenderedRemoteBookmark) {
	// Gather actor info to prevent duplicate fetches from db
	actors := map[string]*types.Actor{}
	for _, raw := range raws {
		actors[raw.ActorID] = nil
	}
	for actorID, _ := range actors {
		actor, _ := db.ActorByID(actorID)
		actors[actorID] = actor // might be nil? I doubt it
	}

	// Rendering
	for _, raw := range raws {
		actor, ok := actors[raw.ActorID]
		if !ok {
			log.Printf("When rendering remote bookmarks: actor %s not found\n", raw.ActorID)
			continue // whatever
		}

		render := types.RenderedRemoteBookmark{
			ID:                  raw.ID,
			AuthorAcct:          actor.Acct(),
			AuthorDisplayedName: actor.PreferredUsername,
			RepostOf:            raw.RepostOf,
			Title:               raw.Title,
			URL:                 raw.URL,
			Tags:                raw.Tags,
		}

		t, err := time.Parse(types.TimeLayout, raw.PublishedAt)
		if err != nil {
			log.Printf("When rendering remote bookmarks: %s\n", err)
			continue // whatever
		}
		render.PublishedAt = t

		if raw.DescriptionMycomarkup.Valid {
			render.Description = myco.MarkupToHTML(raw.DescriptionMycomarkup.String)
		} else {
			render.Description = raw.DescriptionHTML
		}

		renders = append(renders, render)
	}

	return renders
}
